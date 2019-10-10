package inotify

import (
	"reflect"
	"runtime"
	"sync"

	log "github.com/sirupsen/logrus"
)

const (
	SignalExitSuccess = 0
	SignalExitFailure = 1
)

type SignalHandler func(ISignal, ...interface{})

type ISignal interface {
	Send(args ...interface{})
	SendAsync(wait chan int, args ...interface{})
	Connect(handler SignalHandler)
}

type Signal struct {
	name  string
	mutex sync.Mutex

	// Use map to avoid duplications.
	receivers map[string]SignalHandler
	numRecv   int
}

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

func (s *Signal) Name() string {
	return s.name
}

func (s *Signal) Connect(handler SignalHandler) {
	s.mutex.Lock()

	n := s.getHandlerName(handler)
	s.receivers[n] = handler
	s.numRecv = len(s.receivers)

	s.mutex.Unlock()
}

func (s *Signal) Send(args ...interface{}) {
	s.mutex.Lock()
	defer s.recover(s.name, nil)

	for _, h := range s.receivers {
		h(s, args...)
	}
	s.mutex.Unlock()
}

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
		log.Errorf("Recovered in %s: %s", n, r)

		if nil != wait {
			wait <- SignalExitFailure
		}
	}
}
