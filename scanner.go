package statemach

import (
	"bytes"
	"unicode"
)

// A Scanner is a StateMachine which implements a simple, whitespace
// seperated tokenizer.
type Scanner struct {
	mapper func([]byte) (Token, error)

	buf bytes.Buffer
	tok Token
}

// NewScanner returns a new scanner which uses the given mapper
// function to convert byte slices to tokens. It works like this:
//
//     1. The user calls StateMachine.Scan() for the StateMachine that's running the Scanner.
//     2. The Scanner skips over whitespace runes.
//     3. The Scanner begins reading runes one at a time, calling mapper on the slowly growing byte slice that reading the runes results in.
//     4. If mapper returns an error, the StateMachine's error is set to that value and scanning stops.
//     5. If mapper returns a non-nil token, then the scanner's current token is set to that value and scanning stops.
func NewScanner(mapper func([]byte) (Token, error)) *Scanner {
	return &Scanner{
		mapper: mapper,
	}
}

// Config returns a config that can be used when creating a
// StateMachine that will cause the newly initialized StateMachine to
// run the Scanner.
func (s *Scanner) Config() Config {
	return Config{
		Initial: s.initial,
		EOF:     '\n',
	}
}

func (s *Scanner) initial(c Controller, r rune) StateFunc {
	if unicode.IsSpace(r) {
		return s.initial
	}

	s.buf.Reset()
	c.Unread(r)
	return s.match
}

func (s *Scanner) match(c Controller, r rune) StateFunc {
	s.buf.WriteRune(r)
	tok, err := s.mapper(s.buf.Bytes())
	if err != nil {
		c.SetErr(err)
		return nil
	}

	if tok != nil {
		s.tok = tok
		return nil
	}

	return s.match
}

// Tok returns the latest token read by the Scanner.
func (s *Scanner) Tok() Token {
	return s.tok
}

// A Token represents a value yielded by a Scanner. It is defined as
// its own type purely for self-documentation purposes.
type Token interface{}
