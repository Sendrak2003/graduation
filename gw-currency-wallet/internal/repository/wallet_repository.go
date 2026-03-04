package repository

import (
	"context"
	"database/sql"
)

type WalletRepository struct {
	db *sql.DB
}

func NewWalletRepository(db *sql.DB) *WalletRepository {
	return &WalletRepository{db: db}
}

func (r *WalletRepository) GetBalance(ctx context.Context, userID string, currency string) (float64, error) {
	var balance float64
	err := r.db.QueryRowContext(
		ctx,
		`SELECT balance FROM wallets WHERE user_id = $1 AND currency = $2`,
		userID,
		currency,
	).Scan(&balance)

	if err == sql.ErrNoRows {
		return 0, nil
	}
	if err != nil {
		if isInvalidUUIDError(err) {
			return 0, ErrInvalidUUID
		}
		return 0, err
	}

	return balance, nil
}

func (r *WalletRepository) Deposit(ctx context.Context, userID string, currency string, amount float64) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var walletID string
	err = tx.QueryRowContext(
		ctx,
		`SELECT id FROM wallets WHERE user_id = $1 AND currency = $2 FOR UPDATE`,
		userID,
		currency,
	).Scan(&walletID)

	if err == sql.ErrNoRows {
		_, err = tx.ExecContext(
			ctx,
			`INSERT INTO wallets (user_id, currency, balance) VALUES ($1, $2, $3)`,
			userID,
			currency,
			amount,
		)
		if err != nil {
			if isInvalidUUIDError(err) {
				return ErrInvalidUUID
			}
			return err
		}
	} else if err != nil {
		if isInvalidUUIDError(err) {
			return ErrInvalidUUID
		}
		return err
	} else {
		_, err = tx.ExecContext(
			ctx,
			`UPDATE wallets SET balance = balance + $1 WHERE id = $2`,
			amount,
			walletID,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *WalletRepository) Withdraw(ctx context.Context, userID string, currency string, amount float64) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var walletID string
	var balance float64
	err = tx.QueryRowContext(
		ctx,
		`SELECT id, balance FROM wallets WHERE user_id = $1 AND currency = $2 FOR UPDATE`,
		userID,
		currency,
	).Scan(&walletID, &balance)

	if err == sql.ErrNoRows {
		return ErrWalletNotFound
	}
	if err != nil {
		if isInvalidUUIDError(err) {
			return ErrInvalidUUID
		}
		return err
	}

	if balance < amount {
		return ErrInsufficientFunds
	}

	_, err = tx.ExecContext(
		ctx,
		`UPDATE wallets SET balance = balance - $1 WHERE id = $2`,
		amount,
		walletID,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}
