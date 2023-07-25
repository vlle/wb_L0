package main

import (
//	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"github.com/nats-io/stan.go"

)

type cacheHandler struct {
	mu sync.Mutex // guards n
	cache map[string]string
}

func (c * cacheHandler) loadCache() {
 // connect to psql
 // retrieve all data
 // store in cache map
}

func (c *cacheHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
  c.mu.Lock()
  defer c.mu.Unlock()
  v, ok := c.cache["example"]
  w.Header().Add("Content-Type", "application/json")
  fmt.Println(v)
  if ok {
    w.Write([]byte(v))
  } else {
    // request data
    w.Write([]byte(`{"message": "helloWorld"}`))
    c.cache["example"] = `{"message": "cache stored"}`
  }
}


func main() {
  cacheHandler := new(cacheHandler)
  cacheHandler.cache = make(map[string]string)
  cacheHandler.loadCache()
  sc, err := stan.Connect("test-cluster", "321", stan.NatsURL("nats://localhost:4222"))
  if err != nil {
    fmt.Println(err.Error())
    return
  }
  go sc.Subscribe("foo", func(m *stan.Msg) {
    fmt.Printf("Received a message: %s\n", string(m.Data))
  })

  http.Handle("/data", cacheHandler)
  http.ListenAndServe(":8080", nil) 
}
