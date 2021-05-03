package parser

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"
	"unicode/utf8"
)

var (
	eof = rune(0)

	ErrReadingRune  = errors.New("error reading rune")
	ErrEmptyMessage = errors.New("empty message")
)

// https://ircv3.net/specs/extensions/message-tags.html
// https://tools.ietf.org/html/rfc1459.html#section-2.3.1
type (
	Scanner struct {
		*bufio.Reader
	}

	Message struct {
		Tags    map[string]string
		Prefix  string
		Command string
		Params  []string
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
		Tags:   map[string]string{},
		Params: []string{},
	}

	if r == '@' {
		s.consume()
		message.Tags, err = s.readTags()

		if err != nil {
			return nil, fmt.Errorf("failed to read tags due to %w", err)
		}

		r, err = s.peekRune()

		if err != nil {
			return nil, fmt.Errorf("failed to peek rune due to %w", err)
		}
	}

	if r == ':' {
		s.consume()
		message.Prefix, err = s.readPrefix()

		if err != nil {
			return nil, fmt.Errorf("failed to read prefix due to %w", err)
		}
	}

	message.Command, err = s.readCommand()

	if err != nil {
		return nil, fmt.Errorf("failed to read command due to %w", err)
	}

	message.Params, err = s.readParams()

	if err != nil {
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
				// Push key & value pair to map, return tags
				tags[keyBuilder.String()] = valueBuilder.String()
				return tags, nil
			}

			if r == '=' {
				key = false
				continue
			}

			if r == ';' {
				// Push key & value pair to map
				tags[keyBuilder.String()] = valueBuilder.String()

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
func (s *Scanner) readPrefix() (string, error) {
	var sb strings.Builder
	for {
		r := s.read()

		if r == ' ' {
			return sb.String(), nil
		}

		sb.WriteRune(r)
	}
}

// readCommand Reads the command BNF
func (s *Scanner) readCommand() (string, error) {
	var sb strings.Builder
	for {
		r := s.read()

		if r == ' ' {
			return sb.String(), nil
		}

		sb.WriteRune(r)
	}
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
func (s *Scanner) readParamTrailing() (string, error) {
	var sb strings.Builder
	for {
		if s.isCrlf() {
			return sb.String(), nil
		}

		r := s.read()
		if r == eof {
			return "", io.EOF
		}
		sb.WriteRune(r)
	}
}

// readParamMiddle Reads the param middle BNF
func (s *Scanner) readParamMiddle() (string, error) {
	var sb strings.Builder
	for {
		r, err := s.peekRune()

		if err != nil {
			return "", fmt.Errorf("failed to peek rune due to %w", err)
		}

		if r == ' ' {
			// Consume ' '
			s.consume()
			return sb.String(), nil
		}

		// TODO: Handle escaped
		// Some twitch messages omit the ':' & just directly CRLF
		if r == ':' || r == '\r' {
			return sb.String(), nil
		}
		r = s.read()
		sb.WriteRune(r)
	}
}

// read Reads and consumes a single rune from the Scanner
func (s *Scanner) read() rune {
	ch, _, err := s.ReadRune()

	if err != nil {
		return eof
	}
	return ch
}

// consume Consumes a single rune from the Scanner with no response
func (s *Scanner) consume() {
	_, _, _ = s.ReadRune()
}

// peekRune Reads a single rune from the Scanner without consuming it
func (s *Scanner) peekRune() (rune, error) {
	for peekBytes := 4; peekBytes > 0; peekBytes-- { // unicode rune can be up to 4 bytes
		b, err := s.Peek(peekBytes)
		if err == nil {
			r, _ := utf8.DecodeRune(b)
			if r == utf8.RuneError {
				return r, ErrReadingRune
			}
			return r, nil
		}
	}

	return eof, io.EOF
}

// TODO: Error if it's \r without \n?
// Detects if the next runes are CRLF
func (s *Scanner) isCrlf() bool {
	r, err := s.peekRune()

	if err != nil || r != '\r' {
		return false
	}

	// Consume \r
	s.consume()

	return s.read() == '\n'
}
