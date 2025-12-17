# Orders Service

Сервис принимает заказы из Kafka, валидирует и сохраняет их в PostgreSQL, а затем отдаёт через HTTP API с кэшем в памяти.
Репозиторий включает Docker Compose-окружение, миграции БД и минимальный стек наблюдаемости на базе OpenTelemetry.

---

## 🚀 Ключевые возможности

* Kafka consumer с подтверждением смещений и **DLQ** для проблемных сообщений
* Хранение заказов, доставок, платежей и товаров в **PostgreSQL**
* Потокобезопасный **in-memory TTL-кэш** с фоновой очисткой
* HTTP API **v1** по UUID заказа
* OpenAPI-спецификация + **Redoc** (`/docs`)
* Трейсы, метрики и логи через **OpenTelemetry → OTLP**

---

## 🧱 Архитектура

### Слои и ответственность

* **Entry point** (`cmd/main.go`)
  Инициализация приложения, запуск HTTP и Kafka worker, обработка сигналов ОС и graceful shutdown.

* **App bootstrap** (`internal/app/app.go`)
  Конфигурация, OpenTelemetry, логгер, DI-контейнер, инфраструктура, HTTP-сервер, lifecycle.

* **DI-контейнер** (`internal/app/di.go`)
  `pgx` pool, Kafka reader/writer, PostgreSQL-репозиторий, TTL-кэш, сервис заказов, worker с DLQ.

* **HTTP слой** (`internal/http/v1`)
  Chi + ogen, OpenAPI (`/openapi.yaml`), Redoc (`/docs`), `otelhttp` middleware.

* **Доменный сервис** (`internal/service/order`)
  Логика чтения из кэша / БД и обновление кэша.

* **Хранилище** (`internal/repository/order`)
  Работа с PostgreSQL через `pgx/v5`.

* **Кэш** (`internal/cache/memory`)
  In-memory TTL-кэш с фоновой очисткой.

* **kafka** (`internal/adapter/kafka`)
  Kafka consumer → валидация → запись в БД → DLQ при ошибках.

---

### Поток данных (runtime)

1. **Приём заказов**
   Kafka → validation → PostgreSQL → (ошибка → DLQ)

2. **Чтение заказов**
   HTTP → cache → PostgreSQL → cache → response

---

### Карта пакетов

```
api/                    # OpenAPI + сгенерированный сервер
cmd/                    # Точка входа
internal/
  app/                  # DI, lifecycle, bootstrap
  cache/                # TTL in-memory cache
  http/v1/              # HTTP handlers + middleware
  adapter/kafka/         # Kafka consumer + DLQ
  service/order/        # Доменная логика
  repository/order/     # PostgreSQL
  otelx/                # OpenTelemetry SDK init
migrations/             # SQL-миграции
docker-compose.yml
Dockerfile
```

---

## 🛠 Стек технологий

* **Go** 1.25.5
* **PostgreSQL** 15
* **Kafka + Zookeeper**
* **Chi**, **ogen** (OpenAPI codegen)
* **segmentio/kafka-go**
* **pgx/v5**
* **OpenTelemetry SDK + Zap**
* **Docker / Docker Compose**

---

## ▶️ Запуск через Docker Compose

### 1. Создайте `.env`

```env
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres
POSTGRES_DB=wb
POSTGRES_PORT=5432

HTTP_ADDR=:8080
HTTP_PORT=8080
POSTGRES_DSN=postgres://postgres:postgres@postgres:5432/wb?sslmode=disable

KAFKA_BROKERS=kafka:9092
KAFKA_TOPIC=orders
KAFKA_GROUP_ID=orders-consumer
KAFKA_DLQ_TOPIC=orders.dlq

APP_ENV=local
OTEL_SERVICE_NAME=wb-orders
OTEL_EXPORTER_OTLP_ENDPOINT=otel-collector:4317
OTEL_EXPORTER_OTLP_INSECURE=true
OTEL_RESOURCE_ATTRIBUTES=service.name=wb-orders,service.version=local
```

### 2. Запуск окружения

```bash
docker-compose up -d
```

Будут запущены:

