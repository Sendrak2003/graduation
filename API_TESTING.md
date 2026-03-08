# API Testing Guide

## Запуск системы

```bash
docker-compose up -d
```

## Swagger UI

Откройте в браузере: http://localhost:8080/swagger/index.html

## Тестовые данные

### Курсы валют (автоматически загружаются в БД)

- USD → RUB: 90.5
- USD → EUR: 0.92
- EUR → RUB: 98.3
- EUR → USD: 1.09
- RUB → USD: 0.011
- RUB → EUR: 0.010

## Пошаговое тестирование

### 1. Регистрация пользователя

**POST** `/api/v1/register`

```json
{
  "username": "testuser",
  "email": "test@example.com",
  "password": "password123"
}
```

**Ответ:**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

Скопируйте `access_token` для следующих запросов.

### 2. Авторизация (если уже зарегистрированы)

**POST** `/api/v1/login`

```json
{
  "username": "testuser",
  "password": "password123"
}
```

**Ответ:**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

### 3. Настройка Authorization в Swagger

1. Нажмите кнопку **Authorize** (зеленая кнопка с замком)
2. Введите: `Bearer <ваш_токен>`
3. Нажмите **Authorize**

### 4. Проверка баланса

**GET** `/api/v1/balance`

**Ответ:**
```json
{
  "balance": {
    "EUR": 0,
    "RUB": 0,
    "USD": 0
  }
}
```

### 5. Пополнение кошелька

**POST** `/api/v1/wallet/deposit`

```json
{
  "currency": "USD",
  "amount": 1000
}
```

**Ответ:**
```json
{
  "message": "Account topped up successfully",
  "new_balance": {
    "EUR": 0,
    "RUB": 0,
    "USD": 1000
  }
}
```

### 6. Получение курсов валют

**GET** `/api/v1/exchange/rates`

**Ответ:**
```json
{
  "rates": {
    "EUR_RUB": 98.3,
    "EUR_USD": 1.09,
    "RUB_EUR": 0.01,
    "RUB_USD": 0.011,
    "USD_EUR": 0.92,
    "USD_RUB": 90.5
  }
}
```

### 7. Обмен валюты

**POST** `/api/v1/exchange`

```json
{
  "from_currency": "USD",
  "to_currency": "RUB",
  "amount": 100
}
```

**Ответ:**
```json
{
  "exchanged_amount": 9050,
  "message": "Exchange successful",
  "new_balance": {
    "RUB": "updated",
    "USD": "updated"
  }
}
```

### 8. Снятие средств

**POST** `/api/v1/wallet/withdraw`

```json
{
  "currency": "RUB",
  "amount": 1000
}
```

**Ответ:**
```json
{
  "message": "Withdrawal successful",
  "new_balance": {
    "EUR": 0,
    "RUB": 8050,
    "USD": 900
  }
}
```

## Тестирование больших транзакций (Kafka)

Для транзакций >= 30000 в любой валюте, система отправляет уведомление в Kafka.

### Пример большой транзакции:

**POST** `/api/v1/wallet/deposit`

```json
{
  "currency": "USD",
  "amount": 50000
}
```

Эта транзакция будет отправлена в Kafka топик `large-transactions` и обработана сервисом уведомлений.

### Проверка уведомлений в MongoDB:

```bash
docker exec -it mongodb mongosh -u admin -p admin

use notifications
db.transactions.find().pretty()
```

## Проверка логов

### Wallet Service:
```bash
docker logs gw-currency-wallet
```

### Exchanger Service:
```bash
docker logs gw-exchanger
```

### Notification Service:
```bash
docker logs gw-notification
```

## Запуск Unit Tests

### Wallet Service:
```bash
cd gw-currency-wallet
go test ./internal/service/... -v
go test ./internal/grpc/... -v
```

### Exchanger Service:
```bash
cd gw-exchanger
go test ./internal/service/... -v
go test ./internal/grpc/... -v
```

## Остановка системы

```bash
docker-compose down
```

## Полная очистка (включая данные):

```bash
docker-compose down -v
```
