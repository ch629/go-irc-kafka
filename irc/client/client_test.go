package client

import (
	"bytes"
	"context"
	"github.com/ch629/go-irc-kafka/irc/parser"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type stringMessage struct {
	Message string
}

func (mes *stringMessage) Bytes() []byte {
	return []byte(mes.Message)
}

func Test_InputSmall(t *testing.T) {
	var buf bufCloser
	ircClient := NewDefaultClient(context.Background(), &buf)
	defer ircClient.Close()
	buf.WriteString(":tmi.twitch.tv 001 thewolfpack :Welcome, GLHF!\r\n")

	select {
	case msg := <-ircClient.Input():
		assert.Equal(t, parser.Message{
			Tags:    map[string]string{},
			Prefix:  "tmi.twitch.tv",
			Command: "001",
			Params: []string{
				"thewolfpack",
				"Welcome, GLHF!",
			},
		}, msg)
	case err := <-ircClient.Errors():
		assert.NoError(t, err)
	case <-time.After(5 * time.Second):
		assert.Fail(t, "Timed out while getting input")
	}
}

func Test_OutputSmall(t *testing.T) {
	buf := &bufCloser{}
	ircClient := NewDefaultClient(context.Background(), buf)
	defer ircClient.Close()

	go consumeErrors(ircClient)
	go consumeInput(ircClient)

	msg := &stringMessage{"testing"}
	ircClient.Output() <- msg

	assert.Equal(t, "testing\r\n", buf.String())
}

func TestNewDefaultClient_EOF(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cli := NewDefaultClient(ctx, eofReadWriteCloser{})
	t.Cleanup(cancel)
	select {
	case <-cli.Done():
	case <-time.After(5 * time.Second):
		assert.Fail(t, "Client didn't close after EOF")
	}
}

func consumeErrors(client IrcClient) {
	for range client.Errors() {
	}
}

func consumeInput(client IrcClient) {
	for range client.Input() {
	}
}

type eofReadWriteCloser struct{}

func (eofReadWriteCloser) Read(_ []byte) (n int, err error) {
	return 0, io.EOF
}

func (eofReadWriteCloser) Write(_ []byte) (n int, err error) {
	return 0, io.EOF
}

func (eofReadWriteCloser) Close() error {
	return nil
}

type bufCloser struct {
	bytes.Buffer
}

func (bufCloser) Close() error {
	return nil
}
