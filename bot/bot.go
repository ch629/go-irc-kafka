package bot

import (
	"context"
	"errors"
	"github.com/ch629/go-irc-kafka/irc/client"
	"github.com/ch629/go-irc-kafka/irc/parser"
	"github.com/ch629/go-irc-kafka/logging"
	"github.com/ch629/go-irc-kafka/twitch"
	"go.uber.org/zap"
	"time"
)

// Bot
// TODO: Interface
type Bot struct {
	ctx       context.Context
	state     *State
	client    client.IrcClient
	onMessage func(*Bot, parser.Message) error
}

var ErrNotInChannel = errors.New("not in channel")

func NewBot(ctx context.Context, client client.IrcClient, onMessage func(bot *Bot, message parser.Message) error) *Bot {
	b := &Bot{
		ctx: ctx,
		state: &State{
			Channels:     make(map[string]*Channel),
			Capabilities: make([]twitch.Capability, 0),
		},
		client:    client,
		onMessage: onMessage,
	}
	go b.handleMessages()
	return b
}

func (b *Bot) handleMessages() {
	for {
		select {
		case <-b.ctx.Done():
			return
		case msg := <-b.client.Input():
			// TODO: Error channel
			// TODO: Try to run these in a goroutine, might need a channel of updates otherwise we get nil
			//  references with joining channels & then updating them
			if err := b.onMessage(b, msg); err != nil {
				logging.Logger().Error("Failed to handle message", zap.String("command", msg.Command), zap.Error(err))
			}
		}
	}
}

// Login sends the required messages to IRC to login
func (b *Bot) Login(name, pass string) {
	b.Send(twitch.MakePassCommand(pass), twitch.MakeNickCommand(name))
}

// Send sends messages to IRC
func (b *Bot) Send(messages ...client.IrcMessage) {
	for _, message := range messages {
		b.client.Output() <- message
	}
}

func (b *Bot) GetChannelData(channel string) (*ChannelState, error) {
	s := b.state
	s.chanMux.RLock()
	defer s.chanMux.RUnlock()
	ch, ok := s.Channels[channel]
	if !ok {
		return nil, ErrNotInChannel
	}
	data := ch.State
	return &data, nil
}

// AddChannel adds the channel to State
func (b *Bot) AddChannel(channel string) {
	s := b.state
	s.chanMux.Lock()
	defer s.chanMux.Unlock()
	s.Channels[channel] = &Channel{
		State: ChannelState{},
		User:  UserState{},
	}
}

// AddCapability adds the Capability to State
func (b *Bot) AddCapability(capability twitch.Capability) {
	s := b.state
	if !b.HasCapability(capability) {
		// TODO: Should we be sharing the mutex for capabilities & channels?
		s.capMux.Lock()
		defer s.capMux.Unlock()
		s.Capabilities = append(s.Capabilities, capability)
	}
}

// HasCapability returns whether the capability exists within State
func (b *Bot) HasCapability(capability twitch.Capability) bool {
	b.state.capMux.RLock()
	defer b.state.capMux.RUnlock()
	for _, c := range b.state.Capabilities {
		if c == capability {
			return true
		}
	}
	return false
}

// UpdateUserState updates UserState for the given channel
func (b *Bot) UpdateUserState(channel string, mod, subscriber bool) error {
	s := b.state
	s.chanMux.RLock()
	defer s.chanMux.RUnlock()
	c, ok := s.Channels[channel]
	if !ok {
		return ErrNotInChannel
	}
	c.User.Update(mod, subscriber)
	return nil
}

// UpdateChannel updates ChannelState for the provided channel
func (b *Bot) UpdateChannel(channel string, emoteOnly, r9k, subscriber bool, follower, slow time.Duration) error {
	s := b.state
	s.chanMux.RLock()
	defer s.chanMux.RUnlock()
	c, ok := s.Channels[channel]
	if !ok {
		return ErrNotInChannel
	}
	c.State.Update(emoteOnly, r9k, subscriber, follower, slow)
	return nil
}

// RemoveChannel removes the channel from State
func (b *Bot) RemoveChannel(channel string) {
	s := b.state
	s.chanMux.Lock()
	defer s.chanMux.Unlock()
	delete(s.Channels, channel)
}

// Close closes the connection to IRC
func (b *Bot) Close() error {
	return b.client.Close()
}

// RequestJoinChannel sends a Join message to IRC
func (b *Bot) RequestJoinChannel(channel string) {
	b.Send(twitch.MakeJoinCommand(channel))
}

// RequestCapability sends Capability requests to IRC
func (b *Bot) RequestCapability(capabilities ...twitch.Capability) {
	for _, capability := range capabilities {
		b.Send(twitch.MakeCapabilityRequest(capability))
	}
}

// RequestLeaveChannel sends a Part message to IRC & removes the channel from State
func (b *Bot) RequestLeaveChannel(channel string) {
	b.Send(twitch.MakePartCommand(channel))
}

func (b *Bot) InChannel(channel string) bool {
	_, ok := b.state.Channels[channel]
	return ok
}