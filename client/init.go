package client

import (
	"os"
	"os/signal"
)

var Interrupt = make(chan os.Signal, 1)

func init() {
	signal.Notify(Interrupt, os.Interrupt)
}
