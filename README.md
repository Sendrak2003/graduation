# Microservices Architecture - Currency Wallet System

Система из трех микросервисов для управления кошельком, обмена валют и уведомлений о крупных транзакциях.

## Архитектура

- **gw-currency-wallet** - RESTful API для управления кошельком и обмена валют
- **gw-exchanger** - gRPC сервис для получения курсов валют
- **gw-notification** - Kafka consumer для обработки крупных транзакций

## Технологии

- Go 1.21+
- PostgreSQL 15
- MongoDB 7
- Kafka
- gRPC
- Docker & Docker Compose

## Быстрый старт

### 1. Клонирование и настройка

```bash
# Создайте .env файл из примера
cp .env.example .env

# Отредактируйте .env при необходимости
```

### 2. Генерация protobuf кода

```bash
make proto
```

### 3. Запуск всех сервисов

```bash
# Сборка образов
make build

# Запуск всех сервисов
make up

# Просмотр логов
make logs
```

### 4. Проверка работоспособности

Сервисы будут доступны по следующим адресам:

- **gw-currency-wallet**: http://localhost:8080
- **gw-exchanger**: localhost:50051 (gRPC)
- **PostgreSQL (wallet)**: localhost:5432
- **PostgreSQL (exchanger)**: localhost:5433
- **MongoDB**: localhost:27017
- **Kafka**: localhost:9092

## API Endpoints

### Регистрация
```bash
curl -X POST http://localhost:8080/api/v1/register \
  -H "Content-Type: application/json" \
  -d '{"username":"user1","password":"pass123","email":"user1@example.com"}'
```

### Авторизация
```bash
curl -X POST http://localhost:8080/api/v1/login \
  -H "Content-Type: application/json" \
  -d '{"username":"user1","password":"pass123"}'
```

### Получение баланса
```bash
curl -X GET http://localhost:8080/api/v1/balance \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### Пополнение счета
```bash
curl -X POST http://localhost:8080/api/v1/wallet/deposit \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"amount":100.00,"currency":"USD"}'
```

### Вывод средств
```bash
curl -X POST http://localhost:8080/api/v1/wallet/withdraw \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"amount":50.00,"currency":"USD"}'
```

### Получение курсов валют
```bash
curl -X GET http://localhost:8080/api/v1/exchange/rates \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### Обмен валют
```bash
curl -X POST http://localhost:8080/api/v1/exchange \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"from_currency":"USD","to_currency":"EUR","amount":100.00}'
```

## Локальная разработка

### Запуск отдельного сервиса

#### gw-exchanger
```bash
cd gw-exchanger
GOOS=linux GOARCH=amd64 go build -o main ./cmd
./main -c config.env
```

#### gw-currency-wallet
```bash
cd gw-currency-wallet
GOOS=linux GOARCH=amd64 go build -o main ./cmd
./main -c config.env
```

#### gw-notification
```bash
cd gw-notification
GOOS=linux GOARCH=amd64 go build -o main ./cmd
./main
```

## Остановка и очистка

```bash
# Остановка всех сервисов
make down

# Полная очистка (включая volumes и images)
make clean
```

## Структура проекта

```
.
├── docker-compose.yml
├── .env.example
├── Makefile
├── gw-currency-wallet/
│   ├── cmd/
│   ├── internal/
│   ├── docs/
│   ├── Dockerfile
│   └── config.env
├── gw-exchanger/
│   ├── cmd/
│   ├── internal/
│   ├── Dockerfile
│   └── config.env
├── gw-notification/
│   ├── cmd/
│   ├── internal/
│   └── Dockerfile
└── proto-exchange/
    └── exchange/
        └── exchange.proto
```

## Миграции базы данных

Миграции должны быть созданы для каждого сервиса отдельно. Рекомендуется использовать инструменты типа `golang-migrate` или `goose`.

## Мониторинг и логирование

Все сервисы используют structured logging в формате JSON. Логи можно просматривать через:

```bash
# Все сервисы
docker-compose logs -f

# Конкретный сервис
docker-compose logs -f gw-currency-wallet
```

## Troubleshooting

### Проблемы с подключением к базе данных
Убедитесь, что контейнеры баз данных запущены и healthy:
```bash
docker-compose ps
```

### Проблемы с Kafka
Проверьте, что Zookeeper и Kafka запущены:
```bash
docker-compose logs kafka
docker-compose logs zookeeper
```

### Пересоздание контейнеров
```bash
docker-compose down
docker-compose up -d --force-recreate
```
