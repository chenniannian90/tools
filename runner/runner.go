package runner

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

var monitoredSignals = []os.Signal{
	syscall.SIGHUP,
	syscall.SIGINT,
	syscall.SIGTERM,
	syscall.SIGQUIT,
}

type Task interface {
	Run(ctx context.Context)
}

func Run(ctx context.Context, tasks ...Task) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, monitoredSignals...)

	defer func() {
		signal.Stop(quit)
		close(quit)
	}()

	ctx, cancel := context.WithCancel(ctx)

	go func() {
		<-quit
		cancel()
	}()

	var wg sync.WaitGroup

	for i := range tasks {
		task := tasks[i]

		wg.Add(1)

		go func() {
			defer wg.Done()
			task.Run(ctx)
			cancel()
		}()
	}

	wg.Wait()
}
