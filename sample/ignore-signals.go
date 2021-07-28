package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	sigs := []os.Signal{syscall.SIGINT, syscall.SIGTERM}
	signal.Ignore(sigs...)
	for {
		fmt.Printf("I ignore signals: %v\n", sigs)
		time.Sleep(time.Second)
	}
}
