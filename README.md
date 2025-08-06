# Demo Service

Демонстрационный микросервис на Go с использованием Kafka, PostgreSQL и кеша в памяти.

## Описание
Сервис получает данные заказов из очереди Kafka, сохраняет их в базу данных PostgreSQL и кэширует в памяти для быстрого доступа. Реализован HTTP API и простой веб-интерфейс для поиска заказа по ID.

## Стек технологий
- Go
- PostgreSQL
- Kafka
- Docker, Docker Compose
- HTML, CSS, JS (web/static)

## Возможности
- Получение заказов из Kafka
- Сохранение заказов в PostgreSQL (транзакции)
- Кэширование заказов в памяти
- Восстановление кеша из БД при старте
- HTTP API: `GET /order/<order_uid>` — возвращает заказ в формате JSON
- Веб-интерфейс для поиска заказа по ID
- Обработка ошибок и устойчивость к сбоям

## Быстрый старт

### 1. Клонируйте репозиторий
```sh
git clone https://github.com/112Alex/demo-service.git
cd demo-service
```

### 2. Запуск через Docker Compose
```sh
docker compose up --build
```

- Сервис будет доступен на `http://localhost:8081`
- Веб-интерфейс: `http://localhost:8081/`

## Структура проекта
```
├── cmd/service/main.go         # Точка входа
├── internal/                   # Логика приложения
│   ├── cache/                  # Кэш
│   ├── config/                 # Конфиг
│   ├── db/                     # Работа с БД
│   ├── kafka/                  # Kafka consumer
│   ├── model/                  # Модели данных
│   ├── server/                 # HTTP сервер
├── web/static/                 # Веб-интерфейс
│   ├── index.html
│   ├── css/style.css
│   └── js/app.js
├── Dockerfile                  # Dockerfile для сервиса
├── docker-compose.yml          # Docker Compose для всех сервисов
├── init.sql                    # SQL-миграции для БД
├── .github/workflows/compose.yml # CI/CD workflow
├── .gitignore                  # Исключения для git
```

## Переменные окружения
- POSTGRES_USER
- POSTGRES_PASSWORD
- POSTGRES_DB
- DB_HOST
- DB_PORT
- KAFKA_BROKER
- KAFKA_TOPIC

## CI/CD
- Автоматический запуск тестов и сервисов через GitHub Actions (`.github/workflows/compose.yml`)

---

> Для проверки работы сервиса используйте веб-интерфейс или API: `GET /order/<order_uid>`

---
