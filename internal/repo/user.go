package repo

import (
	. "audiodrive/internal/models"
	"context"
	"database/sql"
)

type UserRepo struct {
	db *sql.DB
}

func NewUserRepo(db *sql.DB) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) Create(ctx context.Context, email string) (*User, error) {
	var user User
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO users (email) VALUES ($1) RETURNING id, email`,
		email,
	).Scan(&user.Id, &user.Email)

	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepo) GetById(ctx context.Context, id int64) (*User, error) {
	var user User
	err := r.db.QueryRowContext(ctx,
		`SELECT id, email FROM users WHERE id = $1`,
		id,
	).Scan(&user.Id, &user.Email)
	if err != nil {
		return nil, err
	}
	return &user, nil
}
