package app

import (
	"github.com/vlle/wb_L0/internal/services"
	"github.com/vlle/wb_L0/internal/transport"
	"net/http"

	// "time"
	"fmt"
	"github.com/nats-io/stan.go"
	"sync"
)

func App() {

	var wg sync.WaitGroup
	wg.Add(2)

	db := services.InitDB()
	ch := make(chan services.CacheStorage)
	cacheHandler := transport.NewCacheHandler(ch, db)

	sc, err := stan.Connect("test-cluster", "321", stan.NatsURL("nats://localhost:4222"))
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	sbscr, err := sc.Subscribe("foo", func(m *stan.Msg) {
		val, err := transport.SaveIncomingData(m, db)
		fmt.Println("Received from nats")
		if err != nil {
			fmt.Println(err.Error(), "error in unmarshalling")
		} else {
			fmt.Println("Sending to cache")
			ch <- val
		}
	})
	if err != nil {
		fmt.Println(err.Error(), "error in subscription")
		return
	}
	defer sbscr.Unsubscribe()
	defer sbscr.Close()

	http.Handle("/data", cacheHandler)

	go func() {
		cacheHandler.Listen(ch)
		wg.Done()
	}()

	go func() {
		http.ListenAndServe(":8080", nil)
		wg.Done()
	}()

	wg.Wait()
}
