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

## API Документация

### Получение заказа по ID

**Endpoint:** `GET /order/{order_uid}`

**Описание:** Возвращает информацию о заказе по его уникальному идентификатору.

**Параметры:**
- `order_uid` (string, required) - Уникальный идентификатор заказа

**Ответы:**

#### 200 OK
```json
{
  "order_uid": "b563feb7b2b84b6test",
  "track_number": "WBILMTESTTRACK",
  "entry": "WBIL",
  "delivery": {
    "name": "Test Testov",
    "phone": "+9720000000",
    "zip": "2639809",
    "city": "Kiryat Mozkin",
    "address": "Ploshad Mira 15",
    "region": "Kraiot",
    "email": "test@gmail.com"
  },
  "payment": {
    "transaction": "b563feb7b2b84b6test",
    "request_id": "",
    "currency": "USD",
    "provider": "wbpay",
    "amount": 1817,
    "payment_dt": 1637907727,
    "bank": "alpha",
    "delivery_cost": 1500,
    "goods_total": 317,
    "custom_fee": 0
  },
  "items": [
    {
      "chrt_id": 9934930,
      "track_number": "WBILMTESTTRACK",
      "price": 453,
      "rid": "ab4219087a764ae0btest",
      "name": "Mascaras",
      "sale": 30,
      "size": "0",
      "total_price": 317,
      "nm_id": 2389212,
      "brand": "Vivienne Sabo",
      "status": 202
    }
  ],
  "locale": "en",
  "internal_signature": "",
  "customer_id": "test",
  "delivery_service": "meest",
  "shardkey": "9",
  "sm_id": 99,
  "date_created": "2021-11-26T06:22:19Z",
  "oof_shard": "1"
}
```

#### 404 Not Found
```json
{
  "error": "Заказ не найден"
}
```

#### 500 Internal Server Error
```json
{
  "error": "Внутренняя ошибка сервера"
}
```

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
- `POSTGRES_USER` - Пользователь PostgreSQL (по умолчанию: test_user)
- `POSTGRES_PASSWORD` - Пароль PostgreSQL (по умолчанию: test_password)
- `POSTGRES_DB` - Имя базы данных (по умолчанию: orders_db)
- `DB_HOST` - Хост базы данных (по умолчанию: localhost)
- `DB_PORT` - Порт базы данных (по умолчанию: 5432)
- `KAFKA_BROKER` - Адрес Kafka брокера (по умолчанию: localhost:9092)
- `KAFKA_TOPIC` - Топик Kafka (по умолчанию: orders)
- `HTTP_PORT` - Порт HTTP сервера (по умолчанию: 8081)

## CI/CD
- Автоматический запуск тестов и сервисов через GitHub Actions (`.github/workflows/compose.yml`)

## Тестирование

### Запуск тестов
```bash
go test ./...
```

### Запуск тестов с покрытием
```bash
go test -cover ./...
```

---

> Для проверки работы сервиса используйте веб-интерфейс или API: `GET /order/<order_uid>`

---
