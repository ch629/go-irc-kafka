package irc

const (
	Ping           = "PING"
	Pong           = "PONG"
	Join           = "JOIN"
	Part           = "PART"
	EndOfMOTD      = "376"
	Capability     = "CAP"
	RoomState      = "ROOMSTATE"
	UserState      = "USERSTATE"
	PrivateMessage = "PRIVMSG"
	Password       = "PASS"
	Nickname       = "NICK"
	UserNotice     = "USERNOTICE"
	// ClearChat clears an entire user's chat
	ClearChat = "CLEARCHAT"
	// ClearMessage clears one single message
	ClearMessage = "CLEARMSG"
	HostTarget   = "HOSTTARGET"
	// Notice is received when room state has been updated or a channel is hosting another when initially joining
	Notice = "NOTICE"
)
