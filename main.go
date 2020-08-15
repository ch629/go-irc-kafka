package main

import (
	"bufio"
	"fmt"
	"go-irc/config"
	"go-irc/irc"
	"go-irc/kafka"
	"go-irc/parser"
	"net"
	"os"
	"sync"
)

// https://tools.ietf.org/html/rfc1459.html

// TODO: Maybe add a rest endpoint to join/leave a channel or use a kafka topic with commands to handle from external sources
var (
	conn *net.TCPConn
)

// TODO: Check if we can have a separate IRC connection per channel? Or batch them, then we can use more coroutines for parsing etc
// TODO: Check if the Kafka connection is active before we start parsing messages from IRC, would break the initial startup
// TODO: Pull some of this logic into the irc package & rework it slightly so there's less needed in here
func main() {
	// TODO: Handle this WaitGroup better
	wg := sync.WaitGroup{}
	wg.Add(1)
	config.LoadConfig()

	irc.InitializeConfig()

	// Kafka producer output
	kafka.BatchPoll()

	// Reads entire message objects created by the parser
	go irc.ReadInput()

	// Connect to IRC
	// For some reason bringing this into a method blocks everything...?
	tcpAddr, err := net.ResolveTCPAddr("tcp4", irc.BaseBotConfig.Address)
	checkError(err)
	conn, err = net.DialTCP("tcp", nil, tcpAddr)
	checkError(err)

	// Write data from IRC connection to the parser
	go func() {
		reader := bufio.NewReader(conn)
		_, err = reader.WriteTo(parser.MakeWriter())
		checkError(err)
	}()

	// Setup output back to IRC
	go irc.OutputStream(conn)

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
