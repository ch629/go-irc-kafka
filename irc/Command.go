package irc

import (
	"encoding/json"
	"fmt"
	"go-irc/parser"
)

var commandMap = map[string]func(message parser.NewMessage){
	"PING":    handlePing,
	"ERROR":   handleErrorMessage,
	"PRIVMSG": handleMessage,
	"375": func(message parser.NewMessage) {
		// RPL_MOTDSTART
	},
	"372": handleMotd,
	"376": handleMotdEnd,
	"353": func(message parser.NewMessage) {
		// RPL_NAMREPLY
		//  <channel> :[[@|+]<nick> [[@|+]<nick> [...]]]
		fmt.Println("Got users: ", message.Args)
	},
	"366": func(message parser.NewMessage) {
		// RPL_ENDOFNAMES
		// <channel> :End of /NAMES list
		fmt.Println("End of names list")
	},
	"JOIN": func(message parser.NewMessage) {
		// Joined channel
		fmt.Println("Joined channel: ", message.Args)
	},
	"421": func(message parser.NewMessage) {
		fmt.Println("Invalid command: ", message.Args)
	},
}

func ReadInput() {
	for message := range parser.Output {
		if f, ok := commandMap[message.Command]; ok {
			f(message)
		} else {
			// TODO: Figure out how we can parse this better -> It thinks before the : is metadata & pulling in user-type from the previous message sent
			// PING :tmi.twitch.tv
			if message.Metadata["user-type"] == "PING" {
				handlePing(message)
				return
			}
			// Print out message if not known
			bytes, _ := json.Marshal(message)
			fmt.Printf("Message: %v\n", string(bytes))
		}
	}
}
