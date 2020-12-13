package irc

import (
	"encoding/json"
	"fmt"
	"go-irc/parser"
	"net"
)

type (
	// TODO: Close (when closed, input/output need stopping)
	IrcClient interface {
		Input() <-chan parser.Message
		Output() chan<- IrcMessage
		Errors() <-chan error
	}

	IrcMessage interface {
		Bytes() []byte
	}

	client struct {
		conn       *net.TCPConn
		inputChan  chan parser.Message
		outputChan chan IrcMessage
		errorChan  chan error
	}
)

var crlfBytes = []byte{'\r', '\n'}

func NewDefaultClient(conn *net.TCPConn) IrcClient {
	// TODO: Buffer on the channels?
	errorChan := make(chan error)
	cli := &client{
		conn:       conn,
		inputChan:  make(chan parser.Message),
		outputChan: make(chan IrcMessage),
		errorChan:  errorChan,
	}

	cli.setupOutput()
	cli.readInput()

	return cli
}

// TODO: make a new message type for output instead of []byte?
// Output back to the IRC connection
func (cli *client) Output() chan<- IrcMessage {
	return cli.outputChan
}

// TODO: Should this be <-chan *parser.Message?
// Input from the IRC connection
func (cli *client) Input() <-chan parser.Message {
	return cli.inputChan
}

// Errors from IRC input/output
func (cli *client) Errors() <-chan error {
	return cli.errorChan
}

func (cli *client) readInput() {
	scanner := parser.NewScanner(cli.conn)

	go func() {
		for {
			message, err := scanner.Scan()

			if err != nil {
				// TODO: Make this non-blocking
				cli.errorChan <- err
				continue
			}

			logMessage(message)

			cli.inputChan <- *message
		}
	}()
}

// For testing
func logMessage(msg *parser.Message) {
	bytes, _ := json.MarshalIndent(msg, "", "  ")
	fmt.Println(">", string(bytes))
}

func (cli *client) setupOutput() {
	go func() {
		for output := range cli.outputChan {
			bytes := output.Bytes()
			fmt.Println("< ", string(bytes))
			// Message
			if _, err := cli.conn.Write(bytes); err != nil {
				cli.errorChan <- err
				continue
			}

			// CRLF
			if _, err := cli.conn.Write(crlfBytes); err != nil {
				cli.errorChan <- err
			}
		}
	}()
}
