# Orders Service

Сервис обрабатывает заказы, поступающие из Kafka, сохраняет их в PostgreSQL и предоставляет HTTP API для чтения по UUID. Проект рассчитан на контейнеризированный запуск, но поддерживает и локальную разработку.

## Возможности
- Потребление заказов из Kafka с подтверждением смещения.
- Хранение заказов, доставок, платежей и товаров в PostgreSQL.
- In-memory кэш для ускорения повторных чтений.
- HTTP API с версионированием (`/order/{orderUID}`).
- Готовые Docker Compose сервисы для локального окружения.
- Набор миграций для первичной схемы базы.

## Технологии
- Go 1.25
- PostgreSQL 15
- Kafka + Zookeeper
- Chi (HTTP роутер)
- segmentio/kafka-go
- pgx/v5
- OpenTelemetry + Zap для логирования и метрик

## Архитектура проекта
```
cmd/
  └── main.go              # Точка входа приложения
internal/
  ├── adapter/             # Интеграции с внешними системами
  │   └── kafka/           # Kafka consumer
  ├── api/                 # HTTP слой (v1)
  ├── app/                 # Инициализация приложения и DI
  ├── config/              # Работа с конфигурацией
  ├── model/               # Доменные модели
  ├── repository/          # Доступ к данным (PostgreSQL)
  └── service/             # Бизнес-логика
migrations/                # SQL-миграции схемы
docker-compose.yml         # Инфраструктура и приложение
Dockerfile                 # Сборка образа приложения
```

## Быстрый старт (Docker Compose)
Рекомендуемый способ поднять все зависимости и приложение.

```bash
# Запустить все сервисы (PostgreSQL, Kafka, Kafka UI, приложение)
docker-compose up -d

# Смотреть логи приложения
docker-compose logs -f app

# Остановить сервисы
docker-compose down

# Полная очистка данных
docker-compose down -v
```

После запуска доступны:
- HTTP API: http://localhost:8080
- Kafka UI: http://localhost:8081
- PostgreSQL: localhost:5432
- Kafka broker: localhost:9092

## Локальная разработка (без контейнеров)
1. Установите зависимости:
   - Go 1.25+
   - PostgreSQL 15+
   - Kafka + Zookeeper

2. Скачайте модули:
   ```bash
   go mod download
   ```

3. Создайте базу данных и примените миграции:
   ```bash
   createdb wb
   psql -d wb -f migrations/000001_init_schema.up.sql
   ```

4. Запустите инфраструктуру (если хотите использовать Docker только для неё):
   ```bash
   docker-compose up -d postgres zookeeper kafka kafka-ui
   ```

5. Экспортируйте переменные окружения (или создайте `.env`):
   ```env
   HTTP_ADDR=:8080
   POSTGRES_DSN=postgres://user:password@localhost:5432/wb?sslmode=disable
   KAFKA_BROKERS=localhost:9092
   KAFKA_TOPIC=orders
   KAFKA_GROUP_ID=orders-consumer
   ```

6. Запустите приложение:
   ```bash
   go run ./cmd
   ```

## Конфигурация
Ключевые переменные окружения:

| Переменная        | Назначение                                | Значение по умолчанию                                        |
|-------------------|-------------------------------------------|--------------------------------------------------------------|
| `HTTP_ADDR`       | Адрес HTTP-сервера                        | `:8080`                                                      |
| `POSTGRES_DSN`    | Подключение к PostgreSQL                  | `postgres://user:password@localhost:5432/wb?sslmode=disable` |
| `KAFKA_BROKERS`   | Адреса брокеров Kafka (через запятую)     | `localhost:9092`                                             |
| `KAFKA_TOPIC`     | Топик, из которого читаются заказы        | `orders`                                                     |
| `KAFKA_GROUP_ID`  | Group ID Kafka consumer                   | `orders-consumer`                                            |

## API (v1)
### Получить заказ по UUID
```http
GET /order/{orderUID}
```

Пример запроса:
```bash
curl http://localhost:8080/order/b563feb7b2b84b6test
```

Пример ответа:
```json
{
  "order_id": "b563feb7b2b84b6test",
  "track_number": "WBILMTESTTRACK",
  "entry": "WBIL",
  "locale": "en",
  "internal_signature": "",
  "customer_id": "test",
  "delivery_service": "meest",
  "shard_key": "9",
  "sm_id": 99,
  "date_created": "2021-11-26T06:22:19Z",
  "off_shard": "1",
  "delivery": { "name": "Test Testov", "phone": "+9720000000", "zip": "2639809", "city": "Kiryat Mozkin", "address": "Ploshad Mira 15", "region": "Kraiot", "email": "test@gmail.com" },
  "payment": { "transaction": "b563feb7b2b84b6test", "request_id": "", "currency": "USD", "provider": "wbpay", "amount": 1817, "payment_dt": 1637907727, "bank": "alpha", "delivery_cost": 1500, "goods_total": 317, "custom_fee": 0 },
  "items": [ { "chrt_id": "9934930", "track_number": "WBILMTESTTRACK", "price": 453, "rid": "ab4219087a764ae0btest", "name": "Mascaras", "sale": 30, "size": "0", "total_price": 317, "nm_id": 2389212, "brand": "Vivienne Sabo", "status": 202 } ]
}
```

## Схема данных
Основные таблицы в `migrations/000001_init_schema.up.sql`:
- `orders` — базовая информация о заказе;
- `deliveries` — данные о доставке;
- `payments` — данные о платеже;
- `items` — товары заказа.

## Разработка
- Тесты: `go test ./...`
- Форматирование: `go fmt ./...`
- Линтинг (если установлен): `golangci-lint run`

## Сборка Docker-образа
```bash
docker build -t orders-service .
```

## Лицензия
Проект распространяется под лицензией MIT (если не указано иное).
