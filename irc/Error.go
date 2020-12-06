package irc

import (
	"fmt"
	"go-irc/parser"
)

func handleErrorMessage(message parser.NewMessage) {
	fmt.Println("Error from IRC", message)
}
