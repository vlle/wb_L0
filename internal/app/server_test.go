package app

import (
  "net/http"
  "fmt"
  "time"
  "net/url"
  "testing"
)

const serverPort = 8080

func TestHandleGetQueryNotFound(t *testing.T) {

  var app App
  go func() {
    app.Init()
    app.Run()
  }()

  time.Sleep(3 * time.Second)
  uri := "/data?"
  unlno := "111444" // random not found

  param := make(url.Values)
  param["order_uid"] = []string{unlno}
  requestURL := fmt.Sprintf("http://localhost:%d", serverPort)
  resp, err := http.Get(requestURL + uri + param.Encode())
  if err != nil {
    t.Errorf("got error %s, expected nil", err)
  }
  if resp.StatusCode != http.StatusNotFound {
    t.Errorf("got HTTP status code %d, expected 404", resp.StatusCode) 
  }
}
