package server_types

import "time"

type delivery struct {
  Name string
  Phone string
  Zip string
  City string
  Address string
  Region string
  Email string
}


type payment struct {
  Transaction string
  Request_id string
  Currency string
  Provider string
  Amount int
  Payment_dt int64
  Bank string
  Delivery_cost int
  Goods_total int
  Custom_fee int
}


type Item struct {
  Chrt_id int
  Track_number string
  Price int
  Rid string
  Name string
  Sale int
  Size string
  Total_price int
  Nm_id int
  Brand string
  Status int
}

type Order struct {
  Order_uid string `json:"order_uid"`
  Track_number string `json:"track_number"`
  Entry string  `json:"entry"`

  Delivery delivery `json:"delivery"`
  Payment payment `json:"payment"`
  Items []Item `json:"items"`
  

  Locale string `json:"locale"`
  Internal_signature string `json:"internal_signature"`
  Customer_id string `json:"customer_id"`
  Delivery_service string `json:"delivery_service"`
  Shardkey string  `json:"shardkey"`
  Sm_id int `json:"sm_id"`
  Date_created time.Time `json:"date_created"`
  Oof_shard string `json:"oof_shard"`
}
