.PHONY: help build up up-build down restart logs clean clean-all test-db init ps

help: ## Показать справку
	@grep -E '^[a-zA-Z_-]+:.*?## .*$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

up: ## Запуск всех сервисов (автоматически применяет миграции)
	@echo "Starting all services..."
	docker-compose up -d
	@echo "Services started!"
	@echo "Waiting for services to be healthy..."
	@sleep 15
	@docker-compose ps

up-build: ## Сборка и запуск всех сервисов
	@echo "Building and starting all services..."
	docker-compose up -d --build
	@echo "Services built and started!"
	@echo "Waiting for services to be healthy..."
	@sleep 15
	@docker-compose ps

build: ## Сборка Docker образов
	@echo "Building Docker images..."
	docker-compose build
	@echo "Build complete!"

down: ## Остановка всех сервисов
	@echo "Stopping services..."
	docker-compose down
	@echo "Services stopped!"

restart: down up ## Перезапуск всех сервисов

clean: ## Очистка контейнеров и volumes
	@echo "Cleaning up containers and volumes..."
	docker-compose down -v
	docker system prune -f
	@echo "Cleanup complete!"

clean-all: ## Полная очистка (включая images)
	@echo "Full cleanup..."
	docker-compose down -v --rmi all
	docker system prune -af
	@echo "Full cleanup complete!"

logs: ## Просмотр логов всех сервисов
	docker-compose logs -f

logs-wallet: ## Просмотр логов gw-currency-wallet
	docker-compose logs -f gw-currency-wallet

logs-exchanger: ## Просмотр логов gw-exchanger
	docker-compose logs -f gw-exchanger

logs-notification: ## Просмотр логов gw-notification
	docker-compose logs -f gw-notification

ps: ## Показать статус контейнеров
	docker-compose ps

test-db: ## Проверить подключение к базам данных
	@echo "Testing database connections..."
	@echo "Wallet DB:"
	@docker exec postgres-wallet psql -U wallet -d wallet -c "SELECT 'Wallet DB OK' as status;" 2>/dev/null || echo "Wallet DB: FAILED"
	@echo "Exchanger DB:"
	@docker exec postgres-exchanger psql -U exchanger -d exchanger -c "SELECT 'Exchanger DB OK' as status, COUNT(*) as exchange_rates FROM exchange_rates;" 2>/dev/null || echo "Exchanger DB: FAILED"
	@echo "MongoDB:"
	@docker exec mongodb mongosh --eval "db.adminCommand('ping')" --quiet 2>/dev/null && echo "MongoDB: OK" || echo "MongoDB: FAILED"

init: ## Первоначальная настройка системы
	@echo "Initializing system..."
	make clean
	make up-build
	@echo "Waiting for services to initialize..."
	@sleep 30
	make test-db
	@echo "System ready!"