* PostgreSQL
* Kafka + Zookeeper
* Kafka UI
* OpenTelemetry Collector
* Jaeger
* Elasticsearch + Kibana
* Приложение

### 3. Логи

```bash
docker-compose logs -f app
```

### 4. Остановка

```bash
docker-compose down
# или с удалением данных
docker-compose down -v
```

---

## 🌐 Доступные сервисы

* API: [http://localhost:8080](http://localhost:8080)
* OpenAPI: [http://localhost:8080/openapi.yaml](http://localhost:8080/openapi.yaml)
* Docs (Redoc): [http://localhost:8080/docs](http://localhost:8080/docs)
* Kafka UI: [http://localhost:8081](http://localhost:8081)
* Jaeger: [http://localhost:16686](http://localhost:16686)
* Kibana: [http://localhost:5601](http://localhost:5601) (`otel-*`)
* PostgreSQL: localhost:5432
* Kafka broker: localhost:9092

> ⚠️ Prometheus ожидает `prometheus.yml`. Если он не нужен — уберите сервис из `docker-compose.yml`.

---

## 🧑‍💻 Локальная разработка (без контейнеров)

1. Установите:

    * Go 1.25.5+
    * PostgreSQL 15+
    * Kafka + Zookeeper

2. Зависимости:

   ```bash
   go mod download
   ```

3. Миграции:

   ```bash
   createdb wb
   psql -d wb -f migrations/000001_init_schema.up.sql
   ```

4. Инфраструктура (опционально через Docker):

   ```bash
   docker-compose up -d postgres zookeeper kafka kafka-ui otel-collector jaeger elasticsearch kibana
   ```

5. Переменные окружения:

   ```env
   HTTP_ADDR=:8080
   POSTGRES_DSN=postgres://user:password@localhost:5432/wb?sslmode=disable
   KAFKA_BROKERS=localhost:9092
   KAFKA_TOPIC=orders
   KAFKA_DLQ_TOPIC=orders.dlq
   KAFKA_GROUP_ID=orders-consumer
   CACHE_TTL=5m
   APP_ENV=local
   LOG_LEVEL=info
   LOG_JSON=false
   OTEL_EXPORTER_OTLP_ENDPOINT=localhost:4317
   OTEL_EXPORTER_OTLP_INSECURE=true
   OTEL_RESOURCE_ATTRIBUTES=service.name=wb-orders,service.version=local
   ```

6. Запуск:

   ```bash
   go run ./cmd
   ```

---

## ⚙️ Конфигурация

| Переменная                    | Назначение     | По умолчанию                                                 |
| ----------------------------- | -------------- | ------------------------------------------------------------ |
| `HTTP_ADDR`                   | HTTP адрес     | `:8080`                                                      |
| `POSTGRES_DSN`                | PostgreSQL DSN | `postgres://user:password@localhost:5432/wb?sslmode=disable` |
| `KAFKA_BROKERS`               | Kafka brokers  | `localhost:9092`                                             |
| `KAFKA_TOPIC`                 | Kafka topic    | `orders`                                                     |
| `KAFKA_DLQ_TOPIC`             | DLQ topic      | `orders.dlq`                                                 |
| `KAFKA_GROUP_ID`              | Consumer group | `orders-consumer`                                            |
| `CACHE_TTL`                   | TTL кэша       | `5m`                                                         |
| `APP_ENV`                     | Окружение      | `local`                                                      |
| `LOG_LEVEL`                   | Уровень логов  | `info`                                                       |
| `LOG_JSON`                    | JSON-логи      | `false`                                                      |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | OTLP endpoint  | —                                                            |
| `OTEL_EXPORTER_OTLP_INSECURE` | Insecure OTLP  | `true`                                                       |

---

## 📡 API v1

### Получить заказ по UUID

```http
GET /order/{orderUID}
```

```bash
curl http://localhost:8080/order/b563feb7b2b84b6test
```

Формат ответа описан в `api/openapi.yaml`.

---

## 🧪 Тестирование и качество

```bash
go test ./...
go fmt ./...
golangci-lint run
```

---

## 🐳 Сборка Docker-образа

```bash
docker build -t orders-service .
```
