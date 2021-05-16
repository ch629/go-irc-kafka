package twitch

import (
	"fmt"
	"github.com/ch629/go-irc-kafka/irc/client"
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

func CapabilityFromParam(param string) Capability {
	return map[string]Capability{
		"twitch.tv/membership": MEMBERSHIP,
		"twitch.tv/tags":       TAGS,
		"twitch.tv/commands":   COMMANDS,
	}[param]
}

func (message *capabilityRequestMessage) Bytes() []byte {
	return []byte(fmt.Sprintf("CAP REQ :twitch.tv/%v", message.Capability.String()))
}

func MakeCapabilityRequest(capability Capability) client.IrcMessage {
	return &capabilityRequestMessage{
		Capability: capability,
	}
}
