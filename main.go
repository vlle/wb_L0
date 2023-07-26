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
	query_parameters := r.URL.Query()

  order_uid, ok := query_parameters["order_uid"]
  if !ok {
    w.Write([]byte(`{"message": "order_uid is required"}`))
    return
  }
	v, ok := c.cache[order_uid[0]]
	w.Header().Add("Content-Type", "application/json")
	if ok {
		w.Write(v)
	} else {
		w.Write([]byte(`{"message": "helloWorld"}`))
	}
}

type J struct {
	X map[string]interface{} `json:"-"`
}

func saveIncomingData(m *stan.Msg) {
	var js models.Order
  err := json.Unmarshal(m.Data, &js)
  if err != nil {
    fmt.Println(err.Error())
    return
  }
  // save this stuff by passing by pointer etc
}

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
