# gw-notification

Микросервис для обработки и сохранения информации о крупных денежных переводах (>= 30000 в любой валюте).

## Возможности

- Получение сообщений из Kafka
- Сохранение данных в MongoDB
- Обработка до 1000 сообщений в секунду
- Structured logging (Zap)
- Graceful shutdown
- Надежная обработка сообщений с подтверждением

## Технологии

- **Message Broker**: Apache Kafka
- **Database**: MongoDB 7
- **Logging**: Zap (structured JSON logging)
- **Consumer Group**: для масштабирования

## Структура проекта

```
gw-notification/
├── cmd/
│   └── main.go                     # Точка входа
├── internal/
│   ├── kafka/
│   │   └── consumer.go             # Kafka consumer
│   ├── repository/
│   │   └── mongo_repo.go           # MongoDB repository
│   ├── service/
│   │   ├── notification_service.go # Бизнес-логика
│   │   └── notification_service_test.go
│   ├── logging/
│   │   └── logger.go               # Zap logger setup
│   └── shutdown/
│       └── shutdown.go             # Graceful shutdown
├── config.env                      # Переменные окружения
├── Dockerfile
└── README.md
```

## Формат сообщений Kafka

### Topic: `large-transactions`

```json
{
  "user_id": "123",
  "transaction_id": "txn_abc123",
  "amount": 35000.00,
  "currency": "USD",
  "type": "deposit",
  "timestamp": "2026-03-01T20:00:00Z"
}
```

### Поля:
- `user_id` (string): ID пользователя
- `transaction_id` (string): Уникальный ID транзакции
- `amount` (float64): Сумма транзакции
- `currency` (string): Валюта (USD, RUB, EUR)
- `type` (string): Тип операции (deposit, withdraw, exchange)
- `timestamp` (string): Время транзакции (ISO 8601)

## Конфигурация

Файл `config.env`:

```env
# MongoDB
MONGO_URI=mongodb://admin:admin@localhost:27017
MONGO_DATABASE=notifications

# Kafka
KAFKA_BROKERS=localhost:9092
KAFKA_TOPIC=large-transactions
KAFKA_GROUP_ID=notification-service
```

## Локальный запуск

### 1. Установка зависимостей

```bash
go mod download
```

### 2. Запуск MongoDB

```bash
docker run -d \
  --name mongodb \
  -e MONGO_INITDB_ROOT_USERNAME=admin \
  -e MONGO_INITDB_ROOT_PASSWORD=admin \
  -p 27017:27017 \
  mongo:7
```

### 3. Запуск Kafka и Zookeeper

```bash
# Zookeeper
docker run -d \
  --name zookeeper \
  -p 2181:2181 \
  -e ZOOKEEPER_CLIENT_PORT=2181 \
  confluentinc/cp-zookeeper:7.5.0

# Kafka
docker run -d \
  --name kafka \
  -p 9092:9092 \
  -e KAFKA_BROKER_ID=1 \
  -e KAFKA_ZOOKEEPER_CONNECT=zookeeper:2181 \
  -e KAFKA_ADVERTISED_LISTENERS=PLAINTEXT://localhost:9092 \
  -e KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR=1 \
  --link zookeeper \
  confluentinc/cp-kafka:7.5.0
```

### 4. Сборка и запуск

```bash
GOOS=linux GOARCH=amd64 go build -o main ./cmd
./main
```

Или через Docker:

```bash
docker build -t gw-notification .
docker run --env-file config.env gw-notification
```

## MongoDB Схема

### Collection: `large_transactions`

```javascript
{
  "_id": ObjectId("..."),
  "user_id": "123",
  "transaction_id": "txn_abc123",
  "amount": 35000.00,
  "currency": "USD",
  "type": "deposit",
  "timestamp": ISODate("2026-03-01T20:00:00Z"),
  "processed_at": ISODate("2026-03-01T20:00:01Z")
}
```

### Индексы

```javascript
// Индекс для быстрого поиска по user_id
db.large_transactions.createIndex({ "user_id": 1 })

// Индекс для поиска по transaction_id (уникальный)
db.large_transactions.createIndex({ "transaction_id": 1 }, { unique: true })

// Индекс для поиска по timestamp
db.large_transactions.createIndex({ "timestamp": -1 })
```

## Логирование

Сервис использует structured logging (Zap) в формате JSON:

```json
{
  "level": "info",
  "ts": "2026-03-01T20:00:00.000Z",
  "caller": "service/notification_service.go:45",
  "msg": "transaction processed",
  "transaction_id": "txn_abc123",
  "user_id": "123",
  "amount": 35000.00,
  "currency": "USD",
  "type": "deposit"
}
```

