package irc

import (
	"fmt"
	"go-irc/parser"
)

func handleErrorMessage(message parser.OldMessage) {
	fmt.Println("Error from IRC", message)
}
