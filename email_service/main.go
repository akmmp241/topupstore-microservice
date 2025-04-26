package main

import (
	"log/slog"
	"os"
	"sync"
)

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})))

	app := NewAppServer()

	var wg sync.WaitGroup

	app.RunConsumer(&wg)

	wg.Wait()
}
