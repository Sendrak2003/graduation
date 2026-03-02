# gw-currency-wallet

Микросервис для управления кошельком и обмена валют с поддержкой JWT авторизации.

## Возможности

- Регистрация и авторизация пользователей (JWT)
- Управление балансом (пополнение, вывод средств)
- Получение курсов валют через gRPC
- Обмен валют с кэшированием курсов
- Уведомления о крупных транзакциях (>30000) через Kafka
- Swagger документация API

## Технологии

- **Framework**: Gin
- **Auth**: JWT
- **Database**: PostgreSQL
- **gRPC Client**: для получения курсов от gw-exchanger
- **Message Broker**: Kafka (для уведомлений о крупных транзакциях)
- **Logging**: Zap (structured JSON logging)
- **Documentation**: Swagger

## Структура проекта

```
gw-currency-wallet/
├── cmd/
│   └── main.go                 # Точка входа
├── internal/
│   ├── handler/
│   │   ├── auth_handler.go     # Регистрация/авторизация
│   │   ├── wallet_handler.go   # Операции с кошельком
│   │   └── exchange_handler.go # Обмен валют
│   ├── service/
│   │   ├── wallet_service.go   # Бизнес-логика
│   │   └── wallet_service_test.go
│   ├── repository/
│   │   └── wallet_repository.go # Работа с БД
│   ├── middleware/
│   │   ├── jwt.go              # JWT middleware
│   │   └── recovery.go         # Panic recovery
│   └── storages/
│       ├── storage.go          # Интерфейс хранилища
│       └── postgres/
│           └── connector.go    # PostgreSQL connector
├── docs/
│   ├── docs.go                 # Swagger docs
│   ├── swagger.json
│   └── swagger.yaml
├── config.env                  # Конфигурация
├── Dockerfile
└── README.md
```

## API Endpoints

### Аутентификация

#### POST /api/v1/register
Регистрация нового пользователя

**Request:**
```json
{
  "username": "user1",
  "password": "password123",
  "email": "user1@example.com"
}
```

**Response (201):**
```json
{
  "message": "User registered successfully"
}
```

#### POST /api/v1/login
Авторизация пользователя

**Request:**
```json
{
  "username": "user1",
  "password": "password123"
}
```

**Response (200):**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

### Кошелек (требуется JWT токен)

#### GET /api/v1/balance
Получение баланса пользователя

**Headers:**
```
Authorization: Bearer <JWT_TOKEN>
```

**Response (200):**
```json
{
  "balance": {
    "USD": 1000.50,
    "RUB": 50000.00,
    "EUR": 850.25
  }
}
```

#### POST /api/v1/wallet/deposit
Пополнение счета

**Request:**
```json
{
  "amount": 100.00,
  "currency": "USD"
}
```

**Response (200):**
```json
{
  "message": "Account topped up successfully",
  "new_balance": {
    "USD": 1100.50,
    "RUB": 50000.00,
    "EUR": 850.25
  }
}
```

#### POST /api/v1/wallet/withdraw
Вывод средств

**Request:**
```json
{
  "amount": 50.00,
  "currency": "USD"
}
```

**Response (200):**
```json
{
  "message": "Withdrawal successful",
  "new_balance": {
    "USD": 1050.50,
    "RUB": 50000.00,
    "EUR": 850.25
  }
}
```

### Обмен валют

#### GET /api/v1/exchange/rates
Получение курсов валют

**Response (200):**
```json
{
  "rates": {
    "USD": 1.0,
    "RUB": 90.5,
    "EUR": 0.85
  }
}
```

#### POST /api/v1/exchange
Обмен валют

**Request:**
```json
{
  "from_currency": "USD",
  "to_currency": "EUR",
  "amount": 100.00
}
```

**Response (200):**
```json
{
  "message": "Exchange successful",
  "exchanged_amount": 85.00,
  "new_balance": {
    "USD": 950.50,
    "EUR": 935.25
  }
}
```

## Конфигурация

Файл `config.env`:

