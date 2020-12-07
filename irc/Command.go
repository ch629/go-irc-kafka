package irc

import (
	"encoding/json"
	"fmt"
	"go-irc/parser"
)

var commandMap = map[string]func(message parser.Message){
	"PING":    handlePing,
	"ERROR":   handleErrorMessage,
	"PRIVMSG": handleMessage,
	"375": func(message parser.Message) {
		// RPL_MOTDSTART
	},
	"372": handleMotd,
	"376": handleMotdEnd,
	"353": func(message parser.Message) {
		// RPL_NAMREPLY
		//  <channel> :[[@|+]<nick> [[@|+]<nick> [...]]]
		fmt.Println("Got users: ", message.Params)
	},
	"366": func(message parser.Message) {
		// RPL_ENDOFNAMES
		// <channel> :End of /NAMES list
		fmt.Println("End of names list")
	},
	"JOIN": func(message parser.Message) {
		// Joined channel
		fmt.Println("Joined channel: ", message.Params)
	},
	"421": func(message parser.Message) {
		fmt.Println("Invalid command: ", message.Params)
	},
}

func ReadInput() {
	for message := range parser.Output {
		if f, ok := commandMap[message.Command]; ok {
			f(message)
		} else {
			// Print out message if not known
			bytes, _ := json.Marshal(message)
			fmt.Printf("Message: %v\n", string(bytes))
		}
	}
}
