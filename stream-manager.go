package main

import (
	"github.com/lietu/stream-manager/manager"
	_ "github.com/lietu/stream-manager/streamservice"
)

func main() {
	manager := manager.NewManager()
	manager.Start()
}
