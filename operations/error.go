package operations

import (
	"go-irc/irc/parser"
	"go-irc/logging"
)

func handleErrorMessage(message parser.Message) {
	log := logging.Logger()

	// TODO: Format the message?
	log.Errorw("Received error from IRC", "message", message)
}
