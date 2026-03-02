-- Создание таблицы курсов валют
CREATE TABLE IF NOT EXISTS exchange_rates (
    id SERIAL PRIMARY KEY,
    from_currency VARCHAR(3) NOT NULL CHECK (from_currency IN ('USD', 'RUB', 'EUR')),
    to_currency VARCHAR(3) NOT NULL CHECK (to_currency IN ('USD', 'RUB', 'EUR')),
    rate DECIMAL(10, 6) NOT NULL CHECK (rate > 0),
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(from_currency, to_currency)
);

-- Индекс для быстрого поиска курсов
CREATE INDEX IF NOT EXISTS idx_exchange_rates_currencies 
ON exchange_rates(from_currency, to_currency);

-- Функция для автоматического обновления updated_at
CREATE OR REPLACE FUNCTION update_exchange_rates_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Триггер для автоматического обновления updated_at
CREATE TRIGGER update_exchange_rates_timestamp 
BEFORE UPDATE ON exchange_rates
FOR EACH ROW EXECUTE FUNCTION update_exchange_rates_updated_at();

-- Начальные данные курсов валют
INSERT INTO exchange_rates (from_currency, to_currency, rate) VALUES
-- USD курсы
('USD', 'USD', 1.000000),
('USD', 'RUB', 90.500000),
('USD', 'EUR', 0.850000),

-- RUB курсы
('RUB', 'USD', 0.011050),
('RUB', 'RUB', 1.000000),
('RUB', 'EUR', 0.009392),

-- EUR курсы
('EUR', 'USD', 1.176471),
('EUR', 'RUB', 106.470588),
('EUR', 'EUR', 1.000000)
ON CONFLICT (from_currency, to_currency) DO NOTHING;

-- Комментарии к таблице
COMMENT ON TABLE exchange_rates IS 'Таблица курсов обмена валют';
COMMENT ON COLUMN exchange_rates.from_currency IS 'Исходная валюта';
COMMENT ON COLUMN exchange_rates.to_currency IS 'Целевая валюта';
COMMENT ON COLUMN exchange_rates.rate IS 'Курс обмена';
COMMENT ON COLUMN exchange_rates.updated_at IS 'Время последнего обновления курса';
