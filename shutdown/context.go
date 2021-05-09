package shutdown

import (
	"context"
	"github.com/ch629/go-irc-kafka/logging"
	"go.uber.org/zap"
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
		case s := <-signals:
			logging.Logger().Info("Cancelling context due to", zap.Stringer("signal", s))
			cancelFunc()
		case <-ctx.Done():
		}
	}()
	return newCtx
}
