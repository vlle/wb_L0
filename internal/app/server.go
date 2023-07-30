package app

import (
	"github.com/vlle/wb_L0/internal/database"
	"github.com/vlle/wb_L0/internal/services"
	"github.com/vlle/wb_L0/internal/transport"

	"net/http"

	// "time"
	"fmt"
	"github.com/nats-io/stan.go"
	"sync"
)

type App struct {
  sc stan.Conn
  sbcr stan.Subscription

  db database.DB
  ch chan services.CacheStorage

  cacheHandler *transport.CacheHandler

}

func (a *App) Init() {
	a.db = services.InitDB()
	a.ch = make(chan services.CacheStorage)
	a.cacheHandler = transport.NewCacheHandler(a.ch, a.db)

	sc, err := stan.Connect("test-cluster", "321", stan.NatsURL("nats://localhost:4222"))
	if err != nil {
		fmt.Println(err.Error())
		return
	} else {
    a.sc = sc
  }


	a.sbcr, err = a.sc.Subscribe("foo", func(m *stan.Msg) {
		val, err := transport.SaveIncomingData(m, a.db)
		fmt.Println("Received from nats")
		if err != nil {
			fmt.Println(err.Error(), "error in unmarshalling")
		} else {
			fmt.Println("Sending to cache")
			a.ch <- val
		}
	})
	if err != nil {
		fmt.Println(err.Error(), "error in subscription")
		return
	}
	http.Handle("/data", a.cacheHandler)
}

func (a *App) Run() {
	var wg sync.WaitGroup
	wg.Add(2)

	defer a.sbcr.Unsubscribe()
	defer a.sbcr.Close()


	go func() {
		a.cacheHandler.Listen(a.ch)
		wg.Done()
	}()

	go func() {
		http.ListenAndServe(":8080", nil)
		wg.Done()
	}()

	wg.Wait()
}
