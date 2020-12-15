package operations

import (
	"fmt"
	"go-irc/irc/parser"
	"os"
)

func handleErrorMessage(message parser.Message) {
	// TODO: Format the message?
	fmt.Fprintf(os.Stderr, "Received error from IRC: %v\n", message)
}
