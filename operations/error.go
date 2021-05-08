package operations

import (
	"github.com/ch629/go-irc-kafka/irc/parser"
	"go.uber.org/zap"
)

func handleErrorMessage(handler OperationHandler, message parser.Message) {
	handler.log.Error("Received error from IRC", zap.Any("message", message))
}
