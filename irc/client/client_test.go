package client

import (
	"bytes"
	"go-irc/irc/parser"
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
	var buf bytes.Buffer
	ircClient := NewDefaultClient(&buf)
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
	var buf bytes.Buffer
	ircClient := NewDefaultClient(&buf)

	ircClient.Output() <- &stringMessage{
		Message: "testing",
	}

	str, err := buf.ReadString('\n')

	assert.NoError(t, err)

	assert.Equal(t, "testing\r\n", str)
}
