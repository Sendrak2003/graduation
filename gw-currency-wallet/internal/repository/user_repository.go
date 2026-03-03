package repository

import (
	"context"
	"database/sql"
	"gw-currency-wallet/internal/models"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) CreateUser(
	ctx context.Context,
	userID string,
	username string,
	email string,
	password_hash string,
) error {
	_, err := r.db.ExecContext(
		ctx,
		`INSERT INTO users (id, username, email, password_hash) VALUES ($1, $2, $3, $4)`,
		userID,
		username,
		email,
		password_hash,
	)
	return err
}

func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	var user models.User
	err := r.db.QueryRowContext(
		ctx,
		`SELECT id, username, email, password_hash, created_at, updated_at 
		 FROM users WHERE username = $1`,
		username,
	).Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) GetByID(ctx context.Context, userID string) (*models.User, error) {
	var user models.User
	err := r.db.QueryRowContext(
		ctx,
		`SELECT id, username, email, password_hash, created_at, updated_at 
		 FROM users WHERE id = $1`,
		userID,
	).Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		return nil, err
	}
	return &user, nil
}
