package main

import (
	"fmt"
	"go-irc/config"
	"go-irc/irc"
	"go-irc/parser"
	"net"
	"os"
	"sync"
)

// https://tools.ietf.org/html/rfc1459.html

// TODO: Maybe add a rest endpoint to join/leave a channel or use a kafka topic with commands to handle from external sources
func main() {
	// TODO: Handle this WaitGroup better
	wg := sync.WaitGroup{}
	wg.Add(1)
	config.LoadConfig()

	irc.InitializeConfig()

	// Reads entire message objects created by the parser
	go irc.ReadInput()

	// Connect to IRC
	// For some reason bringing this into a method blocks everything...?
	tcpAddr, err := net.ResolveTCPAddr("tcp4", irc.BaseBotConfig.Address)
	checkError(err)
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	checkError(err)

	ircClient := irc.NewDefaultClient(conn)

	// Take output from the irc parser & send to handlers
	go func() {
		for input := range ircClient.Input() {
			parser.Output <- input
		}
	}()

	// Handle errors from irc parsing
	go func() {
		for err := range ircClient.Errors() {
			fmt.Fprintln(os.Stderr, "error from irc client,", err)
		}
	}()

	// Setup output back to IRC
	go irc.OutputStream(ircClient)

	irc.Login()

	wg.Wait()
}

// TODO: Handle errors propagated through this
func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %v", err)
		os.Exit(1)
	}
}
