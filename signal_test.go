package inotify

import (
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

var toggleChecker = struct {
	Mutex sync.Mutex
	Flag  int
} {
	Mutex: sync.Mutex{},
	Flag: 0,
}

func noopHandler(_ ISignal, _ ...interface{}) {
	toggleChecker.Mutex.Lock()
	toggleChecker.Flag += 1
	toggleChecker.Mutex.Unlock()
}

func panicHandler(_ ISignal, _ ...interface{}) {
	panic("error!")
}

func TestNewSignal(t *testing.T) {
	s := NewSignal("fake-signal", noopHandler)

	assert.Equal(t, s.numRecv, 1)
	assert.Equal(t, 1, len(s.receivers))

	s2 := NewSignal("empty-signal")
	assert.Equal(t, 0, s2.numRecv)
}

func TestSignal_Connect(t *testing.T) {
	s := NewSignal("fake-signal")
	s.Connect(noopHandler)

	assert.Equal(t, 1, s.numRecv)
	assert.Equal(t, 1, len(s.receivers))

	s.Connect(noopHandler)
	assert.Equal(t, 1, s.numRecv)
	assert.Equal(t, 1, len(s.receivers))
}

func TestSignal_getHandlerName(t *testing.T) {
	s := NewSignal("")
	name := s.getHandlerName(noopHandler)

	sep := strings.Split(name, ".")
	if 2 > len(sep) {
		t.Error("bad handler name")
	}

	name = sep[len(sep)-1] // Get the last one as the function name.

	assert.Equal(t, "noopHandler", name)
}

func TestSignal_SendAsync(t *testing.T) {
	toggleChecker.Flag = 0
	s := NewSignal("", noopHandler)
	w := make(chan int)

	s.SendAsync(w)
	ret := <-w

	assert.Equal(t, 1, toggleChecker.Flag)
	assert.Equal(t, SignalExitSuccess, ret)

	s.SendAsync(w)
	s.SendAsync(nil)

	defer func() {
		<-w
		assert.True(t, 2 <= toggleChecker.Flag && 3 >= toggleChecker.Flag)
	}()
}

func TestSignal_SendPanicAsync(t *testing.T) {
	s := NewSignal("", panicHandler)
	w := make(chan int)

	s.SendAsync(w)
	ret := <-w

	assert.Equal(t, SignalExitFailure, ret)
}

func TestSignal_Send(t *testing.T) {
	// Test recovering from panic.
	s := NewSignal("", panicHandler)
	s.Send()

	// Test normal signal calls.
	toggleChecker.Flag = 0
	s = NewSignal("", noopHandler)
	s.Send()

	assert.Equal(t, 1, toggleChecker.Flag)
}
