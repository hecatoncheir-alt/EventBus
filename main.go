package main

import (
	"fmt"
	"github.com/hecatoncheir/Configuration"
	"github.com/hecatoncheir/EventBus/engine"
	"os"
)

func main() {
	config := configuration.New()
	socket := engine.New(config.APIVersion)
	err := socket.SetUp(config.Production.EventBus.Host, config.Production.EventBus.Port)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}

	socket.Listen()
}
