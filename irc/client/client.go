package client

import (
	"context"
	"errors"
	"github.com/ch629/go-irc-kafka/irc/parser"
	"io"
	"sync"
)

type (
	// TODO: Store somewhere why it was closed so we can attempt to reconnect if it was just disconnected
	IrcClient interface {
		io.Closer
		// Input is a channel of messages coming from IRC
		Input() <-chan parser.Message
		// Output is a channel of messages to send to IRC
		// TODO: Replace output with a send func which will just send on the channel?
		Output() chan<- IrcMessage
		// Errors is a channel of errors generated when reading or writing to IRC
		Errors() <-chan error
		// Closed is whether the client connection is closed
		Closed() bool
		// Done is a channel to notify consumers when the client is done closing
		Done() <-chan struct{}
		Err() error
	}

	IrcMessage interface {
		Bytes() []byte
	}

	client struct {
		ctx        context.Context
		cancelFunc context.CancelFunc
		conn       io.ReadWriteCloser
		inputChan  chan parser.Message
		outputChan chan IrcMessage
		errorChan  chan error
		scanner    parser.Scanner
		wg         sync.WaitGroup
		done       chan struct{}
		err        error
	}
)

func NewClient(ctx context.Context, conn io.ReadWriteCloser) IrcClient {
	cli := &client{
		conn:       conn,
		inputChan:  make(chan parser.Message),
		outputChan: make(chan IrcMessage),
		errorChan:  make(chan error),
		done:       make(chan struct{}),
		scanner:    *parser.NewScanner(conn),
	}
	cli.ctx, cli.cancelFunc = context.WithCancel(ctx)

	cli.wg.Add(2)
	go cli.cleanup()
	go cli.setupOutput()
	go cli.readInput()

	return cli
}

func (cli *client) Done() <-chan struct{} {
	return cli.done
}

func (cli *client) Err() error {
	return cli.err
}

// cleanup closes all channels once goroutines are finished
func (cli *client) cleanup() {
	cli.wg.Wait()
	close(cli.inputChan)
	close(cli.outputChan)
	close(cli.errorChan)
	cli.conn.Close()
	close(cli.done)
}

// Output back to the IRC connection
func (cli *client) Output() chan<- IrcMessage {
	return cli.outputChan
}

// Input from the IRC connection
func (cli *client) Input() <-chan parser.Message {
	return cli.inputChan
}

// Errors from IRC input/output
func (cli *client) Errors() <-chan error {
	return cli.errorChan
}

// Close closes the connection & cancels goroutines
func (cli *client) Close() error {
	cli.cancelFunc()
	return nil
}

func (cli *client) Closed() bool {
	return cli.ctx.Err() != nil
}

// messageError is used for scanning to return either a message or an error
type messageError struct {
	Message *parser.Message
	Error   error
}

func (cli *client) scan() <-chan messageError {
	msgChan := make(chan messageError)
	go func() {
		defer close(msgChan)
		message, err := cli.scanner.Scan()
		msgChan <- messageError{message, err}
	}()
	return msgChan
}

// Scans the IRC messages and writes them to the input channel
func (cli *client) readInput() {
	defer cli.wg.Done()
	for {
		select {
		case <-cli.ctx.Done():
			return
		case message := <-cli.scan():
			err := message.Error
			if err != nil {
				if errors.Is(err, io.EOF) {
					cli.err = err
					cli.cancelFunc()
					return
				}
				continue
			}
			msg := *message.Message
			cli.inputChan <- msg
		}
	}
}

func (cli *client) error(err error) {
	if err == nil {
		return
	}
	select {
	case <-cli.ctx.Done():
		return
	default:
	}
	// TODO: Make this non-blocking?
	cli.errorChan <- err
}

// Writes each message from the channel to the IRC Connection
func (cli *client) setupOutput() {
	defer cli.wg.Done()
	// TODO: Rate limit per output type
	for {
		select {
		case <-cli.ctx.Done():
			return
		case output := <-cli.outputChan:
			if err := cli.write(output); err != nil {
				// Connection closed
				if errors.Is(err, io.EOF) {
					return
				}
				cli.error(err)
			}
		}
	}
}

func (cli *client) write(msg IrcMessage) error {
	_, err := cli.conn.Write(append(msg.Bytes(), '\r', '\n'))
	return err
}
