package inotify

import (
	"reflect"
	"runtime"
	"sync"
)

const (
	// SignalExitSuccess returns when handler function has exited successfully.
	SignalExitSuccess = 0

	// SignalExitFailure returns when handler function raises a panic.
	SignalExitFailure = 1
)

// SignalHandler type is the abstract handler function.
type SignalHandler func(ISignal, ...interface{})

// ISignal is the abstract Signal "class".
type ISignal interface {
	Name() string
	Send(args ...interface{})
	SendAsync(wait chan int, args ...interface{})
	Connect(handler SignalHandler)
}

// Signal struct holds its receivers.
type Signal struct {
	name  string
	mutex sync.Mutex

	// Use map to avoid duplications.
	receivers map[string]SignalHandler
	numRecv   int
}

// NewSignal initializes one new signal instance.
func NewSignal(name string, handlers ...SignalHandler) *Signal {
	s := Signal{
		name:      name,
		receivers: make(map[string]SignalHandler),
		numRecv:   0,
	}

	for _, h := range handlers {
		name := s.getHandlerName(h)
		s.receivers[name] = h
	}

	s.numRecv = len(s.receivers)

	return &s
}

// Name returns the signal name.
func (s *Signal) Name() string {
	return s.name
}

// Connect appends one signal handler.
func (s *Signal) Connect(handler SignalHandler) {
	s.mutex.Lock()

	n := s.getHandlerName(handler)
	s.receivers[n] = handler
	s.numRecv = len(s.receivers)

	s.mutex.Unlock()
}

// Send calls each handler one-by-one.
func (s *Signal) Send(args ...interface{}) {
	s.mutex.Lock()
	defer s.recover(s.name, nil)

	for _, h := range s.receivers {
		h(s, args...)
	}
	s.mutex.Unlock()
}

// SendAsync calls handlers asynchonosly.
func (s *Signal) SendAsync(wait chan int, args ...interface{}) {
	s.mutex.Lock()
	for n, h := range s.receivers {
		go func() {
			defer s.recover(n, wait)

			h(s, args...)
			if nil != wait {
				wait <- SignalExitSuccess
			}
		}()
	}
	s.mutex.Unlock()
}

func (s *Signal) getHandlerName(h SignalHandler) string {
	return runtime.FuncForPC(reflect.ValueOf(h).Pointer()).Name()
}

func (s *Signal) recover(n string, wait chan int) {
	if "" == n {
		n = "anonymous signal"
	}

	if r := recover(); nil != r {
		if nil != wait {
			wait <- SignalExitFailure
		}
	}
}
