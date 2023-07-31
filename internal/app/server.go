package app

import (
	"github.com/vlle/wb_L0/internal/database"
	"github.com/vlle/wb_L0/internal/services"
	"github.com/vlle/wb_L0/internal/transport"

	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/nats-io/stan.go"
)

type App struct {
	sc   stan.Conn
	sbcr stan.Subscription

	db database.DB
	ch chan services.CacheStorage

	cacheHandler     *transport.CacheHandler
	interfaceHandler *transport.InterfaceHandler
	allOrdersHandler *transport.AllOrdersHandler
}

func (a *App) Init() {
	a.db = services.InitDB()
	a.ch = make(chan services.CacheStorage)
	a.cacheHandler = transport.NewCacheHandler(a.ch, a.db)
	a.interfaceHandler = transport.NewInterfaceHandler(a.ch)
	a.allOrdersHandler = transport.NewAllOrdersHandler(a.cacheHandler)

	sc, err := stan.Connect("test-cluster", "321", stan.NatsURL("nats://localhost:4222"))
	if err != nil {
		fmt.Println(err.Error())
		return
	} else {
		a.sc = sc
	}

	a.sbcr, err = a.sc.Subscribe("foo", func(m *stan.Msg) {
		fmt.Println("Received from nats")
		mapD := make(map[string]interface{})
		err := json.Unmarshal(m.Data, &mapD)
		if err != nil {
			fmt.Println(err.Error(), "error in unmarshalling")
			return
		}
		fmt.Println(mapD)
		go func() {
			val, err := transport.SaveIncomingData(m, a.db)
			fmt.Println("Received from nats")
			if err != nil {
				fmt.Println(err.Error(), "error in unmarshalling")
			} else {
				fmt.Println("Sending to cache")
				a.ch <- val
			}
		}()
	})
	if err != nil {
		fmt.Println(err.Error(), "error in subscription")
		return
	}
	http.Handle("/data", a.cacheHandler)
	http.Handle("/data/all", a.allOrdersHandler)
	http.Handle("/interface", a.interfaceHandler)
}

func (a *App) Run() {

	defer a.sbcr.Unsubscribe()
	defer a.sbcr.Close()
	defer a.db.Close()

	go func() {
		a.cacheHandler.Listen(a.ch)
	}()

	go func() {
		http.ListenAndServe(":8080", nil)
	}()

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	<-done

	log.Print("Server Stopped")
	log.Print("Server Exited Properly")
	os.Exit(0)
}
