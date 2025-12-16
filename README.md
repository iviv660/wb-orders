# Orders Service

Сервис принимает заказы из Kafka, валидирует и сохраняет их в PostgreSQL, а затем отдаёт через HTTP API с кэшем в памяти. Репозиторий включает Docker Compose-окружение, миграции БД и минимальный стек наблюдаемости через OpenTelemetry Collector.

## Ключевые возможности
- Консьюмер Kafka с подтверждением смещений и DLQ для проблемных сообщений.
- Сохранение заказов, доставок, платежей и товаров в PostgreSQL.
- Потокобезопасный in-memory кэш с TTL и фоновым GC для ускорения чтения.
- HTTP API v1 по UUID заказа с OpenAPI-описанием и Redoc-документацией.
- Экспонирование трасс, метрик и логов через OpenTelemetry SDK → OTLP.

## Архитектура

### Слои и ответственность
- **Входная точка** (`cmd/main.go`): инициализирует приложение, поднимает HTTP и Kafka worker, обрабатывает завершение по сигналам, инициирует graceful shutdown через общий closer.
- **App bootstrap** (`internal/app/app.go`): последовательно настраивает конфигурацию, OpenTelemetry, логгер, DI-контейнер, инфраструктуру, HTTP listener и сервер; регистрирует закрытие ресурсов.
- **DI-контейнер** (`internal/app/di.go`): собирает пул `pgx`, Kafka reader/writer, репозиторий PostgreSQL, кэш заказов, доменный сервис и воркер с DLQ.
- **HTTP слой** (`internal/http/v1`): сгенерированный chi/ogen-сервер, отдача `openapi.yaml`, Redoc по `/docs`, метрики/трейсы через `otelhttp` middleware.
- **Доменный сервис** (`internal/service/order`): решает, читать ли заказ из кэша или БД, обновляет кэш и возвращает агрегированные данные.
- **Хранилище** (`internal/repository/order`): чтение/запись заказов через `pgx/v5`; схема описана в `migrations/000001_init_schema.up.sql`.
- **Кэш** (`internal/cache/memory`): in-memory TTL-кэш с фоновой очисткой и потокобезопасными операциями.
- **Worker** (`internal/worker/order`): читает сообщения Kafka, валидирует payload, пишет в БД, при ошибке шлёт в DLQ и логирует событие.

### Поток данных (runtime)
1. **Приём заказов**: Kafka consumer получает сообщение → валидирует → пишет заказ в PostgreSQL через репозиторий → при ошибке пересылает сообщение в DLQ.
2. **Выдача заказов**: HTTP-хэндлер принимает UUID → сервис проверяет кэш → при отсутствии читает из БД → сохраняет в кэш на TTL → возвращает клиенту.

### Карта пакетов
- `api/` — OpenAPI-описание и сгенерированный chi/ogen сервер.
- `cmd/` — точка входа приложения и wiring.
- `internal/app/` — DI, конфигурация, телеметрия, lifecycle.
- `internal/cache/` — реализация TTL-кэша с GC.
- `internal/http/v1/` — HTTP-ручки, middleware, рендер ошибок.
- `internal/worker/order/` — консьюмер Kafka + DLQ.
- `internal/service/order/` — доменная логика заказа/кэша.
- `internal/repository/order/` — доступ к PostgreSQL.
- `internal/otelx/` — инициализация OpenTelemetry SDK (трейсы, метрики, логи).

## Стек технологий
- Go 1.25.5
- PostgreSQL 15
- Kafka + Zookeeper
- Chi, ogen (OpenAPI server generation)
- segmentio/kafka-go
- pgx/v5
- OpenTelemetry SDK + Zap
- Docker, Docker Compose

