# gw-exchanger

gRPC сервис для получения курсов валют. Предоставляет актуальные курсы обмена для USD, RUB, EUR.

## Возможности

- gRPC API для получения курсов валют
- Поддержка валют: USD, RUB, EUR
- Хранение курсов в PostgreSQL
- Structured logging (Zap)
- Graceful shutdown

## Технологии

- **Protocol**: gRPC
- **Database**: PostgreSQL
- **Logging**: Zap (structured JSON logging)
- **Protobuf**: Protocol Buffers v3

## Структура проекта

```
gw-exchanger/
├── cmd/
│   └── main.go                 # Точка входа
├── internal/
│   ├── service/
│   │   ├── service.go          # gRPC service implementation
│   │   └── service_test.go
│   ├── storages/
│   │   ├── storage.go          # Интерфейс хранилища
│   │   ├── model.go            # Модели данных
│   │   └── postgres/
│   │       ├── connector.go    # PostgreSQL connector
│   │       └── methods.go      # CRUD операции
│   └── config/
│       ├── config.go           # Конфигурация
│       └── defaults.go         # Значения по умолчанию
├── config.env                  # Переменные окружения
├── Dockerfile
└── README.md
```

## gRPC API

### GetExchangeRates

Получение курсов обмена всех валют

**Request:**
```protobuf
message Empty {}
```

**Response:**
```protobuf
message ExchangeRatesResponse {
  map<string, float> rates = 1;
}
```

**Пример:**
```json
{
  "rates": {
    "USD": 1.0,
    "RUB": 90.5,
    "EUR": 0.85
  }
}
```

### GetExchangeRateForCurrency

Получение курса обмена для конкретной пары валют

**Request:**
```protobuf
message CurrencyRequest {
  string from_currency = 1;
  string to_currency = 2;
}
```

**Response:**
```protobuf
message ExchangeRateResponse {
  string from_currency = 1;
  string to_currency = 2;
  float rate = 3;
}
```

**Пример:**
```json
{
  "from_currency": "USD",
  "to_currency": "EUR",
  "rate": 0.85
}
```

## Конфигурация

Файл `config.env`:

```env
# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=exchanger
DB_PASSWORD=exchanger
DB_NAME=exchanger

# gRPC Server
GRPC_PORT=50051
```

## Локальный запуск

### 1. Установка зависимостей

```bash
go mod download
```

### 2. Генерация protobuf кода

```bash
cd ../proto-exchange/exchange
protoc --go_out=. --go_opt=paths=source_relative \
  --go-grpc_out=. --go-grpc_opt=paths=source_relative \
  --experimental_allow_proto3_optional exchange.proto
```

### 3. Запуск PostgreSQL

```bash
docker run -d \
  --name postgres-exchanger \
  -e POSTGRES_USER=exchanger \
  -e POSTGRES_PASSWORD=exchanger \
  -e POSTGRES_DB=exchanger \
  -p 5433:5432 \
  postgres:15
```

### 4. Применение миграций

```bash
psql -h localhost -p 5433 -U exchanger -d exchanger -f migrations/001_init.sql
```

### 5. Сборка и запуск

```bash
GOOS=linux GOARCH=amd64 go build -o main ./cmd
./main -c config.env
```

Или через Docker:

```bash
docker build -t gw-exchanger .
docker run -p 50051:50051 --env-file config.env gw-exchanger
```

## Тестирование gRPC

### С помощью grpcurl

```bash
# Список сервисов
grpcurl -plaintext localhost:50051 list

# Получение всех курсов
grpcurl -plaintext localhost:50051 exchange.ExchangeService/GetExchangeRates

# Получение курса для пары валют
grpcurl -plaintext -d '{"from_currency":"USD","to_currency":"EUR"}' \
  localhost:50051 exchange.ExchangeService/GetExchangeRateForCurrency
```

### С помощью Go клиента

