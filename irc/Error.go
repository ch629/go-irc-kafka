package irc

import (
	"fmt"
	"go-irc/parser"
)

func handleErrorMessage(message parser.Message) {
	fmt.Println("Error from IRC", message)
}
