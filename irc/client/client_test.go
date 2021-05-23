package client

import (
	"bufio"
	"context"
	"github.com/ch629/go-irc-kafka/irc/parser"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type (
	stringMessage struct {
		Message string
	}
	eofReadWriteCloser struct{}
)

func (mes *stringMessage) Bytes() []byte {
	return []byte(mes.Message)
}

func (eofReadWriteCloser) Read([]byte) (n int, err error) {
	return 0, io.EOF
}

func (eofReadWriteCloser) Write([]byte) (n int, err error) {
	return 0, io.EOF
}

func (eofReadWriteCloser) Close() error {
	return nil
}

func Test_Input(t *testing.T) {
	conn := MakeMockConn()
	ircClient := NewClient(context.Background(), conn)
	defer ircClient.Close()
	io.WriteString(conn.ClientWriter, ":tmi.twitch.tv 001 thewolfpack :Welcome, GLHF!\r\n")

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
	case <-time.After(2 * time.Second):
		assert.Fail(t, "Timed out while getting input")
	}
}

func Test_Output(t *testing.T) {
	conn := MakeMockConn()
	ircClient := NewClient(context.Background(), conn)
	defer ircClient.Close()
	ircClient.Output() <- &stringMessage{"testing"}

	go consumeErrors(ircClient)
	go consumeInput(ircClient)

	reader := bufio.NewReader(conn.ClientReader)
	line, err := reader.ReadString('\n')
	assert.NoError(t, err)
	assert.Equal(t, "testing\r\n", line)
}

func TestNewClient_EOF(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cli := NewClient(ctx, eofReadWriteCloser{})
	t.Cleanup(cancel)
	select {
	case <-cli.Done():
	case <-time.After(2 * time.Second):
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
