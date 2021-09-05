package bot

import (
	"sync"
	"time"

	"github.com/ch629/go-irc-kafka/twitch"
)

type (
	State struct {
		chanMux sync.RWMutex
		capMux  sync.RWMutex
		// TODO: Should these be exported?
		Channels     map[string]*Channel
		Capabilities []twitch.Capability
	}
	Channel struct {
		State *ChannelState `json:"state"`
		User  *UserState    `json:"user"`
	}
	ChannelState struct {
		EmoteOnly      bool          `json:"emote_only"`
		FollowerOnly   time.Duration `json:"follower_only"`
		R9k            bool          `json:"r9k"`
		Slow           time.Duration `json:"slow"`
		SubscriberOnly bool          `json:"subscriber_only"`
	}
	UserState struct {
		Mod        bool `json:"mod,omitempty"`
		Subscriber bool `json:"subscriber,omitempty"`
	}
)

func (s *ChannelState) Update(emoteOnly, r9k, subscriber bool, follower, slow time.Duration) {
	s.EmoteOnly = emoteOnly
	s.R9k = r9k
	s.SubscriberOnly = subscriber
	s.FollowerOnly = follower
	s.Slow = slow
}

func (s *UserState) Update(mod, subscriber bool) {
	s.Mod = mod
	s.Subscriber = subscriber
}
