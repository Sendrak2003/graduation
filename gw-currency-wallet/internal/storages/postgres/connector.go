package postgres

import "database/sql"

type PostgresConnector struct {
	db *sql.DB
}

func NewPostgresConnector(db *sql.DB) *PostgresConnector {
	return &PostgresConnector{db: db}
}
