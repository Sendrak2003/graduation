-- Тестовые данные для курсов валют
INSERT INTO exchange_rates (from_currency, to_currency, rate, updated_at) VALUES
('USD', 'RUB', 90.5, NOW()),
('USD', 'EUR', 0.92, NOW()),
('EUR', 'RUB', 98.3, NOW()),
('EUR', 'USD', 1.09, NOW()),
('RUB', 'USD', 0.011, NOW()),
('RUB', 'EUR', 0.010, NOW())
ON CONFLICT (from_currency, to_currency) DO UPDATE 
SET rate = EXCLUDED.rate, updated_at = NOW();
