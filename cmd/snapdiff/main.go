package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	cmd := newRootCmd()
	if err := cmd.ExecuteContext(ctx); err != nil {
		log.Fatal(err)
	}
}
