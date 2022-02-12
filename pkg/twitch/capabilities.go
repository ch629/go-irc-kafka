package twitch

import (
	"fmt"
	"strings"

	"github.com/ch629/go-irc-kafka/pkg/irc/client"
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

func (cap Capability) MarshalText() (text []byte, err error) {
	return []byte(cap.String()), nil
}

func (cap *Capability) UnmarshalText(text []byte) error {
	*cap = CapabilityFromParam(string(text))
	return nil
}

func (cap Capability) String() string {
	return []string{
		"membership",
		"tags",
		"commands",
	}[cap]
}

func CapabilityFromParam(param string) Capability {
	capabilityName := strings.TrimPrefix(param, "twitch.tv/")
	return map[string]Capability{
		"membership": MEMBERSHIP,
		"tags":       TAGS,
		"commands":   COMMANDS,
	}[capabilityName]
}

func (message capabilityRequestMessage) Bytes() []byte {
	return []byte(fmt.Sprintf("CAP REQ :twitch.tv/%v", message.Capability.String()))
}

func MakeCapabilityRequest(capability Capability) client.IrcMessage {
	return capabilityRequestMessage{
		Capability: capability,
	}
}
