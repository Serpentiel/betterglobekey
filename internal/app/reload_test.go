package app

import (
	"sync"
	"testing"
)

type recordingPresser struct {
	mu       sync.Mutex
	reverses []bool
}

func (p *recordingPresser) Press(reverse bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.reverses = append(p.reverses, reverse)
}

func (p *recordingPresser) calls() []bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	return append([]bool(nil), p.reverses...)
}

func TestDaemonPressDispatchesToCurrentSwitcher(t *testing.T) {
	first := &recordingPresser{}
	second := &recordingPresser{}
	daemon := NewDaemon(first, &fakeKeyboard{}, &fakeLogger{})

	daemon.press(false)

	daemon.Reload(second)
	daemon.press(true)

	if got := first.calls(); len(got) != 1 || got[0] != false {
		t.Errorf("first switcher calls = %v, want [false]", got)
	}

	if got := second.calls(); len(got) != 1 || got[0] != true {
		t.Errorf("second switcher calls = %v, want [true]", got)
	}
}
