package repository

import "database/sql"

type Repositories struct {
	UserRepository   *UserRepository
	WalletRepository *WalletRepository
}

func NewRepositories(db *sql.DB) *Repositories {
	return &Repositories{
		UserRepository:   NewUserRepository(db),
		WalletRepository: NewWalletRepository(db),
	}
}
