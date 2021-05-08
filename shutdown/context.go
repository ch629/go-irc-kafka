package shutdown

import (
	"context"
	"github.com/ch629/go-irc-kafka/logging"
	"os"
	"os/signal"
)

func InterruptAwareContext(ctx context.Context) context.Context {
	signals := make(chan os.Signal)
	signal.Notify(signals, os.Interrupt)
	newCtx, cancelFunc := context.WithCancel(ctx)
	go func() {
		defer close(signals)
		select {
		case <-signals:
			log := logging.Logger()
			log.Info("Received interrupt")
			cancelFunc()
		case <-ctx.Done():
		}
	}()
	return newCtx
}
