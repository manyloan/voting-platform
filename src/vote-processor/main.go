// src/vote-processor/main.go
package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
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
	http.Handle("/metrics", promhttp.Handler())

	var conn *amqp.Connection
	var err error
	maxRetries := 10
	retryDelay := 5 * time.Second

	for i := 0; i < maxRetries; i++ {
		conn, err = amqp.Dial("amqp://guest:guest@rabbitmq:5672/")
		if err == nil {
			log.Println("Conectado ao RabbitMQ com sucesso!")
			break // Sai do loop se a conexão for bem-sucedida
		}

		log.Printf("Falha ao conectar ao RabbitMQ. Tentativa %d de %d. Tentando novamente em %s...", i+1, maxRetries, retryDelay)
		time.Sleep(retryDelay)
	}
	failOnError(err, "Não foi possível conectar ao RabbitMQ após várias tentativas")

	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Falha ao abrir um canal")
	defer ch.Close()

	err = ch.ExchangeDeclare(
		"votes_exchange",
		"fanout",
		true,
		false,
		false,
		false,
		nil,
	)
	failOnError(err, "Falha ao declarar a exchange")

	http.HandleFunc("/votes", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
			return
		}

		var vote Vote
		if err := json.NewDecoder(r.Body).Decode(&vote); err != nil {
			http.Error(w, "Corpo da requisição inválido", http.StatusBadRequest)
			return
		}

		body, err := json.Marshal(vote)
		if err != nil {
			http.Error(w, "Erro ao processar o voto", http.StatusInternalServerError)
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err = ch.PublishWithContext(ctx,
			"votes_exchange",
			"",
			false,
			false,
			amqp.Publishing{
				ContentType: "application/json",
				Body:        body,
			})

		if err != nil {
			http.Error(w, "Não foi possível publicar a mensagem", http.StatusInternalServerError)
			return
		}

		log.Printf("Voto publicado: %+v", vote)
		w.WriteHeader(http.StatusAccepted)
	})

	log.Println("Serviço de Votação iniciado na porta 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
