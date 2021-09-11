package irc

const (
	// Inbound

	Ping       = "PING"
	EndOfMOTD  = "376"
	RoomState  = "ROOMSTATE"
	UserState  = "USERSTATE"
	UserNotice = "USERNOTICE"
	// ClearChat clears an entire user's chat
	ClearChat = "CLEARCHAT"
	// ClearMessage clears one single message
	ClearMessage = "CLEARMSG"
	HostTarget   = "HOSTTARGET"
	// Notice is received when room state has been updated or a channel is hosting another when initially joining
	Notice = "NOTICE"

	// Outbound

	Pong     = "PONG"
	Join     = "JOIN"
	Part     = "PART"
	Password = "PASS"
	Nickname = "NICK"

	// Both

	Capability     = "CAP"
	PrivateMessage = "PRIVMSG"

	// Errors

	// ERR_PASSWDMISMATCH
	// TODO: Mappings from IRC error codes to go errors
	ErrPasswordMismatch = "464"
)
