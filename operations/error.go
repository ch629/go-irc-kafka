package operations

import (
	"github.com/ch629/go-irc-kafka/irc/parser"
)

func handleErrorMessage(handler OperationHandler, message parser.Message) {
	// TODO: Format the message?
	handler.log.Errorw("Received error from IRC", "message", message)
}
