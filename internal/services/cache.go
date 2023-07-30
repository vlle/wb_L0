package services

import (
	"encoding/json"
	"fmt"
	"github.com/vlle/wb_L0/internal/database"
	models "github.com/vlle/wb_L0/internal/models"
	// import "github.com/go-playground/validator/v10"
)

type CacheStorage struct {
	unencoded_data models.Order
	json_data      []byte
	is_json_ready  bool
}

func (c *CacheStorage) SaveUnencoded(unencoded_data models.Order) {
	c.unencoded_data = unencoded_data
	c.is_json_ready = false
}

func (c *CacheStorage) SaveEncoded(json_data []byte) {
	c.json_data = json_data
	c.is_json_ready = true
}

func (c *CacheStorage) ReturnOrderUUID() string {
	return c.unencoded_data.Order_uid
}

func (c *CacheStorage) ReturnJson() []byte {
	if c.is_json_ready {
		return c.json_data
	} else {
		json_data, err := json.Marshal(c.unencoded_data)
		if err != nil {
			fmt.Println(err.Error())
			return nil
		}
		c.json_data = json_data
		c.is_json_ready = true
		return c.json_data
	}
}

func NewCache() CacheStorage {
	var cache CacheStorage
	cache.SaveUnencoded(models.Order{})
	return cache
}

func NewCacheFromData(unencoded_data models.Order) CacheStorage {
	var cache CacheStorage
	cache.SaveUnencoded(unencoded_data)
	return cache
}

func NewCacheFromJson(json_data []byte) CacheStorage {
	var cache CacheStorage
	cache.SaveEncoded(json_data)
	return cache
}

func LoadCacheFromDB(db database.DB) map[string]CacheStorage {
	cache := make(map[string]CacheStorage)
	orders := db.LoadOrders()
	for _, order := range orders {
		cache[order.Order_uid] = NewCacheFromData(order)
	}
	return cache
}

func SaveIncomingOrder(model models.Order, db database.DB) {
	db.SaveOrder(model)
}

func InitDB() database.DB {
	return database.InitDB()
}
