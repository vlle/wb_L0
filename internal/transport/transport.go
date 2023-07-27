package transport

import (
  "sync"
  "net/http"
)

type CacheHandler struct {
	Mu    sync.RWMutex
	Cache map[string][]byte
}

func (c *CacheHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c.Mu.RLock()
	defer c.Mu.RUnlock()
	w.Header().Add("Content-Type", "application/json")

	query_parameters := r.URL.Query()

	order_uid, ok := query_parameters["order_uid"]
	if !ok {
		w.Write([]byte(`{"message": "order_uid is required"}`))
		return
	}
	v, ok := c.Cache[order_uid[0]]
	if ok {
		w.Write(v)
	} else {
		w.Write([]byte(`{"message": "helloWorld"}`))
	}
}
