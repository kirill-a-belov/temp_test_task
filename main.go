package main

import (
	"context"
	"log"
	"os"
	"runtime/debug"

	"playGround/cmd"
)

func main() {
	defer func() {
		if err := recover(); err != nil {
			debug.PrintStack()
			log.Fatal("panic while command execution",
				"error", err,
			)
		}
	}()

	ctx := context.Background()

	if err := cmd.New(ctx).ExecuteContext(ctx); err != nil {
		os.Exit(-1)
	}
}
