package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/112Alex/demo-service.git/internal/model"
)

type DBClient struct {
	db *sql.DB
}

// NewDBClient создает и возвращает новый клиент БД.
func NewDBClient(connStr string) (*DBClient, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("ошибка при открытии соединения с БД: %w", err)
	}

	// Configure connection pool sizes from env or defaults
	maxOpen := getEnvAsInt("DB_MAX_OPEN_CONNS", 25)
	maxIdle := getEnvAsInt("DB_MAX_IDLE_CONNS", 25)
	maxLifeMinutes := getEnvAsInt("DB_CONN_MAX_LIFETIME_MIN", 5)

	db.SetMaxOpenConns(maxOpen)
	db.SetMaxIdleConns(maxIdle)
	db.SetConnMaxLifetime(time.Duration(maxLifeMinutes) * time.Minute)

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("ошибка при проверке соединения с БД: %w", err)
	}

	log.Println("Успешно подключено к PostgreSQL!")
	return &DBClient{db: db}, nil
}

// Close закрывает соединение с БД.
func (c *DBClient) Close() error {
	return c.db.Close()
}

// SaveOrder сохраняет полную информацию о заказе в БД, используя транзакцию.
func (c *DBClient) SaveOrder(ctx context.Context, order *model.Order) error {
	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("не удалось начать транзакцию: %w", err)
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `
		INSERT INTO orders (order_uid, track_number, entry, locale, internal_signature, customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (order_uid) DO NOTHING`,
		order.OrderUID, order.TrackNumber, order.Entry, order.Locale, order.InternalSignature, order.CustomerID, order.DeliveryService, order.Shardkey, order.SmID, order.DateCreated, order.OofShard)
	if err != nil {
		return fmt.Errorf("не удалось сохранить заказ: %w", err)
	}

	// Сохранение информации о доставке
	_, err = tx.ExecContext(ctx, `
		INSERT INTO delivery (order_uid, name, phone, zip, city, address, region, email)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (order_uid) DO NOTHING`,
		order.OrderUID, order.Delivery.Name, order.Delivery.Phone, order.Delivery.Zip, order.Delivery.City, order.Delivery.Address, order.Delivery.Region, order.Delivery.Email)
	if err != nil {
		return fmt.Errorf("не удалось сохранить доставку: %w", err)
	}

	// Сохранение информации об оплате
	_, err = tx.ExecContext(ctx, `
		INSERT INTO payment (transaction, order_uid, request_id, currency, provider, amount, payment_dt, bank, delivery_cost, goods_total, custom_fee)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (transaction) DO NOTHING`,
		order.Payment.Transaction, order.OrderUID, order.Payment.RequestID, order.Payment.Currency, order.Payment.Provider, order.Payment.Amount, order.Payment.PaymentDt, order.Payment.Bank, order.Payment.DeliveryCost, order.Payment.GoodsTotal, order.Payment.CustomFee)
	if err != nil {
		return fmt.Errorf("не удалось сохранить оплату: %w", err)
	}

	// Сохранение товаров
	for _, item := range order.Items {
		_, err = tx.ExecContext(ctx, `
			INSERT INTO items (chrt_id, order_uid, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
			ON CONFLICT (chrt_id) DO NOTHING`,
			item.ChrtID, order.OrderUID, item.TrackNumber, item.Price, item.Rid, item.Name, item.Sale, item.Size, item.TotalPrice, item.NmID, item.Brand, item.Status)
		if err != nil {
			return fmt.Errorf("не удалось сохранить товар %d: %w", item.ChrtID, err)
		}
	}

	return tx.Commit() // Фиксация транзакции
}

