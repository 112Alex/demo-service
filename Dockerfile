# Сборка приложения
FROM golang:1.24.0-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -o /demo-service cmd/service/main.go

# Запуск приложения
FROM alpine:latest
WORKDIR /root/
COPY --from=builder /demo-service .
COPY web/static ./web/static
COPY init.sql .
CMD ["./demo-service"]