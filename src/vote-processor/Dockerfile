# src/vote-processor/Dockerfile

# --- STAGE 1: Build ---
FROM golang:1.24.5-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /vote-processor .

# --- STAGE 2: Final Image ---
FROM alpine:latest
WORKDIR /
COPY --from=builder /vote-processor /vote-processor
EXPOSE 8080
ENTRYPOINT ["/vote-processor"]