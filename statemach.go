package statemach

import (
	"bufio"
	"fmt"
	"io"
	"reflect"
)

type StateMachine struct {
	r    io.RuneReader
	rbuf []rune
	err  error

	mach Config
}

func New(r io.Reader, mach Config) *StateMachine {
	return &StateMachine{
		r:    bufio.NewReader(r),
		mach: mach,
	}
}

func (sm *StateMachine) read() (r rune, err error) {
	if len(sm.rbuf) > 0 {
		r = sm.rbuf[len(sm.rbuf)-1]
		sm.rbuf = sm.rbuf[:len(sm.rbuf)-1]
		return
	}

	r, _, err = sm.r.ReadRune()
	return
}

func (sm *StateMachine) unread(r rune) {
	sm.rbuf = append(sm.rbuf, r)
}

func (sm *StateMachine) Run() bool {
	if sm.err != nil {
		return false
	}

	state := sm.mach.Initial
	for state != nil {
		r, err := sm.read()
		if err != nil {
			if sm.err != io.EOF {
				sm.err = err
				return false
			}

			switch n := sm.mach.EOF.(type) {
			case EOFFunc:
				state = n(&controller{sm})
				if sm.err != nil {
					return false
				}
				continue

			case rune:
				sm.err = err
				r = n

			default:
				panic(fmt.Errorf("Unexpected EOF handler type: %q", reflect.TypeOf(n)))
			}
		}

		state = state(&controller{sm}, r)
	}

	return true
}

func (sm *StateMachine) Err() error {
	return sm.err
}

type Config struct {
	Initial StateFunc
	EOF     interface{}
}

type StateFunc func(Controller, rune) StateFunc

type EOFFunc func(Controller) StateFunc

type Controller interface {
	Unread(rune)
	SetErr(error)
}

type controller struct {
	*StateMachine
}

func (c *controller) Read() (r rune, err error) {
	return c.read()
}

func (c *controller) Unread(r rune) {
	c.unread(r)
}

func (c *controller) SetErr(err error) {
	c.err = err
}
