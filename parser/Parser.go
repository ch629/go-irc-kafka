package parser

import (
	"strings"
	"time"
)

type Message struct {
	Timestamp time.Time         `json:"timestamp"`
	Metadata  map[string]string `json:"metadata"`
	Username  string            `json:"username"`
	Prefix    string            `json:"prefix"`
	Command   string            `json:"command"`
	Params    string            `json:"params"`
}

var (
	sb     = strings.Builder{}
	bytes  = make(chan byte)
	Output = make(chan Message)

	metaName     = ""
	metadataDone = false

	metadata map[string]string
	prefix   *string

	// Prefix is the following split by ' '
	//  Username/Server
	//  Command
	//  Channel (For PRIVMSG)
	user       *string
	command    *string
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

func emptyOrValue(str *string) string {
	if str == nil {
		return ""
	}
	return *str
}

func makeMessage() Message {
	return Message{
		Timestamp: time.Now().UTC(),
		Metadata:  metadata,
		Username:  emptyOrValue(user),
		Command:   emptyOrValue(command),
		Prefix:    emptyOrValue(prefix),
		Params:    parameters,
	}
}

func finishMessage() {
	outputMessage(makeMessage())
}

func resetMessage() {
	metadataDone = false
	metadata = make(map[string]string)
	prefix = nil
	user = nil
	command = nil
	parameters = ""

	sb.Reset()
}

func write(b byte) {
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

	if handleMetadata(b) {
		return
	}

	if handlePrefix(b) {
		return
	}

	write(b)
}

func handleMetadata(b byte) bool {
	if !metadataDone {
		if b == '=' {
			metaName = sb.String()
			sb.Reset()
		} else if b == ';' {
			metadata[metaName] = sb.String()
			metaName = ""
			sb.Reset()
		} else if b == ':' && (len(metadata) == 0 || metaName == "user-type") {
			metadataDone = true
			sb.Reset()
		} else {
			write(b)
		}

		return true
	}

	return false
}

func handlePrefix(b byte) bool {
	if prefix == nil {
		if b == ' ' {
			if user == nil {
				tmpUser := sb.String()
				user = &tmpUser
				sb.Reset()
			} else if command == nil {
				tmpCommand := sb.String()
				command = &tmpCommand
				sb.Reset()
			} else {
				// We need this space, if we aren't breaking it down any further here (I don't think we can as it changes dependent on the commands)
				//  Unless we store the rest in an array of strings?
				write(b)
			}
		} else if b == ':' {
			tmpPrefix := sb.String()
			prefix = &tmpPrefix

			sb.Reset()
		} else {
			write(b)
		}
		return true
	}
	return false
}
