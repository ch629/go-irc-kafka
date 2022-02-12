package client

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/ch629/go-irc-kafka/pkg/irc/parser"
)

type (
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
	ctx, cancelFunc := context.WithCancel(ctx)
	return &client{
		conn:       conn,
		inputChan:  make(chan parser.Message),
		errorChan:  make(chan error),
		done:       make(chan struct{}),
		scanner:    parser.NewScanner(conn),
		ctx:        ctx,
		cancelFunc: cancelFunc,
	}
}

// TODO: Pass ctx?
func (c *client) ConsumeMessages() {
	defer c.cleanup()
	c.readInput()
}

func (c *client) Send(messages ...IrcMessage) error {
	for _, msg := range messages {
		// TODO: Retry?
		if _, err := c.conn.Write(append(msg.Bytes(), '\r', '\n')); err != nil {
			return fmt.Errorf("conn.Write: %w", err)
		}
	}
	return nil
}

func (c *client) Err() error {
	return c.err
}

func (c *client) Done() <-chan struct{} {
	return c.done
}

// cleanup closes all channels once goroutines are finished
func (c *client) cleanup() {
	close(c.inputChan)
	close(c.errorChan)
	c.conn.Close()
	close(c.done)
}

// Input from the IRC connection
func (c *client) Input() <-chan parser.Message {
	return c.inputChan
}

// Errors from IRC input/output
func (c *client) Errors() <-chan error {
	return c.errorChan
}

// Close closes the connection & cancels goroutines
func (c *client) Close() error {
	c.cancelFunc()
	return nil
}

func (c *client) Closed() bool {
	return c.ctx.Err() != nil
}

// messageError is used for scanning to return either a message or an error
type messageError struct {
	Message *parser.Message
	Error   error
}

func (c *client) scan() <-chan messageError {
	msgChan := make(chan messageError)
	go func() {
		defer close(msgChan)
		message, err := c.scanner.Scan()
		msgChan <- messageError{message, err}
	}()
	return msgChan
}

// Scans the IRC messages and writes them to the input channel
func (c *client) readInput() {
	for {
		select {
		case <-c.ctx.Done():
			return
		case message := <-c.scan():
			err := message.Error
			if err != nil {
				if errors.Is(err, io.EOF) {
					c.err = err
					c.cancelFunc()
					return
				}
				c.error(err)
				continue
			}
			msg := *message.Message
			c.inputChan <- msg
		}
	}
}

func (c *client) error(err error) {
	if err != nil && !c.Closed() {
		select {
		case c.errorChan <- err:
		default:
		}
	}
}
