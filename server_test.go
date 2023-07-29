package main

import (
    "net/http"
    "net/http/httptest"
    "strings"
    "testing"
)

func TestHandleHelloWorld(t *testing.T) {
    wr := httptest.NewRecorder()
    // req := httptest.NewRequest(http.MethodGet, "/sloth", nil)

    // handleHelloWorld(wr, req)
    if wr.Code != http.StatusOK {
        t.Errorf("got HTTP status code %d, expected 200", wr.Code)
    }

    if !strings.Contains(wr.Body.String(), "helloWorld") {
        t.Errorf(
            `response body "%s" does not contain "Stay slothful!"`,
            wr.Body.String(),
        )
    }
  }

