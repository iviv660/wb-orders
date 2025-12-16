# Orders Service

Сервис для обработки и хранения заказов. Получает заказы из Kafka, сохраняет их в PostgreSQL и предоставляет HTTP API для получения заказов по UUID.

## Архитектура

Проект следует принципам чистой архитектуры:

```
cmd/
  └── main.go              # Точка входа приложения
internal/
  ├── adapter/             # Адаптеры внешних систем
  │   └── kafka/          # Kafka consumer
  ├── api/                # HTTP API слой
  │   └── v1/             # Версия API
  ├── app/                # Инициализация приложения и DI
  ├── config/             # Конфигурация
  ├── model/              # Доменные модели
  ├── repository/         # Слой работы с БД
  │   ├── converter/      # Конвертеры между слоями
  │   └── order/          # Реализация репозитория заказов
  └── service/            # Бизнес-логика
      └── order/           # Сервис заказов
```

## Функциональность

- **Kafka Consumer**: Получение заказов из Kafka топика
- **PostgreSQL**: Хранение заказов, доставок, платежей и товаров
- **HTTP API**: REST API для получения заказов по UUID
- **In-memory Cache**: Кеширование заказов для быстрого доступа

## Требования

### Для запуска через Docker Compose (рекомендуется)
- Docker 20.10+
- Docker Compose 2.0+

### Для локального запуска
- Go 1.25+
- PostgreSQL 12+
- Kafka (для consumer)

## Установка

1. Клонируйте репозиторий:
```bash
git clone https://github.com/iviv660/wb-orders
cd wb
```

### Вариант 1: Запуск через Docker Compose (рекомендуется)

Все зависимости (PostgreSQL, Kafka, Zookeeper) будут запущены автоматически:

```bash
# Запустить все сервисы
docker-compose up -d

# Просмотр логов приложения
docker-compose logs -f app

# Остановить все сервисы
docker-compose down

# Остановить и удалить volumes (очистить данные)
docker-compose down -v
```

После запуска будут доступны:
- **HTTP API**: http://localhost:8080
- **Kafka UI**: http://localhost:8081 (веб-интерфейс для управления Kafka)
- **PostgreSQL**: localhost:5432
- **Kafka**: localhost:9092

### Вариант 2: Локальный запуск

2. Установите зависимости:
```bash
go mod download
```

3. Настройте базу данных:
```bash
# Создайте базу данных
createdb wb

# Примените миграции
psql -d wb -f migrations/000001_init_schema.up.sql
```

4. Запустите Kafka и Zookeeper локально или используйте Docker Compose только для инфраструктуры:
```bash
# Запустить только инфраструктуру (без приложения)
docker-compose up -d postgres zookeeper kafka kafka-ui
```

## Конфигурация

Приложение использует переменные окружения для конфигурации:

| Переменная | Описание | Значение по умолчанию |
|------------|----------|----------------------|
| `HTTP_ADDR` | Адрес HTTP сервера | `:8080` |
| `POSTGRES_DSN` | DSN для подключения к PostgreSQL | `postgres://user:password@localhost:5432/wb?sslmode=disable` |
| `KAFKA_BROKERS` | Адреса Kafka брокеров (через запятую) | `localhost:9092` |
| `KAFKA_TOPIC` | Название Kafka топика | `orders` |
| `KAFKA_GROUP_ID` | ID группы Kafka consumer | `orders-consumer` |

Пример `.env` файла:
```env
HTTP_ADDR=:8080
POSTGRES_DSN=postgres://user:password@localhost:5432/wb?sslmode=disable
KAFKA_BROKERS=localhost:9092
KAFKA_TOPIC=orders
KAFKA_GROUP_ID=orders-consumer
```

## Запуск

### Через Docker Compose

```bash
# Запустить все сервисы (включая приложение)
docker-compose up -d

# Пересобрать и запустить приложение
docker-compose up -d --build app
```

### Локальный запуск

```bash
go run cmd/main.go
```

Или соберите бинарник:
```bash
go build -o orders-service cmd/main.go
./orders-service
```

Приложение поддерживает graceful shutdown при получении сигналов SIGINT или SIGTERM.

## API

### Получить заказ по UUID

```http
GET /order/{orderUID}
```

**Пример запроса:**
```bash
curl http://localhost:8080/order/b563feb7b2b84b6test
```

**Пример ответа:**
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
      "chrt_id": "9934930",
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
  ]
}
```

## Структура базы данных

Проект использует PostgreSQL со следующими таблицами:

- `orders` - основная информация о заказах
- `deliveries` - информация о доставке
- `payments` - информация о платежах
- `items` - товары в заказе

Подробности в файлах миграций: `migrations/000001_init_schema.up.sql`

## Разработка

### Запуск тестов
```bash
go test ./...
```

### Форматирование кода
```bash
go fmt ./...
```

### Линтинг
```bash
golangci-lint run
```

## Docker

Проект включает `Dockerfile` и `docker-compose.yml` для удобного развертывания.

### Сборка образа

```bash
docker build -t orders-service .
```

### Docker Compose сервисы

- **postgres** - PostgreSQL 15 с автоматическим применением миграций
- **zookeeper** - Zookeeper для Kafka
- **kafka** - Kafka broker с автосозданием топиков
- **kafka-ui** - Веб-интерфейс для управления Kafka (порт 8081)
- **app** - Приложение Orders Service

Все сервисы связаны через внутреннюю сеть и имеют health checks для правильной последовательности запуска.

## Технологии

- **Go 1.25+** - основной язык
- **Chi** - HTTP роутер
- **pgx/v5** - PostgreSQL драйвер
- **segmentio/kafka-go** - Kafka клиент
- **PostgreSQL** - база данных
- **Kafka** - message broker
- **Docker** - контейнеризация

