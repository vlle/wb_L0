package database

import (
  "context"
  // import "github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"
)

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
