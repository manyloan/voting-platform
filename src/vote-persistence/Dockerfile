# src/vote-persistence/Dockerfile
FROM golang:1.24.5-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /vote-persistence .

FROM alpine:latest
WORKDIR /
COPY --from=builder /vote-persistence /vote-persistence
ENTRYPOINT ["/vote-persistence"]