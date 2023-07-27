package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/stan.go"
	models "github.com/vlle/wb_L0/models"
)

func main() {
	cacheHandler := new(cacheHandler)
	cacheHandler.cache = make(map[string][]byte)
	cacheHandler.loadCache()

	sc, err := stan.Connect("test-cluster", "321", stan.NatsURL("nats://localhost:4222"))
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	sc.Subscribe("foo", saveIncomingData)

	http.Handle("/data", cacheHandler)
	http.ListenAndServe(":8080", nil)
}

type cacheHandler struct {
	mu    sync.RWMutex
	cache map[string][]byte
}

func (c *cacheHandler) loadCache() {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		url = "postgres://postgres:postgres@localhost:5500/rec"
	}
	dbpool, err := pgxpool.New(context.Background(), url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to create connection pool: %v\n", err)
		os.Exit(1)
	}
	defer dbpool.Close()

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
	items_stmt := `select chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status from item where chrt_id = $1`
	rows, err := dbpool.Query(context.Background(), stmt)
	var order models.Order
	if err != nil {
		fmt.Fprintf(os.Stderr, "QueryRow failed: %v\n", err)
		os.Exit(1)
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
			fmt.Fprintf(os.Stderr, "QueryRow scan failed: %v\n", err)
			os.Exit(1)
		}
		items_rows, err := dbpool.Query(context.Background(), items_stmt, order.Order_uid)
		if err != nil {
			fmt.Fprintf(os.Stderr, "QueryRow scan failed: %v\n", err)
			os.Exit(1)
		}
		for items_rows.Next() {
			var item models.Item
			err := items_rows.Scan(
				&item.Chrt_id, &item.Track_number, &item.Price, &item.Rid, &item.Name, &item.Sale, &item.Size, &item.Total_price, &item.Nm_id, &item.Brand, &item.Status,
			)
			if err != nil {
				fmt.Fprintf(os.Stderr, "QueryRow scan failed: %v\n", err)
				os.Exit(1)
			}
			order.Items = append(order.Items, item)
		}
		c.cache[order.Order_uid], err = json.Marshal(order)
		if err != nil {
			fmt.Fprintf(os.Stderr, "QueryRow marshal failed: %v\n", err)
			os.Exit(1)
		}
	}
}

func (c *cacheHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	w.Header().Add("Content-Type", "application/json")

	query_parameters := r.URL.Query()

	order_uid, ok := query_parameters["order_uid"]
	if !ok {
		w.Write([]byte(`{"message": "order_uid is required"}`))
		return
	}
	v, ok := c.cache[order_uid[0]]
	if ok {
		w.Write(v)
	} else {
		w.Write([]byte(`{"message": "helloWorld"}`))
	}
}

func saveIncomingData(m *stan.Msg) {
	var js models.Order
	err := json.Unmarshal(m.Data, &js)
	fmt.Println(js)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		url = "postgres://postgres:postgres@localhost:5500/rec"
	}
	dbpool, err := pgxpool.New(context.Background(), url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to create connection pool: %v\n", err)
		os.Exit(1)
	}
	defer dbpool.Close()
	delivery_id := insertDelivery(dbpool, &js.Delivery)
	payment_transaction := insertPayment(dbpool, &js.Payment)
	insertOrder(dbpool, &js, delivery_id, payment_transaction)
	bulkInsertItems(dbpool, js.Items, js.Order_uid)
}

func insertDelivery(dbpool *pgxpool.Pool, d *models.Delivery) int {

	insert_delivery_stmt := "insert into delivery (name, phone, zip, city, address, region, email) values ($1, $2, $3, $4, $5, $6, $7) RETURNING id"
	row := dbpool.QueryRow(context.Background(), insert_delivery_stmt, d.Name, d.Phone, d.Zip, d.City, d.Address, d.Region, d.Email)
	id := 0
	err := row.Scan(&id)
	if err != nil {
		fmt.Println(err.Error(), "delivery")
	}

	return id
}

func insertPayment(dbpool *pgxpool.Pool, p *models.Payment) string {
	insert_payment_stmt := "insert into payment (transaction, request_id, currency, provider, amount, payment_dt, bank, delivery_cost, goods_total, custom_fee) values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING transaction"
	row := dbpool.QueryRow(context.Background(), insert_payment_stmt, p.Transaction, p.Request_id, p.Currency, p.Provider, p.Amount, p.Payment_dt, p.Bank, p.Delivery_cost, p.Goods_total, p.Custom_fee)
	transaction := ""

	err := row.Scan(&transaction)
	if err != nil {
		fmt.Println(err.Error(), "payment")
	}
  fmt.Println(transaction)
	return transaction
}

func bulkInsertItems(dbpool *pgxpool.Pool, items []models.Item, order_uid string) {
	insert_items_stmt := "insert into item values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)"
	for _, item := range items {
		row := dbpool.QueryRow(context.Background(), insert_items_stmt, item.Chrt_id, order_uid, item.Track_number, item.Price, item.Rid, item.Name, item.Sale, item.Size, item.Total_price, item.Nm_id, item.Brand, item.Status)

		err := row.Scan()
		if err != nil {
			fmt.Println(err.Error(), "item")
		}
	}
}

func insertOrder(dbpool *pgxpool.Pool, o *models.Order, delivery_id int, payment_id string)  {
	insert_order_stmt := `insert into 
                        orders (order_uid, track_number, entry, delivery_id, order_transaction, locale, internal_signature, customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard)
                        values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`
  fmt.Println(payment_id)
	dbpool.QueryRow(context.Background(), insert_order_stmt, o.Order_uid, o.Track_number, o.Entry, delivery_id, payment_id, o.Locale, o.Internal_signature, o.Customer_id, o.Delivery_service, o.Shardkey, o.Sm_id, o.Date_created, o.Oof_shard)
}
