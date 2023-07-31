package transport

import (
	"github.com/nats-io/stan.go"
	ffjson "github.com/pquerna/ffjson/ffjson"
	database "github.com/vlle/wb_L0/internal/database"
	models "github.com/vlle/wb_L0/internal/models/codegen_json"
	service "github.com/vlle/wb_L0/internal/services"
	"log"
	"net/http"
	"time"
)

type CacheHandler struct {
	Cache map[string]service.CacheStorage
	ch    chan service.CacheStorage
}

type AllOrdersHandler struct {
	cache_handler *CacheHandler
}

type InterfaceHandler struct {
	ch chan service.CacheStorage
}

func NewInterfaceHandler(ch chan service.CacheStorage) *InterfaceHandler {
	interfaceHandler := new(InterfaceHandler)
	interfaceHandler.ch = ch
	return interfaceHandler
}

func NewAllOrdersHandler(ch *CacheHandler) *AllOrdersHandler {
	allOrdersHandler := new(AllOrdersHandler)
	allOrdersHandler.cache_handler = ch
	return allOrdersHandler
}

func (i *InterfaceHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "internal/static/index.html")
}

func (A *AllOrdersHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	orders_uid := make([]string, 0, len(A.cache_handler.Cache))
	for _, v := range A.cache_handler.Cache {
		orders_uid = append(orders_uid, v.ReturnOrderUUID())
	}
	string_builder := "["
	for _, v := range orders_uid {
		link_to_order := "<a href=\"http://localhost:8080/data?order_uid=" + v + "\">" + v + "</a>"
		string_builder += link_to_order + "\n"
	}
	string_builder = string_builder[:len(string_builder)-1]
	string_builder += "]"
	w.Write([]byte(string_builder))
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
	log.Println("Cache listening")
	for {
		time.Sleep(1 * time.Second)
		val := <-ch
		order_uuid := val.ReturnOrderUUID()
		c.Cache[order_uuid] = val
		log.Println("Cache updated")
	}
}

func (c *CacheHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
