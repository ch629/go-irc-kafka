package bot

import "github.com/ch629/go-irc-kafka/domain"

type MessageHandler struct {
	onPrivateMessage func(msg domain.ChatMessage)
	onBan            func(ban domain.Ban)
	onError          func(err error)
}

func (h *MessageHandler) OnPrivateMessage(f func(msg domain.ChatMessage)) {
	h.onPrivateMessage = f
}

func (h *MessageHandler) OnBan(f func(ban domain.Ban)) {
	h.onBan = f
}

func (h *MessageHandler) OnError(f func(err error)) {
	h.onError = f
}
