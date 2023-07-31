package database

import (
	"context"
	"log"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	models "github.com/vlle/wb_L0/internal/models"
)

type DB struct {
	pool *pgxpool.Pool
}

func InitDB() DB {
	return DB{pool: createPool()}
}

func (d *DB) Close() {
	d.closePool()
}

// Creates connection pool
func createPool() *pgxpool.Pool {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		url = "postgres://postgres:postgres@localhost:5500/rec"
	}
	dbpool, err := pgxpool.New(context.Background(), url)
	if err != nil {
		log.Fatal("Unable to connect to pool", err)
	}
	return dbpool
}

// Closes connection pool
func (d *DB) closePool() {
	d.pool.Close()
}

// Inserts order into database
func (d *DB) SaveOrder(js models.Order) {
	delivery_id := d.insertDelivery(d.pool, &js.Delivery)
	payment_transaction := d.insertPayment(d.pool, &js.Payment)
	d.insertOrder(d.pool, &js, delivery_id, payment_transaction)
	d.bulkInsertItems(d.pool, js.Items, js.Order_uid)
}

// Loading orders from Postgres
func (d *DB) LoadOrders() []models.Order {
	conn := d.pool
	// if err != nil {
	//   log.Fatal(err)
	// }
	orders := make([]models.Order, 0)

	stmt := `select order_uid, track_number, entry,

           delivery.name, delivery.phone, delivery.zip, delivery.city,
           delivery.address, delivery.region, delivery.email,

           payment.transaction, coalesce(payment.request_id, ''), payment.currency,
           payment.provider, payment.amount, payment.payment_dt,
           payment.bank, payment.delivery_cost, payment.goods_total,
           payment.custom_fee,

           locale, internal_signature, customer_id, delivery_service, shardkey,
           sm_id, date_created, oof_shard

           from orders 
           join delivery on orders.delivery_id = delivery.id
           join payment on orders.order_transaction = payment.transaction`
	items_stmt := `select chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status from item where order_id = $1`
	rows, err := conn.Query(context.Background(), stmt)
	var order models.Order
	if err != nil {
		log.Fatal(err)
	}
	// conn.Release()
	// conn, err = d.pool.Acquire(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	for rows.Next() {
		err := rows.Scan(
			&order.Order_uid, &order.Track_number, &order.Entry,
			&order.Delivery.Name, &order.Delivery.Phone, &order.Delivery.Zip,
			&order.Delivery.City, &order.Delivery.Address, &order.Delivery.Region,
			&order.Delivery.Email,
			&order.Payment.Transaction, &order.Payment.Request_id, &order.Payment.Currency,
			&order.Payment.Provider, &order.Payment.Amount, &order.Payment.Payment_dt,
			&order.Payment.Bank, &order.Payment.Delivery_cost, &order.Payment.Goods_total,
			&order.Payment.Custom_fee,
			&order.Locale, &order.Internal_signature, &order.Customer_id,
			&order.Delivery_service, &order.Shardkey, &order.Sm_id,
			&order.Date_created, &order.Oof_shard,
		)
		if err != nil {
			log.Fatal(err)
		}
		items_rows, err := conn.Query(context.Background(), items_stmt, order.Order_uid)
		if err != nil {
			log.Fatal(err)
		}
		order.Items = make([]models.Item, 0)
		for items_rows.Next() {
			var item models.Item
			err := items_rows.Scan(
				&item.Chrt_id, &item.Track_number, &item.Price, &item.Rid, &item.Name, &item.Sale, &item.Size, &item.Total_price, &item.Nm_id, &item.Brand, &item.Status,
			)
			if err != nil {
				log.Fatal(err)
			}
			order.Items = append(order.Items, item)
		}
		orders = append(orders, order)
	}

	return orders
}

func (db *DB) insertDelivery(pool *pgxpool.Pool, d *models.Delivery) int {

	conn, err := pool.Acquire(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Release()
	insert_delivery_stmt := "insert into delivery (name, phone, zip, city, address, region, email) values ($1, $2, $3, $4, $5, $6, $7) RETURNING id"
	row := conn.QueryRow(context.Background(), insert_delivery_stmt, d.Name, d.Phone, d.Zip, d.City, d.Address, d.Region, d.Email)
	id := 0
	err = row.Scan(&id)
	if err != nil {
		log.Println(err)
	}

	return id
}

func (db *DB) insertPayment(pool *pgxpool.Pool, p *models.Payment) string {
	conn, err := pool.Acquire(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Release()
	insert_payment_stmt := "insert into payment (transaction, request_id, currency, provider, amount, payment_dt, bank, delivery_cost, goods_total, custom_fee) values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING transaction"
	row := conn.QueryRow(context.Background(), insert_payment_stmt, p.Transaction, p.Request_id, p.Currency, p.Provider, p.Amount, p.Payment_dt, p.Bank, p.Delivery_cost, p.Goods_total, p.Custom_fee)
	transaction := ""

	err = row.Scan(&transaction)
	if err != nil {
		log.Println(err)
	}
	return transaction
}

func (db *DB) bulkInsertItems(pool *pgxpool.Pool, items []models.Item, order_uid string) {
	conn, err := pool.Acquire(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Release()

	rows_items := make([][]interface{}, len(items))
	for i, item := range items {
		rows_items[i] = []interface{}{item.Chrt_id, order_uid, item.Track_number, item.Price, item.Rid, item.Name, item.Sale, item.Size, item.Total_price, item.Nm_id, item.Brand, item.Status}
	}

	copyCount, err := conn.CopyFrom(
		context.Background(),
		pgx.Identifier{"item"},
		[]string{"chrt_id", "order_id", "track_number", "price", "rid", "name", "sale", "size", "total_price", "nm_id", "brand", "status"},
		pgx.CopyFromRows(rows_items),
	)

	if err != nil {
		log.Fatal("unable to copy", err)
	}
	log.Println("Inserted", copyCount, "rows of data")
}

func (db *DB) insertOrder(pool *pgxpool.Pool, o *models.Order, delivery_id int, payment_id string) {
	conn, err := pool.Acquire(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Release()
	insert_order_stmt := `insert into 
  orders (order_uid, track_number, entry, delivery_id, order_transaction, locale, internal_signature, customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard)
  values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`
	conn.QueryRow(context.Background(), insert_order_stmt, o.Order_uid, o.Track_number, o.Entry, delivery_id, payment_id, o.Locale, o.Internal_signature, o.Customer_id, o.Delivery_service, o.Shardkey, o.Sm_id, o.Date_created, o.Oof_shard)
}
