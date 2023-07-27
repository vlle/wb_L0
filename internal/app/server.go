package app

import (
  "net/http"
  "github.com/vlle/wb_L0/internal/transport"
)

func app() {
	cacheHandler := new(transport.CacheHandler)
	cacheHandler.Cache = make(map[string][]byte)
	cacheHandler.LoadCache()

	sc, err := stan.Connect("test-cluster", "321", stan.NatsURL("nats://localhost:4222"))
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	sc.Subscribe("foo", saveIncomingData)

	http.Handle("/data", cacheHandler)
	http.ListenAndServe(":8080", nil)
}
