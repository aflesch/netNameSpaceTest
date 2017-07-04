package main

import (
	"time"

	"fmt"
	"os"

	"github.com/vishvananda/netns"
)

func main() {
	// Create a new network namespace
	newns, err := netns.New()
	netns.Set(newns)
	defer newns.Close()
	fmt.Printf("Start new process PID=%d with network namespace %d (err=%v)\n", os.Getpid(), newns.String(), err)
	// run loop forever (or until ctrl-c)
	for {
		time.Sleep(10000000)
	}
}
