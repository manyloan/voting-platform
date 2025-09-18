# --- STAGE 1: Build ---
# Usamos a imagem oficial do Go (com a sua versão) como base para compilar nosso código.
# O 'AS builder' dá um nome a este estágio, que usaremos mais tarde.
FROM golang:1.24.5-alpine AS builder

# Define o diretório de trabalho dentro do container.
WORKDIR /app

# Copia os arquivos de gerenciamento de dependências primeiro.
# Isso aproveita o cache do Docker. Se esses arquivos não mudarem, o Docker não baixará as dependências de novo.
COPY go.mod go.sum ./
RUN go mod download

# Copia todo o resto do código-fonte para o container.
COPY . .

# Compila nosso código Go.
# CGO_ENABLED=0 desabilita o CGO, criando um binário estático.
# GOOS=linux garante que o executável seja para Linux (o ambiente dos containers).
# O resultado da compilação será um arquivo chamado '/gateway-api'.
RUN CGO_ENABLED=0 GOOS=linux go build -o /gateway-api .


# --- STAGE 2: Final Image ---
# Começamos de uma imagem super leve, o Alpine Linux. Nossa imagem final será minúscula.
FROM alpine:latest

# Define o diretório de trabalho.
WORKDIR /

# A parte mais importante: copiamos APENAS o executável compilado do estágio 'builder'.
# Nada de código-fonte ou ferramentas de compilação na imagem final!
COPY --from=builder /gateway-api /gateway-api

# Expõe a porta 8080 para que o container possa receber tráfego nela.
EXPOSE 8080

# O comando que será executado quando o container iniciar.
# Ele simplesmente executa nosso programa compilado.
ENTRYPOINT ["/gateway-api"]