### Уровни логирования:
- **DEBUG**: Детальная информация для отладки
- **INFO**: Общая информация о работе сервиса
- **WARN**: Предупреждения (например, повторная обработка сообщения)
- **ERROR**: Ошибки обработки

## Производительность

- **Throughput**: до 1000 сообщений/сек
- **Batch processing**: обработка сообщений пакетами
- **Connection pooling**: для MongoDB
- **Consumer group**: для горизонтального масштабирования

### Масштабирование

Для увеличения производительности можно запустить несколько экземпляров сервиса:

```bash
# Экземпляр 1
KAFKA_GROUP_ID=notification-service ./main

# Экземпляр 2
KAFKA_GROUP_ID=notification-service ./main

# Экземпляр 3
KAFKA_GROUP_ID=notification-service ./main
```

Kafka автоматически распределит партиции между экземплярами.

## Надежность

### Обработка ошибок

1. **Ошибка подключения к MongoDB**: Retry с exponential backoff
2. **Дубликаты сообщений**: Проверка по `transaction_id` (unique index)
3. **Невалидные сообщения**: Логирование и пропуск

### Гарантии доставки

- **At-least-once delivery**: Сообщение подтверждается только после успешного сохранения в MongoDB
- **Idempotency**: Повторная обработка одного и того же сообщения безопасна благодаря unique index

## Graceful Shutdown

Сервис корректно обрабатывает сигналы SIGINT и SIGTERM:

```go
quit := make(chan os.Signal, 1)
signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
<-quit

log.Println("Shutting down notification service...")
// Завершение обработки текущих сообщений
// Закрытие соединений с Kafka и MongoDB
```

## Тестирование

```bash
# Запуск всех тестов
go test ./...

# Запуск с покрытием
go test -cover ./...

# Запуск конкретного теста
go test -v ./internal/service -run TestNotificationService
```

### Отправка тестового сообщения в Kafka

```bash
# С помощью kafka-console-producer
docker exec -it kafka kafka-console-producer \
  --broker-list localhost:9092 \
  --topic large-transactions

# Введите JSON:
{"user_id":"123","transaction_id":"txn_test001","amount":35000.00,"currency":"USD","type":"deposit","timestamp":"2026-03-01T20:00:00Z"}
```

### Проверка данных в MongoDB

```bash
# Подключение к MongoDB
docker exec -it mongodb mongosh -u admin -p admin

# Переключение на БД
use notifications

# Просмотр документов
db.large_transactions.find().pretty()

# Подсчет документов
db.large_transactions.countDocuments()
```

## Мониторинг

### Метрики для отслеживания:

- Количество обработанных сообщений
- Количество ошибок
- Время обработки сообщения
- Размер очереди Kafka
- Задержка обработки (lag)

### Проверка lag в Kafka

```bash
docker exec -it kafka kafka-consumer-groups \
  --bootstrap-server localhost:9092 \
  --group notification-service \
  --describe
```

## Troubleshooting

### Сервис не получает сообщения

```bash
# Проверьте, что топик существует
docker exec -it kafka kafka-topics \
  --bootstrap-server localhost:9092 \
  --list

# Проверьте consumer group
docker exec -it kafka kafka-consumer-groups \
  --bootstrap-server localhost:9092 \
  --list
```

### Ошибки подключения к MongoDB

```bash
# Проверьте доступность MongoDB
docker exec -it mongodb mongosh -u admin -p admin --eval "db.adminCommand('ping')"

# Проверьте переменные окружения
echo $MONGO_URI
```

### Дубликаты в MongoDB

```bash
# Проверьте unique index
db.large_transactions.getIndexes()

# Если индекса нет, создайте его
db.large_transactions.createIndex({ "transaction_id": 1 }, { unique: true })
```

## Примеры использования

### Отправка сообщения из gw-currency-wallet

```go
import (
    "encoding/json"
    "github.com/confluentinc/confluent-kafka-go/kafka"
)

type LargeTransaction struct {
    UserID        string    `json:"user_id"`
    TransactionID string    `json:"transaction_id"`
    Amount        float64   `json:"amount"`
    Currency      string    `json:"currency"`
    Type          string    `json:"type"`
    Timestamp     time.Time `json:"timestamp"`
}

func sendToKafka(tx LargeTransaction) error {
    producer, _ := kafka.NewProducer(&kafka.ConfigMap{
        "bootstrap.servers": "localhost:9092",
    })
    defer producer.Close()
    
    data, _ := json.Marshal(tx)
    
    return producer.Produce(&kafka.Message{
        TopicPartition: kafka.TopicPartition{
            Topic:     kafka.StringPointer("large-transactions"),
            Partition: kafka.PartitionAny,
        },
        Value: data,
    }, nil)
}
```

## Лицензия

MIT
