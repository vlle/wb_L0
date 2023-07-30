// Copyright 2016-2019 The NATS Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	nats "github.com/nats-io/nats.go"
	"github.com/nats-io/stan.go"
)

var usageStr = `
Usage: stan-pub [options] <subject> <message>

Options:
	-s,  --server   <url>            NATS Streaming server URL(s)
	-c,  --cluster  <cluster name>   NATS Streaming cluster name
	-id, --clientid <client ID>      NATS Streaming client ID
	-a,  --async                     Asynchronous publish mode
	-cr, --creds    <credentials>    NATS 2.0 Credentials
`

const (
)

// NOTE: Use tls scheme for TLS, e.g. stan-pub -s tls://demo.nats.io:4443 foo hello
func usage() {
	fmt.Printf("%s\n", usageStr)
	os.Exit(0)
}

func publish(filepath []byte, subj string, sc stan.Conn) {
  sc.Publish(subj, filepath)
}


func main() {
  v, _ := os.ReadFile("json_mock/model5.json")
  j := string(v)
  formatted_slice := make([][]byte, 0)

  rand.Seed(time.Now().UnixNano())
  for i := 0; i < 5; i++ { 
    v := rand.Intn(100000 - 1451) + 3123
    formatted := fmt.Sprintf(j, v, v)
    fmt.Println(formatted)
    formatted_slice = append(formatted_slice, []byte(formatted))
  }
  fmt.Println(formatted_slice)

	var (
		clusterID string
		clientID  string
		URL       string
		async     bool
		userCreds string
	)

	flag.StringVar(&URL, "s", stan.DefaultNatsURL, "The nats server URLs (separated by comma)")
	flag.StringVar(&URL, "server", stan.DefaultNatsURL, "The nats server URLs (separated by comma)")
	flag.StringVar(&clusterID, "c", "test-cluster", "The NATS Streaming cluster ID")
	flag.StringVar(&clusterID, "cluster", "test-cluster", "The NATS Streaming cluster ID")
	flag.StringVar(&clientID, "id", "stan-pub", "The NATS Streaming client ID to connect with")
	flag.StringVar(&clientID, "clientid", "stan-pub", "The NATS Streaming client ID to connect with")
	flag.BoolVar(&async, "a", false, "Publish asynchronously")
	flag.BoolVar(&async, "async", false, "Publish asynchronously")
	flag.StringVar(&userCreds, "cr", "", "Credentials File")
	flag.StringVar(&userCreds, "creds", "", "Credentials File")

	log.SetFlags(0)
	flag.Usage = usage
	flag.Parse()

	args := flag.Args()

	if len(args) < 1 {
		usage()
	}

	// Connect Options.
	opts := []nats.Option{nats.Name("NATS Streaming Example Publisher")}
	// Use UserCredentials
	if userCreds != "" {
		opts = append(opts, nats.UserCredentials(userCreds))
	}

	// Connect to NATS
	nc, err := nats.Connect(URL, opts...)
	if err != nil {
		log.Fatal(err)
	}
	defer nc.Close()

	sc, err := stan.Connect(clusterID, clientID, stan.NatsConn(nc))
	if err != nil {
		log.Fatalf("Can't connect: %v.\nMake sure a NATS Streaming Server is running at: %s", err, URL)
	}
	defer sc.Close()

  subj := args[0]
  for i := 0; i < 5; i++ {
    go publish(formatted_slice[i], subj, sc)
  }
  time.Sleep(5 * time.Second)
}

// type Delivery struct {
// 	Name    string `json:"name"`
// 	Phone   string `json:"phone"`
// 	Zip     string `json:"zip"`
// 	City    string `json:"city"`
// 	Address string `json:"address"`
// 	Region  string `json:"region"`
// 	Email   string `json:"email"`
// }
// 
// type Payment struct {
// 	Transaction   string `json:"transaction"`
// 	Request_id    string `json:"request_id"`
// 	Currency      string `json:"currency"`
// 	Provider      string `json:"provider"`
// 	Amount        int    `json:"amount"`
// 	Payment_dt    int64  `json:"payment_dt"`
// 	Bank          string `json:"bank"`
// 	Delivery_cost int    `json:"delivery_cost"`
// 	Goods_total   int    `json:"goods_total"`
// 	Custom_fee    int    `json:"custom_fee"`
// }
// 
// type Item struct {
// 	Chrt_id      int    `json:"chrt_id"`
// 	Track_number string `json:"track_number"`
// 	Price        int    `json:"price"`
// 	Rid          string `json:"rid"`
// 	Name         string `json:"name"`
// 	Sale         int    `json:"sale"`
// 	Size         string `json:"size"`
// 	Total_price  int    `json:"total_price"`
// 	Nm_id        int    `json:"nm_id"`
// 	Brand        string `json:"brand"`
// 	Status       int    `json:"status"`
// }

// Order scheme

//	Order_uid    string `json:"order_uid"`
//	Track_number string `json:"track_number"`
//	Entry        string `json:"entry"`
//
//	Delivery Delivery `json:"delivery"`
//	Payment  Payment  `json:"payment"`
//	Items    []Item   `json:"items"`
//
//	Locale             string    `json:"locale"`
//	Internal_signature string    `json:"internal_signature"`
//	Customer_id        string    `json:"customer_id"`
//	Delivery_service   string    `json:"delivery_service"`
//	Shardkey           string    `json:"shardkey"`
//	Sm_id              int       `json:"sm_id"`
//	Date_created       time.Time `json:"date_created"`
//	Oof_shard          string    `json:"oof_shard"`
