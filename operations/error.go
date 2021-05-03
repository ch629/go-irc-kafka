package operations

import (
	"github.com/ch629/go-irc-kafka/irc/parser"
	"github.com/ch629/go-irc-kafka/logging"
)

func handleErrorMessage(message parser.Message) {
	log := logging.Logger()

	// TODO: Format the message?
	log.Errorw("Received error from IRC", "message", message)
}
