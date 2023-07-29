package transport

import (
  "sync"
  "time"
  "net/http"
	service "github.com/vlle/wb_L0/internal/services"
  "github.com/nats-io/stan.go"
  "fmt"
  "encoding/json"
  models "github.com/vlle/wb_L0/internal/models"
  database "github.com/vlle/wb_L0/internal/database"
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
	err := json.Unmarshal(m.Data, &js)
  if err != nil {
    fmt.Println(err.Error(), "error in unmarshalling")
    return service.CacheStorage{}, err
  }
  var wg sync.WaitGroup
  wg.Add(1)
  go func() {
    service.SaveIncomingOrder(js, db)
    wg.Done()
  }()
  wg.Wait()
  val := service.NewCacheFromData(js)
  return val, nil
}

func (c *CacheHandler) Listen(ch chan service.CacheStorage) {
  for {
    time.Sleep(1 * time.Second)
    fmt.Println("Cache listening (tired)")
    val := <- ch
    fmt.Println("Cache received")
    order_uuid := val.ReturnOrderUUID()
    c.Cache[order_uuid] = val
    fmt.Println("Cache updated")
    fmt.Println(c.Cache)
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
  fmt.Println(order_uid[0])
  v, ok := c.Cache[order_uid[0]]
  fmt.Println(c.Cache)
  if ok {
    order_data := v.ReturnJson()
    if order_data == nil {
      w.WriteHeader(http.StatusInternalServerError)
      w.Write([]byte(`{"message": "problem with encoding json"}`))
    } else {
      w.Write(v.ReturnJson())
    }
  } else {
    w.WriteHeader(http.StatusNotFound)
    w.Write([]byte(`{"message": "order_uid not found"}`))
  }
}
