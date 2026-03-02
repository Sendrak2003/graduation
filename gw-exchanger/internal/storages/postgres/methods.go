package postgres

import (
	"context"
	"database/sql"
	"fmt"
)

func (s *PostgresStorage) GetExchangeRate(ctx context.Context, from, to string) (float64, error) {
	var rate float64
	query := `SELECT rate FROM exchange_rates WHERE from_currency = $1 AND to_currency = $2`

	err := s.db.QueryRowContext(ctx, query, from, to).Scan(&rate)
	if err == sql.ErrNoRows {
		return 0, fmt.Errorf("exchange rate not found for %s to %s", from, to)
	}
	if err != nil {
		return 0, fmt.Errorf("failed to get exchange rate: %w", err)
	}

	return rate, nil
}

func (s *PostgresStorage) GetAllExchangeRates(ctx context.Context) (map[string]float64, error) {
	rates := make(map[string]float64)
	query := `SELECT from_currency, to_currency, rate FROM exchange_rates`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query exchange rates: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var from, to string
		var rate float64
		if err := rows.Scan(&from, &to, &rate); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		key := fmt.Sprintf("%s_%s", from, to)
		rates[key] = rate
	}

	return rates, nil
}

func (s *PostgresStorage) UpdateExchangeRate(ctx context.Context, from, to string, rate float64) error {
	query := `
		INSERT INTO exchange_rates (from_currency, to_currency, rate, updated_at)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT (from_currency, to_currency)
		DO UPDATE SET rate = $3, updated_at = NOW()
	`

	_, err := s.db.ExecContext(ctx, query, from, to, rate)
	if err != nil {
		return fmt.Errorf("failed to update exchange rate: %w", err)
	}

	return nil
}
