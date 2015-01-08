package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {

	fmt.Println("Hi, my pid is", os.Getpid())

	trap := make(chan os.Signal, 1)
	signal.Notify(trap, os.Interrupt, os.Kill, syscall.SIGTERM)
	go func() {
		sig := <-trap
		fmt.Println("Caught signal", sig, "in my signal handler")
		os.Exit(0)
	}()

	for i := 0; ; i++ {
		fmt.Println(i)
		time.Sleep(1 * time.Second)
	}
}
