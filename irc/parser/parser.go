package parser

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"
)

// maxMessageLength is the maximum amount of runes to read before returning an error
const maxMessageLength = 512

var (
	eof = rune(0)

	ErrEmptyMessage = errors.New("empty message")
	ErrNoCommand    = errors.New("no command")
	ErrNoPrefix     = errors.New("no prefix")
	ErrTooLong      = errors.New("read for too long")
)

// https://ircv3.net/specs/extensions/message-tags.html
// https://tools.ietf.org/html/rfc1459.html#section-2.3.1
type (
	Scanner struct {
		*bufio.Reader
	}

	Message struct {
		Tags    map[string]string `json:"tags,omitempty"`
		Prefix  string            `json:"prefix,omitempty"`
		Command string            `json:"command"`
		Params  []string          `json:"params,omitempty"`
	}
)

func (m *Message) HasTags() bool {
	return len(m.Tags) > 0
}

func (m *Message) HasPrefix() bool {
	return len(m.Prefix) > 0
}

func NewScanner(r io.Reader) *Scanner {
	return &Scanner{
		Reader: bufio.NewReader(r),
	}
}

// TODO: EOF checks
// TODO: Escape inside of other funcs
// TODO: Maximum length check, currently we do a check in readUntil, but there is a maximum of 512 chars in an entire message including CRLF
// Scan scans a line from the reader
func (s *Scanner) Scan() (*Message, error) {
	if s.isCrlf() {
		return nil, ErrEmptyMessage
	}

	r, err := s.peekRune()

	if err != nil {
		return nil, fmt.Errorf("failed to peek rune due to %w", err)
	}

	if r == eof {
		return nil, io.EOF
	}

	message := &Message{
		Tags:   make(map[string]string),
		Params: make([]string, 0),
	}

	if r == '@' {
		s.consume()
		if message.Tags, err = s.readTags(); err != nil {
			return nil, fmt.Errorf("failed to read tags due to %w", err)
		}

		if r, err = s.peekRune(); err != nil {
			return nil, fmt.Errorf("failed to peek rune due to %w", err)
		}
	}

	if r == ':' {
		s.consume()
		if message.Prefix, err = s.readPrefix(); err != nil {
			return nil, fmt.Errorf("failed to read prefix due to %w", err)
		}
	}

	if message.Command, err = s.readCommand(); err != nil {
		return nil, fmt.Errorf("failed to read command due to %w", err)
	}

	if message.Params, err = s.readParams(); err != nil {
		return nil, fmt.Errorf("failed to read params due to %w", err)
	}

	return message, nil
}

var escapedMap = map[rune]rune{
	':':  ';',
	's':  ' ',
	'\\': '\\',
	'r':  '\r',
	'n':  '\n',
}

// readTags Reads the tags BNF
func (s *Scanner) readTags() (map[string]string, error) {
	tags := make(map[string]string)
	key := true
	var keyBuilder strings.Builder
	var valueBuilder strings.Builder

	for {
		escaped := false
		r := s.read()

		if r == '\\' {
			escaped = true
			r = s.read()
		}

		if !escaped {
			if r == ' ' {
				// Push key & value pair to map if not empty, return tags
				if keyBuilder.Len() > 0 && valueBuilder.Len() > 0 {
					tags[keyBuilder.String()] = valueBuilder.String()
				}
				return tags, nil
			}

			if r == '=' {
				key = false
				continue
			}

			if r == ';' {
				// Push key & value pair to map if not empty
				if keyBuilder.Len() > 0 && valueBuilder.Len() > 0 {
					tags[keyBuilder.String()] = valueBuilder.String()
				}

				keyBuilder.Reset()
				valueBuilder.Reset()
				key = true
				continue
			}

		}

		if escaped {
			if escapedRune, ok := escapedMap[r]; ok {
				r = escapedRune
			}
		}

		if key {
			keyBuilder.WriteRune(r)
		} else {
			valueBuilder.WriteRune(r)
		}
	}
}

