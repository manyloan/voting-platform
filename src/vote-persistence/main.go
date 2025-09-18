// src/vote-persistence/main.go
package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Vote struct {
	PollId   string `json:"pollId"`
	OptionId string `json:"optionId"`
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func main() {
	// --- Conexão com o PostgreSQL com Retry ---
	dbpool, err := connectToDB()
	failOnError(err, "Não foi possível conectar ao PostgreSQL após várias tentativas")
	defer dbpool.Close()

	// Garante que a tabela 'votes' exista
	_, err = dbpool.Exec(context.Background(), `
        CREATE TABLE IF NOT EXISTS votes (
            id SERIAL PRIMARY KEY,
            poll_id VARCHAR(255) NOT NULL,
            option_id VARCHAR(255) NOT NULL,
            created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
        );
    `)
	failOnError(err, "Não foi possível criar a tabela 'votes'")
	log.Println("Tabela 'votes' verificada/criada com sucesso.")

	// --- Conexão com o RabbitMQ com Retry ---
	conn, err := connectToRabbitMQ()
	failOnError(err, "Não foi possível conectar ao RabbitMQ após várias tentativas")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Falha ao abrir um canal")
	defer ch.Close()

	// Declara a mesma exchange que o produtor usa
	err = ch.ExchangeDeclare("votes_exchange", "fanout", true, false, false, false, nil)
	failOnError(err, "Falha ao declarar a exchange")

	// Declara a fila que vai receber as mensagens
	q, err := ch.QueueDeclare(
		"votes_persistence_queue", // nome da fila
		true,                      // durável (sobrevive a reinicializações do RabbitMQ)
		false,                     // auto-delete
		false,                     // exclusiva
		false,                     // no-wait
		nil,                       // arguments
	)
	failOnError(err, "Falha ao declarar a fila")

	// Faz o "bind": liga a fila à exchange. Agora a exchange enviará mensagens para nossa fila.
	err = ch.QueueBind(q.Name, "", "votes_exchange", false, nil)
	failOnError(err, "Falha ao fazer o bind da fila na exchange")

	// Começa a consumir as mensagens da fila
	msgs, err := ch.Consume(q.Name, "", false, false, false, false, nil)
	failOnError(err, "Falha ao registrar um consumidor")

	var forever chan struct{}

	go func() {
		for d := range msgs {
			log.Printf("Recebida uma mensagem: %s", d.Body)
			var vote Vote
			err := json.Unmarshal(d.Body, &vote)
			if err != nil {
				log.Printf("Erro ao decodificar JSON: %s", err)
				d.Nack(false, false) // Descarta a mensagem com erro
				continue
			}

			// Insere o voto no banco de dados
			_, err = dbpool.Exec(context.Background(),
				"INSERT INTO votes (poll_id, option_id) VALUES ($1, $2)",
				vote.PollId, vote.OptionId)

			if err != nil {
				log.Printf("Erro ao salvar voto no banco: %s", err)
				d.Nack(false, true) // Devolve a mensagem à fila para tentar novamente
				continue
			}

			log.Printf("Voto para a enquete '%s' salvo com sucesso!", vote.PollId)
			d.Ack(false) // Confirma para o RabbitMQ que a mensagem foi processada com sucesso
		}
	}()

	log.Printf(" [*] Aguardando por mensagens. Para sair, pressione CTRL+C")
	<-forever
}

// Funções de conexão com retry
func connectToDB() (*pgxpool.Pool, error) {
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		connStr = "postgres://admin:supersecret@postgres:5432/voting_db"
	}

	var dbpool *pgxpool.Pool
	var err error
	for i := 0; i < 10; i++ {
		dbpool, err = pgxpool.New(context.Background(), connStr)
		if err == nil {
			err = dbpool.Ping(context.Background())
			if err == nil {
				log.Println("Conectado ao PostgreSQL com sucesso!")
				return dbpool, nil
			}
		}
		log.Printf("Falha ao conectar ao DB. Tentativa %d. Tentando novamente...", i+1)
		time.Sleep(5 * time.Second)
	}
	return nil, err
}

func connectToRabbitMQ() (*amqp.Connection, error) {
	var conn *amqp.Connection
	var err error
	for i := 0; i < 10; i++ {
		conn, err = amqp.Dial("amqp://guest:guest@rabbitmq:5672/")
		if err == nil {
			log.Println("Conectado ao RabbitMQ com sucesso!")
			return conn, nil
		}
		log.Printf("Falha ao conectar ao RabbitMQ. Tentativa %d. Tentando novamente...", i+1)
		time.Sleep(5 * time.Second)
	}
	return nil, err
}
