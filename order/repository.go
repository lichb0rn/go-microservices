package order

import (
	"context"
	"database/sql"

	"github.com/lib/pq"
)

type Repository interface {
	Close()
	Put(ctx context.Context, o Order) error
	GetByAccountId(ctx context.Context, accountId string) ([]Order, error)
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

func (r *postgresRepository) Put(ctx context.Context, o Order) (err error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
		err = tx.Commit()
	}()

	_, err = tx.ExecContext(ctx,
		"INSERT INTO order (id, created_at, account_id, total_price) VALUES ($1, $2, $3, $4)",
		o.ID,
		o.CreatedAt,
		o.AccountId,
		o.TotalPrice,
	)

	if err != nil {
		return
	}

	stmt, _ := tx.PrepareContext(ctx, pq.CopyIn("order_product", "order_id", "product_id", "quantity"))
	for _, p := range o.Products {
		_, err = stmt.ExecContext(ctx, o.ID, p.ID, p.Quantity)
		if err != nil {
			return
		}
	}
	_, err = stmt.ExecContext(ctx)
	if err != nil {
		return
	}
	stmt.Close()

	return
}

func (r *postgresRepository) GetByAccountId(ctx context.Context, accountId string) ([]Order, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT
		o.id,
		o.created_at,
		o.account_id,
		o.total_price::money::numeric::float8,
		op.product_id,
		op.quantity
		FROM orders o JOIN order_products op ON (o.id = op.order_id)
		WHERE o.account_id = $1
		ORDER BY o.id`,
		accountId,
	)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	orders := []Order{}
	order := &Order{}
	lastOrder := &Order{}
	orderedProduct := &OrderedProduct{}
	prodcuts := []OrderedProduct{}

	for rows.Next() {

		if err := rows.Scan(
			&order.ID,
			&order.CreatedAt,
			&order.AccountId,
			&order.CreatedAt,
			&order.TotalPrice,
			&orderedProduct.ID,
			&orderedProduct.Quantity,
		); err != nil {
			return nil, err
		}

		if lastOrder.ID != "" && lastOrder.ID != order.ID {
			newOrder := Order{
				ID:         lastOrder.ID,
				CreatedAt:  lastOrder.CreatedAt,
				AccountId:  lastOrder.AccountId,
				TotalPrice: lastOrder.TotalPrice,
				Products:   lastOrder.Products,
			}

			orders = append(orders, newOrder)
			prodcuts = []OrderedProduct{}
		}
		prodcuts = append(prodcuts, OrderedProduct{
			ID:       orderedProduct.ID,
			Quantity: orderedProduct.Quantity,
		})
		*lastOrder = *order
	}

	if lastOrder != nil {
		newOrder := Order{
			ID:         lastOrder.ID,
			CreatedAt:  lastOrder.CreatedAt,
			AccountId:  lastOrder.AccountId,
			TotalPrice: lastOrder.TotalPrice,
			Products:   lastOrder.Products,
		}

		orders = append(orders, newOrder)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return orders, nil
}