```go
package main

import (
    "context"
    "log"
    
    "google.golang.org/grpc"
    pb "github.com/proto-exchange/exchange_grpc"
)

func main() {
    conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()
    
    client := pb.NewExchangeServiceClient(conn)
    
    // Получение всех курсов
    resp, err := client.GetExchangeRates(context.Background(), &pb.Empty{})
    if err != nil {
        log.Fatal(err)
    }
    
    log.Printf("Rates: %+v", resp.Rates)
}
```

## Логирование

Сервис использует structured logging (Zap) в формате JSON:

```json
{
  "level": "info",
  "ts": "2026-03-01T20:00:00.000Z",
  "caller": "service/service.go:45",
  "msg": "exchange rate requested",
  "from_currency": "USD",
  "to_currency": "EUR",
  "rate": 0.85
}
```

## База данных

### Схема

```sql
CREATE TABLE exchange_rates (
    id SERIAL PRIMARY KEY,
    from_currency VARCHAR(3) NOT NULL,
    to_currency VARCHAR(3) NOT NULL,
    rate DECIMAL(10, 6) NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(from_currency, to_currency)
);

CREATE INDEX idx_exchange_rates_currencies 
ON exchange_rates(from_currency, to_currency);
```

### Начальные данные

```sql
INSERT INTO exchange_rates (from_currency, to_currency, rate) VALUES
('USD', 'USD', 1.000000),
('USD', 'RUB', 90.500000),
('USD', 'EUR', 0.850000),
('RUB', 'USD', 0.011050),
('RUB', 'RUB', 1.000000),
('RUB', 'EUR', 0.009392),
('EUR', 'USD', 1.176471),
('EUR', 'RUB', 106.470588),
('EUR', 'EUR', 1.000000);
```

## Интерфейс Storage

```go
type Storage interface {
    GetExchangeRate(ctx context.Context, from, to string) (float64, error)
    GetAllExchangeRates(ctx context.Context) (map[string]float64, error)
    UpdateExchangeRate(ctx context.Context, from, to string, rate float64) error
}
```

Это позволяет легко заменить PostgreSQL на другую БД (MongoDB, Redis и т.д.).

## Производительность

- Connection pooling для PostgreSQL
- Индексы на часто запрашиваемые поля
- Graceful shutdown для корректного завершения активных соединений

## Тестирование

```bash
# Запуск всех тестов
go test ./...

# Запуск с покрытием
go test -cover ./...

# Запуск конкретного теста
go test -v ./internal/service -run TestExchangeService
```

## Миграции

Создайте файл `migrations/001_init.sql`:

```sql
-- Создание таблицы курсов валют
CREATE TABLE exchange_rates (
    id SERIAL PRIMARY KEY,
    from_currency VARCHAR(3) NOT NULL,
    to_currency VARCHAR(3) NOT NULL,
    rate DECIMAL(10, 6) NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(from_currency, to_currency)
);

-- Индекс для быстрого поиска
CREATE INDEX idx_exchange_rates_currencies 
ON exchange_rates(from_currency, to_currency);

-- Начальные данные
INSERT INTO exchange_rates (from_currency, to_currency, rate) VALUES
('USD', 'USD', 1.000000),
('USD', 'RUB', 90.500000),
('USD', 'EUR', 0.850000),
('RUB', 'USD', 0.011050),
('RUB', 'RUB', 1.000000),
('RUB', 'EUR', 0.009392),
('EUR', 'USD', 1.176471),
('EUR', 'RUB', 106.470588),
('EUR', 'EUR', 1.000000);
```

## Troubleshooting

### Ошибка подключения к БД
```bash
# Проверьте доступность PostgreSQL
psql -h localhost -p 5433 -U exchanger -d exchanger

# Проверьте переменные окружения
echo $DB_HOST $DB_PORT
```

### gRPC сервер не запускается
```bash
# Проверьте, не занят ли порт
lsof -i :50051

# Проверьте логи
docker logs gw-exchanger
```

## Graceful Shutdown

Сервис корректно обрабатывает сигналы SIGINT и SIGTERM:

```go
quit := make(chan os.Signal, 1)
signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
<-quit

log.Println("Shutting down gRPC server...")
grpcServer.GracefulStop()
```

## Лицензия

MIT