## Запуск через Docker Compose
1. Создайте `.env` рядом с `docker-compose.yml` (значения по умолчанию ниже):
   ```env
   POSTGRES_USER=postgres
   POSTGRES_PASSWORD=postgres
   POSTGRES_DB=wb
   POSTGRES_PORT=5432

   HTTP_PORT=8080
   HTTP_ADDR=:8080
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
2. Соберите окружение:
   ```bash
   docker-compose up -d
   ```
   Будут запущены PostgreSQL, Kafka+Zookeeper, Kafka UI, OpenTelemetry Collector, Jaeger, Elasticsearch, Kibana, Prometheus и приложение.
3. Логи приложения:
   ```bash
   docker-compose logs -f app
   ```
4. Остановка окружения:
   ```bash
   docker-compose down
   ```
   Добавьте `-v` для удаления данных.

После запуска:
- API: http://localhost:8080
- OpenAPI: http://localhost:8080/openapi.yaml
- Документация: http://localhost:8080/docs
- Kafka UI: http://localhost:8081
- Jaeger UI: http://localhost:16686
- Kibana: http://localhost:5601 (индексы `otel-*`)
- Prometheus: http://localhost:9090 (требует `prometheus.yml` рядом с `docker-compose.yml`)
- PostgreSQL: localhost:5432
- Kafka broker: localhost:9092

> Примечание: сервис Prometheus ожидает файл `prometheus.yml`; если он отсутствует, создайте минимальный конфиг или уберите сервис из `docker-compose.yml` перед запуском.

## Локальная разработка без контейнеров
1. Установите Go 1.25.5+, PostgreSQL 15+, Kafka и Zookeeper.
2. Загрузите зависимости:
   ```bash
   go mod download
   ```
3. Создайте базу и примените миграции:
   ```bash
   createdb wb
   psql -d wb -f migrations/000001_init_schema.up.sql
   ```
4. Поднимите инфраструктуру (опционально через Docker):
   ```bash
   docker-compose up -d postgres zookeeper kafka kafka-ui otel-collector jaeger elasticsearch kibana
   ```
5. Задайте переменные окружения (или используйте `.env`):
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
6. Запустите приложение:
   ```bash
   go run ./cmd
   ```

## Конфигурация
Ключевые переменные окружения и их значения по умолчанию:

| Переменная                    | Назначение                               | По умолчанию                                                  |
|-------------------------------|------------------------------------------|---------------------------------------------------------------|
| `HTTP_ADDR`                   | Адрес HTTP-сервера                       | `:8080`                                                       |
| `POSTGRES_DSN`                | DSN для подключения к PostgreSQL         | `postgres://user:password@localhost:5432/wb?sslmode=disable`  |
| `KAFKA_BROKERS`               | Брокеры Kafka (через запятую)            | `localhost:9092`                                              |
| `KAFKA_TOPIC`                 | Топик для чтения заказов                 | `orders`                                                      |
| `KAFKA_DLQ_TOPIC`             | Топик для записи некорректных сообщений  | `orders.dlq`                                                  |
| `KAFKA_GROUP_ID`              | Group ID консьюмера                      | `orders-consumer`                                             |
| `CACHE_TTL`                   | TTL in-memory кэша                       | `5m`                                                          |
| `APP_ENV`                     | Окружение приложения                     | `local`                                                       |
| `LOG_LEVEL`                   | Уровень логирования                      | `info`                                                        |
| `LOG_JSON`                    | Формат логов в JSON (true/false)         | `false`                                                       |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | OTLP endpoint для трейсов/метрик/логов   | — (обязателен при включённой телеметрии)                      |
| `OTEL_EXPORTER_OTLP_INSECURE` | Использовать незащищённый OTLP           | `true`                                                        |
| `OTEL_SERVICE_NAME`           | Имя сервиса в телеметрии                 | `wb-orders`                                                   |
| `OTEL_RESOURCE_ATTRIBUTES`    | Доп. атрибуты ресурса OTEL               | `service.name=wb-orders,service.version=local`                |

## API v1
- **Получить заказ по UUID**
  ```http
  GET /order/{orderUID}
  ```
- Пример:
  ```bash
  curl http://localhost:8080/order/b563feb7b2b84b6test
  ```
- Формат ответа описан в `api/openapi.yaml`; интерактивная документация доступна по `/docs`.

## Миграции и схема
Базовая схема создаёт таблицы `orders`, `deliveries`, `payments` и `items`, связанные через `order_uid`. Запустите `migrations/000001_init_schema.up.sql` перед стартом приложения.

## Тестирование и качество
- Запуск юнит-тестов: `go test ./...`
- Форматирование: `go fmt ./...`
- Линтинг (если установлен): `golangci-lint run`

## Сборка Docker-образа
```bash
docker build -t orders-service .
```

## Лицензия
MIT, если не указано иное в исходниках.
