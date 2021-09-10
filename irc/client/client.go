package client

import (
	"context"
	"errors"
	"io"

	"github.com/ch629/go-irc-kafka/irc/parser"
)

type (
	// TODO: Store somewhere why it was closed so we can attempt to reconnect if it was just disconnected
	IrcClient interface {
		io.Closer
		// ConsumeMessages reads the bytes from the connection & parses them, writing to the input channel
		ConsumeMessages()
		// Input is a channel of messages coming from IRC
		Input() <-chan parser.Message
		// Send sends the IrcMessage to the IRC client
		Send(message ...IrcMessage) error
		// Errors is a channel of errors generated when reading or writing to IRC
		Errors() <-chan error
		// Closed is whether the client connection is closed
		Closed() bool
		// Done is a channel to notify consumers when the client is done closing
		Done() <-chan struct{}
		// Err contains an error when the client closes suddenly
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
		errorChan  chan error
		scanner    parser.Scanner
		done       chan struct{}
		err        error
	}
)

func NewClient(ctx context.Context, conn io.ReadWriteCloser) IrcClient {
	cli := &client{
		conn:      conn,
		inputChan: make(chan parser.Message),
		errorChan: make(chan error),
		done:      make(chan struct{}),
		scanner:   parser.NewScanner(conn),
	}
	cli.ctx, cli.cancelFunc = context.WithCancel(ctx)

	return cli
}

// TODO: Pass ctx?
func (cli *client) ConsumeMessages() {
	defer cli.cleanup()
	cli.readInput()
}

func (cli *client) Send(messages ...IrcMessage) error {
	for _, msg := range messages {
		// TODO: Retry?
		if _, err := cli.conn.Write(append(msg.Bytes(), '\r', '\n')); err != nil {
			return err
		}
	}
	return nil
}

func (cli *client) Err() error {
	return cli.err
}

func (cli *client) Done() <-chan struct{} {
	return cli.done
}

// cleanup closes all channels once goroutines are finished
func (cli *client) cleanup() {
	close(cli.inputChan)
	close(cli.errorChan)
	cli.conn.Close()
	close(cli.done)
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
				cli.error(err)
				continue
			}
			msg := *message.Message
			cli.inputChan <- msg
		}
	}
}

func (cli *client) error(err error) {
	if err != nil && !cli.Closed() {
		select {
		case cli.errorChan <- err:
		default:
		}
	}
}

func (cli *client) write(msg IrcMessage) error {
	_, err := cli.conn.Write(append(msg.Bytes(), '\r', '\n'))
	return err
}
