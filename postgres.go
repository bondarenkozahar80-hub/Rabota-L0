package postgres

import (
	"context"
	"database/sql"
	"encoding/json "
	"fmt"
	"order-service/internal/domain"

	_ "github.com/lib/pq"
)

type OrderRepo struct {
	db *sql.DB
}

func NewOrderRepo(connStr string) (*OrderRepo, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &OrderRepo{db: db}, nil
}

func (r *OrderRepo) Create(ctx context.Context, order *domain.Order) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Insert delivery
	deliveryQuery := `INSERT INTO deliveries (order_uid, name, phone, zip, city, address, region, email) 
	                  VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	_, err = tx.ExecContext(ctx, deliveryQuery,
		order.OrderUID, order.Delivery.Name, order.Delivery.Phone, order.Delivery.Zip,
		order.Delivery.City, order.Delivery.Address, order.Delivery.Region, order.Delivery.Email)
	if err != nil {
		return fmt.Errorf("failed to insert delivery: %w", err)
	}

	// Insert payment
	paymentQuery := `INSERT INTO payments (order_uid, transaction, request_id, currency, provider, 
	                  amount, payment_dt, bank, delivery_cost, goods_total, custom_fee) 
	                  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`
	_, err = tx.ExecContext(ctx, paymentQuery,
		order.OrderUID, order.Payment.Transaction, order.Payment.RequestID, order.Payment.Currency,
		order.Payment.Provider, order.Payment.Amount, order.Payment.PaymentDT, order.Payment.Bank,
		order.Payment.DeliveryCost, order.Payment.GoodsTotal, order.Payment.CustomFee)
	if err != nil {
		return fmt.Errorf("failed to insert payment: %w", err)
	}

	// Insert items
	itemQuery := `INSERT INTO items (order_uid, chrt_id, track_number, price, rid, name, sale, size, 
	                total_price, nm_id, brand, status) 
	                VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`
	for _, item := range order.Items {
		_, err = tx.ExecContext(ctx, itemQuery,
			order.OrderUID, item.ChrtID, item.TrackNumber, item.Price, item.RID, item.Name,
			item.Sale, item.Size, item.TotalPrice, item.NMID, item.Brand, item.Status)
		if err != nil {
			return fmt.Errorf("failed to insert item: %w", err)
		}
	}

	// Insert order
	orderQuery := 'INSERT INTO orders ( order_uid, track_number, entry, locale, internal_signature, customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard )VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)'
	_, err = tx.ExecContext(ctx, orderQuery, order.OrderUID, order.TrackNumber, order.Entry, order.Locale, order.InternalSignature,order.CustomerID, order.DeliveryService, order.ShardKey, order.SMID, order.DateCreated, order.OOFShard)
	if err != nil {
		return fmt.Errorf("failed to insert order: %w", err)
	}

	return tx.Commit()
}

func (r *OrderRepo) GetByUID(ctx context.Context, orderUID string) (*domain.Order, error) {
	query := `
		SELECT o.order_uid, o.track_number, o.entry, o.locale, o.internal_signature, 
		       o.customer_id, o.delivery_service, o.shardkey, o.sm_id, o.date_created, o.oof_shard,
		       d.name, d.phone, d.zip, d.city, d.address, d.region, d.email,
		       p.transaction, p.request_id, p.currency, p.provider, p.amount, p.payment_dt, 
		       p.bank, p.delivery_cost, p.goods_total, p.custom_fee
		FROM orders o
		LEFT JOIN deliveries d ON o.order_uid = d.order_uid
		LEFT JOIN payments p ON o.order_uid = p.order_uid
		WHERE o.order_uid = $1`

	var order domain.Order
	var delivery domain.Delivery
	var payment domain.Payment

	err := r.db.QueryRowContext(ctx, query, orderUID).Scan(
		&order.OrderUID, &order.TrackNumber, &order.Entry, &order.Locale, &order.InternalSignature,
		&order.CustomerID, &order.DeliveryService, &order.ShardKey, &order.SMID, &order.DateCreated, &order.OOFShard,
		&delivery.Name, &delivery.Phone, &delivery.Zip, &delivery.City, &delivery.Address, &delivery.Region, &delivery.Email,
		&payment.Transaction, &payment.RequestID, &payment.Currency, &payment.Provider, &payment.Amount,
		&payment.PaymentDT, &payment.Bank, &payment.DeliveryCost, &payment.GoodsTotal, &payment.CustomFee,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	order.Delivery = delivery
	order.Payment = payment

	// Get items
	itemsQuery := `SELECT chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status 
	               FROM items WHERE order_uid = $1`
	rows, err := r.db.QueryContext(ctx, itemsQuery, orderUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get items: %w", err)
	}
	defer rows.Close()

	var items []domain.Item
	for rows.Next() {
		var item domain.Item
		err := rows.Scan(
			&item.ChrtID, &item.TrackNumber, &item.Price, &item.RID, &item.Name, &item.Sale,
			&item.Size, &item.TotalPrice, &item.NMID, &item.Brand, &item.Status,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan item: %w", err)
		}
		items = append(items, item)
	}

	order.Items = items
	return &order, nil
}

func (r *OrderRepo) GetAll(ctx context.Context) ([]*domain.Order, error) {
	query := `SELECT order_uid FROM orders`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get order UIDs: %w", err)
	}
	defer rows.Close()

	var orders []*domain.Order
	for rows.Next() {
		var orderUID string
		if err := rows.Scan(&orderUID); err != nil {
			return nil, fmt.Errorf("failed to scan order UID: %w", err)
		}
		order, err := r.GetByUID(ctx, orderUID)
		if err != nil {
			return nil, fmt.Errorf("failed to get order %s: %w", orderUID, err)
		}
		orders = append(orders, order)
	}

	return orders, nil
}

func (r *OrderRepo) Close() error {
	return r.db.Close()
}