// GetOrderFromDB загружает полную информацию о заказе из БД.
func (c *DBClient) GetOrderFromDB(ctx context.Context, orderUID string) (*model.Order, error) {
	order := &model.Order{}

	// Загрузка основной информации о заказе
	err := c.db.QueryRowContext(ctx, `
		SELECT order_uid, track_number, entry, locale, internal_signature, customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard
		FROM orders WHERE order_uid = $1`, orderUID).
		Scan(&order.OrderUID, &order.TrackNumber, &order.Entry, &order.Locale, &order.InternalSignature, &order.CustomerID, &order.DeliveryService, &order.Shardkey, &order.SmID, &order.DateCreated, &order.OofShard)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Заказ не найден
		}
		return nil, fmt.Errorf("ошибка при получении заказа: %w", err)
	}

	// Загрузка информации о доставке
	err = c.db.QueryRowContext(ctx, `
		SELECT name, phone, zip, city, address, region, email
		FROM delivery WHERE order_uid = $1`, orderUID).
		Scan(&order.Delivery.Name, &order.Delivery.Phone, &order.Delivery.Zip, &order.Delivery.City, &order.Delivery.Address, &order.Delivery.Region, &order.Delivery.Email)
	if err != nil {
		return nil, fmt.Errorf("ошибка при получении доставки: %w", err)
	}

	// Загрузка информации об оплате
	err = c.db.QueryRowContext(ctx, `
		SELECT transaction, request_id, currency, provider, amount, payment_dt, bank, delivery_cost, goods_total, custom_fee
		FROM payment WHERE order_uid = $1`, orderUID).
		Scan(&order.Payment.Transaction, &order.Payment.RequestID, &order.Payment.Currency, &order.Payment.Provider, &order.Payment.Amount, &order.Payment.PaymentDt, &order.Payment.Bank, &order.Payment.DeliveryCost, &order.Payment.GoodsTotal, &order.Payment.CustomFee)
	if err != nil {
		return nil, fmt.Errorf("ошибка при получении оплаты: %w", err)
	}

	// Загрузка товаров
	rows, err := c.db.QueryContext(ctx, `
		SELECT chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status
		FROM items WHERE order_uid = $1`, orderUID)
	if err != nil {
		return nil, fmt.Errorf("ошибка при получении товаров: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		item := model.Item{OrderUID: orderUID}
		if err := rows.Scan(&item.ChrtID, &item.TrackNumber, &item.Price, &item.Rid, &item.Name, &item.Sale, &item.Size, &item.TotalPrice, &item.NmID, &item.Brand, &item.Status); err != nil {
			return nil, fmt.Errorf("ошибка при сканировании товара: %w", err)
		}
		order.Items = append(order.Items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка итерации по товарам: %w", err)
	}

	return order, nil
}

// GetAllOrders загружает все заказы из БД. Используется для восстановления кеша.
func (c *DBClient) GetAllOrders(ctx context.Context) ([]*model.Order, error) {
	var orders []*model.Order

	// Загружаем все order_uid из таблицы orders
	rows, err := c.db.QueryContext(ctx, `SELECT order_uid FROM orders`)
	if err != nil {
		return nil, fmt.Errorf("ошибка при получении списка заказов: %w", err)
	}
	defer rows.Close()

	var orderUIDs []string
	for rows.Next() {
		var orderUID string
		if err := rows.Scan(&orderUID); err != nil {
			return nil, fmt.Errorf("ошибка при сканировании order_uid: %w", err)
		}
		orderUIDs = append(orderUIDs, orderUID)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка итерации по заказам: %w", err)
	}

	// Загружаем полную информацию для каждого заказа
	for _, orderUID := range orderUIDs {
		order, err := c.GetOrderFromDB(ctx, orderUID)
		if err != nil {
			log.Printf("Ошибка при загрузке заказа %s: %v", orderUID, err)
			continue // Пропускаем проблемный заказ
		}
		if order != nil {
			orders = append(orders, order)
		}
	}

	return orders, nil
}

// helper
func getEnvAsInt(key string, defaultVal int) int {
	if v, ok := os.LookupEnv(key); ok {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return defaultVal
}
