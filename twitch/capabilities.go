package twitch

import (
	"fmt"
	"go-irc/irc/client"
)

type (
	Capability               int
	capabilityRequestMessage struct {
		Capability Capability
	}
)

const (
	MEMBERSHIP Capability = iota
	TAGS
	COMMANDS
)

func (cap Capability) String() string {
	return []string{
		"membership",
		"tags",
		"commands",
	}[cap]
}

func (message *capabilityRequestMessage) Bytes() []byte {
	return []byte(fmt.Sprintf("CAP REQ :twitch.tv/%v", message.Capability.String()))
}

func MakeCapabilityRequest(capability Capability) client.IrcMessage {
	return &capabilityRequestMessage{
		Capability: capability,
	}
}
