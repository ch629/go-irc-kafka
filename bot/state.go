package bot

import (
	"github.com/ch629/go-irc-kafka/twitch"
	"sync"
	"time"
)

type (
	State struct {
		chanMux sync.RWMutex
		capMux  sync.RWMutex
		// TODO: Should these be exported?
		Channels     map[string]Channel
		Capabilities []twitch.Capability
	}
	Channel struct {
		State ChannelState
		User  UserState
	}
	ChannelState struct {
		EmoteOnly      bool
		FollowerOnly   time.Duration
		R9k            bool
		Slow           time.Duration
		SubscriberOnly bool
	}
	UserState struct {
		Mod        bool
		Subscriber bool
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
