package statemach

import (
	"bytes"
	"unicode"
)

type Scanner struct {
	mapper func([]byte) (Token, error)

	buf bytes.Buffer
	tok Token
}

func NewScanner(mapper func([]byte) (Token, error)) *Scanner {
	return &Scanner{
		mapper: mapper,
	}
}

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

func (s *Scanner) Tok() Token {
	return s.tok
}

type Token interface{}
