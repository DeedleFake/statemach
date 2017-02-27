// Package statemach defines a number of helpers for creating state
// machine implementations.
package statemach

import (
	"bufio"
	"fmt"
	"io"
	"reflect"
)

// A StateMachine is a simplistic implementation of a DFA-style
// automaton. Since this is Go, however, it isn't limited by the
// traditional DFA definition. For example, a StateMachine, depending
// on implementation, can have memory, and the states themselves can
// have side effects.
type StateMachine struct {
	r    io.RuneReader
	rbuf []rune
	err  error

	mach Config
}

// New returns a new StateMachine which reads from the given source
// and uses the given config.
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

// Run runs a single iteration of a StateMachine. It starts at the
// initial state specified in the machines configuration and
// repeatedly calls state functions until one of them returns nil or
// an error occurs.
//
// Note that io.EOF is a special-case error. See the documentation for
// Config for more information.
//
// Run returns true if more iterations can be run. In other words, if the machine was stopped because a state told it too, rather than because of an error.
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

			case nil:
				return false

			default:
				panic(fmt.Errorf("Unexpected EOF handler type: %q", reflect.TypeOf(n)))
			}
		}

		state = state(&controller{sm}, r)
	}

	return sm.err == nil
}

// Err returns the error that stopped the StateMachine. If that error was io.EOF, Err returns nil.
func (sm *StateMachine) Err() error {
	if sm.err == io.EOF {
		return nil
	}

	return sm.err
}

// A Config defines a StateMachine's behavior.
type Config struct {
	// Initial specifies the state that the StateMachine should start in
	// whenever Run is called.
	Initial StateFunc

	// EOF defines how the StateMachine should handle an io.EOF error
	// while reading from the underlying input. When an io.EOF is
	// encountered, the behavior of the machine is dependent on the
	// type of EOF:
	//
	// If EOF is of type EOFFunc, then that function is called. If
	// that function returns a non-nil StateFunc, then that returned
	// function becomes the next state. Note that this will likely
	// call EOF again if no runes were unread. If the function
	// returned is nil, then the StateMachine will stop immediately.
	//
	// If EOF is of type rune, then the current StateFunc of the
	// StateMachine will be called with that rune as its argument. In
	// this case, a StateFunc must eventually return a nil StateFunc
	// in order to stop the machine.
	//
	// If EOF is nil, then the machine will stop immediately.
	//
	// All other types will panic.
	EOF interface{}
}

// A StateFunc represents a state in the StateMachine. As the machine
// reads runes from the underlying input, it calls states repeatedly.
// After each state, the state returned by the StateFunc becomes the
// next state. If a StateFunc returns nil, then the StateMachine
// stops.
type StateFunc func(Controller, rune) StateFunc

// An EOFFunc is a pseudo-state that is called when a StateMachine
// encouters an io.EOF. For more information, see the documentation
// for Config.
type EOFFunc func(Controller) StateFunc

// Controller gives a limited, specialized interface that allows some
// control of parts of a StateMachine from inside state functions.
type Controller interface {
	// Unread pushes a rune onto a stack, causing it to be the next rune
	// read by the StateMachine instead of whatever would have been read
	// from the underlying input.
	Unread(rune)

	// SetErr sets the StateMachine's error value. This does not stop
	// the current StateMachine run, but it will prevent future runs,
	// and it will cause the Err method to return the value given to it.
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
