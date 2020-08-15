package parser

import (
	"strings"
	"time"
)

var (
	lastByte       byte
	hasInitialized bool

	inMetadata bool
	metadata   map[string]string
	inPrefix   bool
	prefix     string
	inCommand  bool
	command    string
	inArgs     bool
	args       []string

	buffer strings.Builder

	metaName string

	Output = make(chan NewMessage)
)

func init() {
	resetState()
}

func write(b byte) {
	buffer.WriteByte(b)
}

func resetState() {
	lastByte = '\n'
	hasInitialized = false

	inMetadata = false
	metadata = make(map[string]string)
	inPrefix = false
	prefix = ""
	inCommand = false
	command = ""
	inArgs = false
	args = []string{}

	buffer.Reset()
}

func initState(b byte) {
	hasInitialized = true
	if b == '@' {
		inMetadata = true
	} else if b == ':' {
		inPrefix = true
	} else {
		inCommand = true
		// First byte of the command
		write(b)
	}
}

func nextState() {
	if inMetadata {
		inMetadata = false
		inPrefix = true

		metadata[metaName] = popBuffer()
		metaName = ""
	} else if inPrefix {
		inPrefix = false
		inCommand = true
	} else if inCommand {
		inCommand = false
		inArgs = true
	} else if inArgs {
		args = append(args, popBuffer())
	}
}

type NewMessage struct {
	Timestamp time.Time         `json:"timestamp"`
	Metadata  map[string]string `json:"metadata"`
	Prefix    string            `json:"prefix"`
	Command   string            `json:"command"`
	Args      []string          `json:"args"`
}

func buildMessage() {
	args = append(args, popBuffer())

	newMes := NewMessage{
		Timestamp: time.Now().UTC(),
		Metadata:  metadata,
		Prefix:    prefix,
		Command:   command,
		Args:      args,
	}

	Output <- newMes
}

// TODO: Rather than pushing straight into this, would it be better to be polling a byte channel?
func HandleByte(b byte) {
	defer func() {
		lastByte = b
	}()

	if b == '\n' {
		// Ignore this as we are handling it in the \r
		// TODO: We could look at previous byte being \r then handling it in here? Would allow us to handle \n if it is received in a message normally?
		return
	}

	// TODO: Any validation if we get a malformed message?
	if !hasInitialized {
		initState(b)
		return
	}

	if b == '\r' {
		buildMessage()
		resetState()
		return
	}

	if lastByte == ' ' && b == ':' {
		nextState()
		return
	}

	if inMetadata {
		handleMetadata(b)
		return
	}

	if inPrefix {
		handlePrefix(b)
		return
	}

	if inCommand {
		handleCommand(b)
		return
	}

	if inArgs {
		handleArgs(b)
	}
}

func popBuffer() string {
	tmp := buffer.String()
	buffer.Reset()
	return tmp
}

func handleMetadata(b byte) {
	if b == '=' {
		metaName = popBuffer()
	} else if b == ';' {
		if buffer.Len() > 0 {
			metadata[metaName] = popBuffer()
		}
		metaName = ""
	} else {
		write(b)
	}
}

func handlePrefix(b byte) {
	if b == ' ' {
		// TODO: Do we break this down into servername, nick, user & host?
		prefix = popBuffer()
		nextState()
	} else {
		write(b)
	}
}

func handleCommand(b byte) {
	if b == ' ' {
		command = popBuffer()
		nextState()
	} else {
		write(b)
	}
}

func handleArgs(b byte) {
	write(b)
}
