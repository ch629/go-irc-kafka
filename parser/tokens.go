package parser

type (
	TokenType int

	Token struct {
		Type    TokenType
		Literal string
	}
)

const (
	// Prefix
	ServerName TokenType = iota
	Nick
	User
	Host

	Command
	Param
)