// TODO: Prefix struct?
// readPrefix Reads the prefix BNF
func (s *Scanner) readPrefix() (str string, err error) {
	if str, err = s.readUntil([]rune{' '}, []rune{}); err != nil {
		return
	}
	if len(str) == 0 {
		return "", ErrNoPrefix
	}
	return
}

// readCommand Reads the command BNF
func (s *Scanner) readCommand() (str string, err error) {
	if str, err = s.readUntil([]rune{' '}, []rune{}); err != nil {
		return
	}
	if len(str) == 0 {
		return "", ErrNoCommand
	}
	return
}

// readParams Reads the param BNF
func (s *Scanner) readParams() ([]string, error) {
	params := make([]string, 0)
	for {
		r, err := s.peekRune()

		if err != nil {
			return nil, fmt.Errorf("failed to peek rune due to %w", err)
		}

		if r == ' ' {
			continue
		}

		if s.isCrlf() {
			return params, nil
		}

		if r == ':' {
			// Consume :
			s.consume()
			trailing, err := s.readParamTrailing()

			if err != nil {
				return nil, fmt.Errorf("failed to read trailing param due to %w", err)
			}

			params = append(params, trailing)
			return params, nil
		}

		middle, err := s.readParamMiddle()

		if err != nil {
			return nil, fmt.Errorf("failed to read param middle due to %w", err)
		}

		params = append(params, middle)
	}
}

// readParamTrailing Reads the param trailing BNF
// returns ErrTooLong if too many runes have been read
func (s *Scanner) readParamTrailing() (string, error) {
	var sb strings.Builder
	for i := 0; i < maxMessageLength; i++ {
		if s.isCrlf() {
			return sb.String(), nil
		}

		r := s.read()
		if r == eof {
			return "", io.EOF
		}
		sb.WriteRune(r)
	}
	return "", ErrTooLong
}

// readParamMiddle Reads the param middle BNF
func (s *Scanner) readParamMiddle() (string, error) {
	return s.readUntil([]rune{' '}, []rune{':', '\r'})
}

// readUntil reads runes up until is is either in the untilInclusive slice or untilExclusive
// untilInclusive will consume the rune if it is found
// untilExclusive will not consume the rune if it is found
// returns ErrTooLong if too many runes have been read
func (s *Scanner) readUntil(untilInclusive []rune, untilExclusive []rune) (string, error) {
	var contains = func(runes []rune, r rune) bool {
		for _, u := range runes {
			if r == u {
				return true
			}
		}
		return false
	}

	var sb strings.Builder

	for i := 0; i < maxMessageLength; i++ {
		r, err := s.peekRune()
		if err != nil {
			return "", err
		}
		if contains(untilInclusive, r) {
			s.consume()
			return sb.String(), nil
		}
		if contains(untilExclusive, r) {
			return sb.String(), nil
		}
		sb.WriteRune(s.read())
	}
	return "", ErrTooLong
}

// read Reads and consumes a single rune from the Scanner
func (s *Scanner) read() (r rune) {
	var err error
	if r, _, err = s.ReadRune(); err != nil {
		r = eof
	}
	return
}

// consume Consumes a single rune from the Scanner with no response
func (s *Scanner) consume() {
	_, _, _ = s.ReadRune()
}

// peekRune Reads a single rune from the Scanner without consuming it
func (s *Scanner) peekRune() (rune, error) {
	r, _, err := s.ReadRune()
	if err != nil {
		return r, err
	}
	return r, s.UnreadRune()
}

// Detects if the next runes are CRLF
func (s *Scanner) isCrlf() bool {
	if bs, err := s.Peek(2); err != nil || bs[0] != '\r' || bs[1] != '\n' {
		return false
	}

	// Consume
	_, _ = s.Read(make([]byte, 2))
	return true
}
