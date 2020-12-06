package parser

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"unicode/utf8"
)

var eof = rune(0)

// https://ircv3.net/specs/extensions/message-tags.html
// https://tools.ietf.org/html/rfc1459.html#section-2.3.1
type (
	Scanner struct {
		*bufio.Reader
	}

	TestMessage struct {
		Tags    map[string]string
		Prefix  string
		Command string
		Params  []string
	}
)

func (m *TestMessage) HasTags() bool {
	return len(m.Tags) > 0
}

func (m *TestMessage) HasPrefix() bool {
	return len(m.Prefix) > 0
}

func NewScanner(r io.Reader) *Scanner {
	return &Scanner{
		Reader: bufio.NewReader(r),
	}
}

// TODO: A blank line with just CRLF should be valid, but just ignored
// TODO: Rework with peek instead of read & unread
func (s *Scanner) Scan() (*TestMessage, error) {
	// TODO: If command is first, then r needs to be passed in
	r := s.read()

	if r == eof {
		return nil, fmt.Errorf("first character was an eof")
	}

	message := &TestMessage{}

	if r == '@' {
		tags, err := s.readTags()

		if err != nil {
			return nil, fmt.Errorf("failed to read tags due to %w", err)
		}

		message.Tags = tags
		r = s.read()
	}

	if r == ':' {
		prefix, err := s.readPrefix()

		if err != nil {
			return nil, fmt.Errorf("failed to read prefix due to %w", err)
		}

		message.Prefix = prefix
	}

	cmd, err := s.readCommand()

	if err != nil {
		return nil, fmt.Errorf("failed to read command due to %w", err)
	}

	message.Command = cmd

	params, err := s.readParams()

	if err != nil {
		return nil, fmt.Errorf("failed to read params due to %w", err)
	}

	message.Params = params

	return message, nil
}

var escapedMap = map[rune]rune{
	':':  ';',
	's':  ' ',
	'\\': '\\',
	'r':  '\r',
	'n':  '\n',
}

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

		if !escaped && r == ' ' {
			// Push key & value pair to map, return tags
			tags[keyBuilder.String()] = valueBuilder.String()
			return tags, nil
		}

		if !escaped && r == '=' {
			// Swap from Key to Value
			key = false
			continue
		}

		if !escaped && r == ';' {
			// Push key & value pair to map
			tags[keyBuilder.String()] = valueBuilder.String()

			keyBuilder.Reset()
			valueBuilder.Reset()
			key = true
			continue
		}

		if escaped {
			escapedRune, ok := escapedMap[r]

			if ok {
				r = escapedRune
			}
		}

		if key {
			keyBuilder.WriteRune(r)
		} else {
			valueBuilder.WriteRune(r)
		}

	}

	return tags, nil
}

// TODO: Prefix struct?
func (s *Scanner) readPrefix() (string, error) {
	var sb strings.Builder
	for {
		r := s.read()

		if r == ' ' {
			return sb.String(), nil
		}

		sb.WriteRune(r)
	}
	return "", nil
}

func (s *Scanner) readCommand() (string, error) {
	var sb strings.Builder
	for {
		r := s.read()

		if r == ' ' {
			return sb.String(), nil
		}

		sb.WriteRune(r)
	}
	return "", nil
}

func (s *Scanner) readParams() ([]string, error) {
	params := make([]string, 0)
	for {
		r := s.read()

		if r == ' ' {
			continue
		}

		if r == '\r' {
			r = s.read()

			if r == '\n' {
				fmt.Println("returning")
				return params, nil
			}
			return nil, fmt.Errorf("expected a LF after the CR but received %v", r)
		}

		if r == ':' {
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
	return nil, fmt.Errorf("reached the end of params without a CRLF")
}

// TODO: EOF checks
func (s *Scanner) readParamTrailing() (string, error) {
	var sb strings.Builder
	for {
		r := s.read()
		if r == '\r' {
			r = s.read()
			if r == '\n' {
				return sb.String(), nil
			}
			return "", fmt.Errorf("expected a LF after the CR but received %v", r)
		}

		sb.WriteRune(r)
	}
	return "", nil
}

func (s *Scanner) readParamMiddle() (string, error) {
	var sb strings.Builder
	if err := s.UnreadRune(); err != nil {
		return "", fmt.Errorf("failed to unread rune for middle param due to %w", err)
	}
	for {
		r := s.read()

		if r == ' ' {
			return sb.String(), nil
		}

		if r == ':' {
			if err := s.UnreadRune(); err != nil {
				return "", fmt.Errorf("failed to unread ':' rune due to %w", err)
			}
			return sb.String(), nil
		}
		sb.WriteRune(r)
	}
	return "", nil
}

// Reads a single rune from the Scanner
func (s *Scanner) read() rune {
	ch, _, err := s.ReadRune()

	if err != nil {
		return eof
	}
	return ch
}

func (r *Scanner) PeekRune() (rune, error) {
	for peekBytes := 4; peekBytes > 0; peekBytes-- { // unicode rune can be up to 4 bytes
		b, err := r.Peek(peekBytes)
		if err == nil {
			rune, _ := utf8.DecodeRune(b)
			if rune == utf8.RuneError {
				return rune, fmt.Errorf("Rune error")
			}
			// success
			return rune, nil
		}
		// Otherwise, we ignore Peek errors and try the next smallest number of bytes
	}

	// Pretty sure we can assume EOF if we get this far
	return eof, io.EOF
}
