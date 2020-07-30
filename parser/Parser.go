package parser

import (
	"strings"
	"time"
)

type Message struct {
	Timestamp time.Time `json:"timestamp"`
	Prefix    string    `json:"prefix"`
	Command   string    `json:"command"`
	Params    string    `json:"params"`
}

var (
	sb     = strings.Builder{}
	bytes  = make(chan byte)
	Output = make(chan Message)

	inPrefix   = false
	index      = 0
	donePrefix = false

	prefix     = ""
	command    = ""
	parameters string
)

func PushByte(b byte) {
	bytes <- b
}

func PollChannel() {
	go func() {
		for b := range bytes {
			handleByte(b)
		}
	}()
}

func outputMessage(message Message) {
	Output <- message
}

func makeMessage() Message {
	return Message{
		Timestamp: time.Now().UTC(),
		Prefix:    prefix,
		Command:   command,
		Params:    parameters,
	}
}

func finishMessage() {
	outputMessage(makeMessage())
}

func resetMessage() {
	inPrefix = false
	donePrefix = false
	index = 0

	prefix = ""
	command = ""
	parameters = ""

	sb.Reset()
}

func write(b byte) {
	index++
	sb.WriteByte(b)
}

func addParameter() {
	parameters = sb.String()
	sb.Reset()
}

func handleByte(b byte) {
	if b == '\r' {
		addParameter()
		finishMessage()
		resetMessage()
		return
	}

	if b == '\n' {
		return
	}

	if index == 0 && b == ':' {
		inPrefix = true
		return
	}

	if inPrefix {
		if b == ' ' {
			prefix = sb.String()
			inPrefix = false
			sb.Reset()
			index++
			return
		}

		write(b)
		return
	}

	// TODO: Check performance vs using a nil pointer
	if len(command) == 0 {
		if b == ' ' {
			command = sb.String()
			sb.Reset()
			index++
			return
		}

		write(b)
		return
	}

	if !donePrefix && b == ':' {
		donePrefix = true
		index++
		return
	}

	write(b)
}
