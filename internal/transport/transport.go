package transport

import (
  "log"
	"github.com/nats-io/stan.go"
	database "github.com/vlle/wb_L0/internal/database"
	models "github.com/vlle/wb_L0/internal/models/codegen_json"
	service "github.com/vlle/wb_L0/internal/services"
  ffjson "github.com/pquerna/ffjson/ffjson"
	"net/http"
	"sync"
	"time"
)

type CacheHandler struct {
	Mu    sync.RWMutex
	Cache map[string]service.CacheStorage
	ch    chan service.CacheStorage
}

func NewCacheHandler(ch chan service.CacheStorage, dbpool database.DB) *CacheHandler {
	cacheHandler := new(CacheHandler)
	cacheHandler.Cache = service.LoadCacheFromDB(dbpool)
	cacheHandler.ch = ch
	return cacheHandler
}

type MessageHandler struct {
	ch chan service.CacheStorage
}

func SaveIncomingData(m *stan.Msg, db database.DB) (service.CacheStorage, error) {
	var js models.Order
	err := ffjson.Unmarshal(m.Data, &js)
	if err != nil {
    log.Println(err.Error(), "error in unmarshalling")
		return service.CacheStorage{}, err
  }
  service.SaveIncomingOrder(js, db)
  val := service.NewCacheFromData(js)
  return val, nil
}

// Saves incoming order data to cache
func (c *CacheHandler) Listen(ch chan service.CacheStorage) {
  for {
    time.Sleep(1 * time.Second)
    log.Println("Cache listening")
    val := <-ch
    log.Println("Cache received")
    order_uuid := val.ReturnOrderUUID()
    c.Cache[order_uuid] = val
    log.Println("Cache updated")
  }
}

func (c *CacheHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
  c.Mu.RLock()
  defer c.Mu.RUnlock()
  w.Header().Add("Content-Type", "application/json")

  query_parameters := r.URL.Query()

  order_uid, ok := query_parameters["order_uid"]
  if !ok {
    w.WriteHeader(http.StatusBadRequest)
    w.Write([]byte(`{"message": "order_uid is required"}`))
    return
  }
  v, ok := c.Cache[order_uid[0]]
  if ok {
    order_data := v.ReturnJson()
    if order_data == nil {
      w.WriteHeader(http.StatusInternalServerError)
      w.Write([]byte(`{"message": "problem with encoding json"}`))
    } else {
      w.Write(order_data)
    }
  } else {
    w.WriteHeader(http.StatusNotFound)
    w.Write([]byte(`{"message": "order_uid not found"}`))
  }
}
