# Orders Service

Сервис принимает заказы из Kafka, валидирует и сохраняет их в PostgreSQL, а затем отдаёт через HTTP API с кэшем в памяти. Проект готов для контейнерного запуска и включает инструменты для локальной разработки, миграций и наблюдаемости.

## Ключевые возможности
- Консьюмер Kafka с подтверждением смещений и DLQ для проблемных сообщений.
- Сохранение заказов, доставок, платежей и товаров в PostgreSQL.
- In-memory кэш с TTL и фоном очистки для быстрых повторных чтений.
- HTTP API v1 по UUID заказа с OpenAPI-описанием и Redoc-документацией.
- Экспонирование метрик и трассировок через OpenTelemetry.

## Архитектура

### Как устроено приложение
- **Слои**: HTTP-слой (chi + ogen), доменный сервис заказов, инфраструктурные адаптеры (Kafka consumer/producer, PostgreSQL), in-memory кэш.
- **Паттерны**: DI через `internal/app/di.go`, конфигурация через env, репозиторий/сервис для изоляции хранения, сигналы graceful shutdown через контекст.
- **Телеметрия**: единая инициализация OpenTelemetry в `internal/app/app.go` (трейсы и метрики), логирование через Zap.

### Основные компоненты
- **Точка входа** (`cmd/main.go`): собирает приложение, поднимает HTTP-сервер и воркер Kafka, ловит сигналы и закрывает ресурсы.
- **Инициализация** (`internal/app/app.go`): загружает конфигурацию, настраивает телеметрию и логгер, готовит HTTP-сервер, регистрирует завершалки.
- **DI-контейнер** (`internal/app/di.go`): строит пул `pgx`, Kafka reader/writer, репозиторий PostgreSQL, кэш заказов, сервис домена и воркер с DLQ.
- **HTTP слой** (`internal/http/v1`): сгенерированный сервер из `api/openapi.yaml`, хэндлер `GET /order/{orderUID}`, Redoc (`/docs`) и отдача OpenAPI-файла (`/openapi.yaml`).
- **Доменный сервис** (`internal/service/order`): решает, брать ли заказ из кэша или из БД, обновляет кэш и возвращает агрегированные данные.
- **Хранилище** (`internal/repository/order`): чтение/запись заказов через `pgx/v5`; схема описана в `migrations/000001_init_schema.up.sql`.
- **Кэш** (`internal/cache/memory`): потокобезопасный in-memory TTL-кэш с фоновым GC и ограничением по времени жизни записей.
- **Worker** (`internal/worker/order`): читает сообщения Kafka, валидирует, пишет в БД, отправляет сбойные сообщения в DLQ.

### Поток данных (run-time)
1. **Приём заказов**: Kafka consumer получает сообщение → валидирует → пишет заказ в PostgreSQL через репозиторий → при ошибке пересылает сообщение в DLQ и логирует событие.
2. **Выдача заказов**: HTTP-хэндлер принимает UUID → сервис проверяет кэш → при отсутствии читает из БД → сохраняет в кэш на TTL → возвращает клиенту.

### Схема коммуникаций
```
Kafka -> internal/worker/order -> internal/repository/order -> PostgreSQL
                     |
                     v
                internal/service/order <-> internal/cache/memory
                             |
                             v
                        internal/http/v1
```

### Как ориентироваться в коде
- `api/` — OpenAPI-описание и сгенерированный chi/ogen сервер.
- `cmd/` — точка входа приложения и wiring.
- `internal/app/` — DI, конфигурация, телеметрия, lifecycle.
- `internal/cache/` — реализация TTL-кэша с GC.
- `internal/http/v1/` — HTTP-ручки, middleware, рендер ошибок.
- `internal/worker/order/` — консьюмер Kafka + DLQ.
- `internal/service/order/` — доменная логика заказа/кэша.
- `internal/repository/order/` — доступ к PostgreSQL.

## Стек технологий
- Go 1.25
- PostgreSQL 15
- Kafka + Zookeeper
- Chi, ogen (OpenAPI server generation)
- segmentio/kafka-go
- pgx/v5
- OpenTelemetry + Zap
- Docker, Docker Compose

## Запуск через Docker Compose
1. Установите Docker и Docker Compose.
2. Соберите зависимости и поднимите инфраструктуру:
   ```bash
   docker-compose up -d
   ```
   Будут запущены PostgreSQL, Kafka+Zookeeper, Kafka UI и приложение.
3. Просмотреть логи приложения:
   ```bash
   docker-compose logs -f app
   ```
4. Остановить окружение:
   ```bash
   docker-compose down
   ```
   Для полной очистки данных добавьте флаг `-v`.

После запуска:
- API: http://localhost:8080
- OpenAPI: http://localhost:8080/openapi.yaml
- Документация: http://localhost:8080/docs
- Kafka UI: http://localhost:8081
- PostgreSQL: localhost:5432
- Kafka broker: localhost:9092

## Локальная разработка без контейнеров
1. Установите Go 1.25+, PostgreSQL 15+, Kafka и Zookeeper.
2. Загрузите модули:
   ```bash
   go mod download
   ```
3. Создайте базу и примените миграции:
   ```bash
   createdb wb
   psql -d wb -f migrations/000001_init_schema.up.sql
   ```
4. Поднимите инфраструктуру отдельно (при желании):
   ```bash
   docker-compose up -d postgres zookeeper kafka kafka-ui
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
   LOG_LEVEL=info
   LOG_JSON=false
   ```
6. Запустите приложение:
   ```bash
   go run ./cmd
   ```

## Конфигурация
Ключевые переменные окружения и их значения по умолчанию:

| Переменная         | Назначение                              | По умолчанию                                               |
|--------------------|-----------------------------------------|------------------------------------------------------------|
| `HTTP_ADDR`        | Адрес HTTP-сервера                      | `:8080`                                                    |
| `POSTGRES_DSN`     | DSN для подключения к PostgreSQL        | `postgres://user:password@localhost:5432/wb?sslmode=disable` |
| `KAFKA_BROKERS`    | Брокеры Kafka (через запятую)           | `localhost:9092`                                           |
| `KAFKA_TOPIC`      | Топик для чтения заказов                | `orders`                                                   |
| `KAFKA_DLQ_TOPIC`  | Топик для записи некорректных сообщений | `orders.dlq`                                               |
| `KAFKA_GROUP_ID`   | Group ID консьюмера                     | `orders-consumer`                                          |
| `CACHE_TTL`        | TTL in-memory кэша                      | `5m`                                                       |
| `LOG_LEVEL`        | Уровень логирования                     | `info`                                                     |
| `LOG_JSON`         | Формат логов в JSON (true/false)        | `false`                                                    |

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
Базовая схема создаёт четыре таблицы: `orders` (основные данные), `deliveries`, `payments` и `items`, связанные через `order_uid`.
Запустите миграции из `migrations/000001_init_schema.up.sql` перед стартом приложения.

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
