package account

import (
	"context"
	"database/sql"

	_ "github.com/lib/pq"
)

type Repository interface {
	Close()
	Put(ctx context.Context, a Account) error
	GetById(ctx context.Context, id string) (*Account, error)
	List(ctx context.Context, skip uint64, take uint64) ([]Account, error)
}

type postgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(url string) (Repository, error) {
	db, err := sql.Open("postgres", url)
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return &postgresRepository{db}, nil
}

func (r *postgresRepository) Close() {
	r.db.Close()
}

func (r *postgresRepository) Ping() error {
	return r.db.Ping()
}

func (r *postgresRepository) Put(ctx context.Context, a Account) error {
	_, err := r.db.ExecContext(ctx, "INSERT INTO account (id, name) VALUES ($1, $2)", a.ID, a.Name)
	if err != nil {
		return err
	}
	return nil
}

func (r *postgresRepository) GetById(ctx context.Context, id string) (*Account, error) {
	row := r.db.QueryRowContext(ctx, "SELECT id, name FROM account WHERE id = $1", id)
	a := &Account{}
	if err := row.Scan(&a.ID, &a.Name); err != nil {
		return nil, err
	}
	return a, nil
}

func (r *postgresRepository) List(ctx context.Context, skip uint64, take uint64) ([]Account, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT id, name FROM account DESC OFFSET $1 LIMIT $2", skip, take)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	accounts := []Account{}
	for rows.Next() {
		a := &Account{}
		if err := rows.Scan(&a.ID, &a.Name); err != nil {
			return nil, err
		}
		accounts = append(accounts, *a)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return accounts, nil
}