```env
# Application
APP_PORT=8080

# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=wallet
DB_PASSWORD=wallet
DB_NAME=wallet

# JWT
JWT_SECRET=your-secret-key-change-in-production

# gRPC Exchanger Service
EXCHANGER_GRPC_HOST=localhost
EXCHANGER_GRPC_PORT=50051

# Kafka
KAFKA_BROKERS=localhost:9092
KAFKA_TOPIC=large-transactions
```

## Локальный запуск

### 1. Установка зависимостей

```bash
go mod download
```

### 2. Запуск PostgreSQL

```bash
docker run -d \
  --name postgres-wallet \
  -e POSTGRES_USER=wallet \
  -e POSTGRES_PASSWORD=wallet \
  -e POSTGRES_DB=wallet \
  -p 5432:5432 \
  postgres:15
```

### 3. Применение миграций

```bash
# Создание таблиц
psql -h localhost -U wallet -d wallet -f migrations/001_init.sql
```

### 4. Генерация Swagger документации

```bash
swag init -g cmd/main.go -o docs
```

### 5. Сборка и запуск

```bash
GOOS=linux GOARCH=amd64 go build -o main ./cmd
./main -c config.env
```

Или через Docker:

```bash
docker build -t gw-currency-wallet .
docker run -p 8080:8080 --env-file config.env gw-currency-wallet
```

## Swagger UI

После запуска сервиса Swagger UI доступен по адресу:
```
http://localhost:8080/swagger/index.html
```

## Логирование

Сервис использует structured logging (Zap) в формате JSON:

```json
{
  "level": "info",
  "ts": "2026-03-01T20:00:00.000Z",
  "caller": "service/wallet_service.go:45",
  "msg": "deposit successful",
  "user_id": "123",
  "amount": 100.00,
  "currency": "USD"
}
```

Уровни логирования:
- **DEBUG**: Детальная информация для отладки
- **INFO**: Общая информация о работе сервиса
- **WARN**: Предупреждения
- **ERROR**: Ошибки

## Тестирование

```bash
# Запуск всех тестов
go test ./...

# Запуск с покрытием
go test -cover ./...

# Запуск конкретного теста
go test -v ./internal/service -run TestWalletService
```

## Производительность

- Время отклика: < 200ms для большинства операций
- Кэширование курсов валют для снижения нагрузки на gRPC
- Connection pooling для PostgreSQL

## Безопасность

- JWT токены для аутентификации
- Bcrypt для хеширования паролей
- Валидация всех входных данных
- Rate limiting (рекомендуется настроить на уровне reverse proxy)

## Kafka Integration

При транзакциях > 30000 в любой валюте, сервис отправляет сообщение в Kafka топик `large-transactions`:

```json
{
  "user_id": "123",
  "amount": 35000.00,
  "currency": "USD",
  "type": "deposit",
  "timestamp": "2026-03-01T20:00:00Z"
}
```

## Troubleshooting

### Ошибка подключения к БД
```bash
# Проверьте доступность PostgreSQL
psql -h localhost -U wallet -d wallet

# Проверьте переменные окружения
echo $DB_HOST $DB_PORT
```

### Ошибка gRPC соединения
```bash
# Проверьте доступность gw-exchanger
grpcurl -plaintext localhost:50051 list
```

### Проблемы с JWT
```bash
# Проверьте JWT_SECRET в config.env
# Убедитесь, что токен не истек
```

## Миграции

Создайте файл `migrations/001_init.sql`:

```sql
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(255) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE wallets (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id),
    currency VARCHAR(3) NOT NULL,
    balance DECIMAL(15, 2) DEFAULT 0.00,
    UNIQUE(user_id, currency)
);

CREATE TABLE transactions (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id),
    type VARCHAR(50) NOT NULL,
    amount DECIMAL(15, 2) NOT NULL,
    currency VARCHAR(3) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_wallets_user_id ON wallets(user_id);
CREATE INDEX idx_transactions_user_id ON transactions(user_id);
```

## Лицензия

MIT
