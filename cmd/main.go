package main

import (
  // "github.com/spf13/cobra"
  "github.com/vlle/wb_L0/internal/app"
)


func main() {
  var application app.App

  application.Init()
  application.Run()
}
