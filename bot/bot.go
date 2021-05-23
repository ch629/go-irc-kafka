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
			if err := b.onMessage(b, msg); err != nil {
				logging.Logger().Error("Failed to handle message", zap.String("command", msg.Command), zap.Error(err))
			}
		}
	}
}

// Login sends the required messages to IRC to login
func (b *Bot) Login(name, pass string) {
	b.send(twitch.MakePassCommand(pass), twitch.MakeNickCommand(name))
}

// Pong sends a pong message back to the IRC server
func (b *Bot) Pong(params parser.Params) {
	b.send(twitch.MakePongCommand(params[0]))
}

// send sends messages to IRC
func (b *Bot) send(messages ...client.IrcMessage) {
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
	return data, nil
}

// AddChannel adds the channel to State
func (b *Bot) AddChannel(channel string) {
	s := b.state
	s.chanMux.Lock()
	defer s.chanMux.Unlock()
	s.Channels[channel] = &Channel{
		User:  &UserState{},
		State: &ChannelState{},
	}
}

// AddCapability adds the Capability to State
func (b *Bot) AddCapability(capability twitch.Capability) {
	s := b.state
	if !b.HasCapability(capability) {
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
	b.send(twitch.MakeJoinCommand(channel))
}

// RequestCapability sends Capability requests to IRC
func (b *Bot) RequestCapability(capabilities ...twitch.Capability) {
	for _, capability := range capabilities {
		b.send(twitch.MakeCapabilityRequest(capability))
	}
}

// RequestLeaveChannel sends a Part message to IRC & removes the channel from State
func (b *Bot) RequestLeaveChannel(channel string) {
	b.send(twitch.MakePartCommand(channel))
}

// InChannel returns whether the bot is in the given channel
func (b *Bot) InChannel(channel string) bool {
	_, ok := b.state.Channels[channel]
	return ok
}

// Channels returns the names of the channels the Bot is in
func (b *Bot) Channels() []string {
	b.state.chanMux.RLock()
	defer b.state.chanMux.RUnlock()
	channels := make([]string, len(b.state.Channels))
	i := 0
	for key := range b.state.Channels {
		channels[i] = key
		i++
	}
	return channels
}

// Capabilities returns the capabilities given to the Bot
func (b *Bot) Capabilities() []twitch.Capability {
	b.state.capMux.RLock()
	defer b.state.capMux.RUnlock()
	return b.state.Capabilities
}
