package statemach

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
)

type Scanner struct {
	r    io.RuneReader
	rbuf []rune

	buf bytes.Buffer
	tok Token

	err error
}

func NewScanner(r io.Reader) *Scanner {
	return &Scanner{
		r: bufio.NewReader(r),
	}
}

func (s *Scanner) read() (rune, error) {
	if len(s.rbuf) > 0 {
		r := s.rbuf[len(s.rbuf)-1]
		s.rbuf = s.rbuf[:len(s.rbuf)-1]
		return r, nil
	}

	r, _, err := s.r.ReadRune()
	return r, err
}

func (s *Scanner) unread(r rune) {
	s.rbuf = append(s.rbuf, r)
}

func (s *Scanner) Scan() bool {
	if s.err != nil {
		return false
	}

	state := s.whitespace
	for state != nil {
		r, err := s.read()
		if err != nil {
			s.err = err

			if err != io.EOF {
				return false
			}

			r = '\n'
		}

		state = state(r)
	}

	return true
}

func (s *Scanner) Tok() Token {
	return s.tok
}

func (s *Scanner) Err() error {
	if s.err == io.EOF {
		return nil
	}

	return s.err
}

type scannerState func(r rune) scannerState

func (s *Scanner) whitespace(r rune) scannerState {
}

type Token interface {
	fmt.Stringer
	fmt.GoStringer
}
