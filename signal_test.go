package inotify

import (
	"github.com/stretchr/testify/assert"
	"strings"
	"sync"
	"testing"
)

type toggleChecker struct {
	m    sync.RWMutex
	flag int
}

func (c *toggleChecker) Inc() {
	c.m.Lock()
	defer c.m.Unlock()

	c.flag++
}

func (c *toggleChecker) Reset() {
	c.m.Lock()
	defer c.m.Unlock()
	
	c.flag = 0
}

func (c *toggleChecker) GetFlag() int {
	c.m.RLock()
	defer c.m.RUnlock()

	return c.flag
}

var checker = toggleChecker {
	flag: 0,
}

func noopHandler(_ ISignal, _ ...interface{}) {
	checker.Inc()
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
	checker.Reset()
	s := NewSignal("", noopHandler)
	w := make(chan int)

	s.SendAsync(w)
	ret := <-w

	assert.Equal(t, 1, checker.GetFlag())
	assert.Equal(t, SignalExitSuccess, ret)

	s.SendAsync(w)
	s.SendAsync(nil)

	defer func() {
		<-w

		v := checker.GetFlag()
		assert.True(t, 2 <= v && 3 >= v)
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
	checker.Reset()
	s = NewSignal("", noopHandler)
	s.Send()

	assert.Equal(t, 1, checker.GetFlag())
}
