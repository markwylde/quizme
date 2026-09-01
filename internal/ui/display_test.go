package ui

import (
	"runtime"
	"testing"
)

func TestCheckDisplayOnALocalSession(t *testing.T) {
	// The suite runs from a normal login, so a display should be reported as
	// available unless the environment says otherwise.
	if runtime.GOOS == "linux" {
		t.Setenv("DISPLAY", ":0")
	}
	t.Setenv("SSH_CONNECTION", "")
	t.Setenv("SSH_TTY", "")
	if err := CheckDisplay(); err != nil {
		t.Errorf("CheckDisplay() = %v, want nil for a local session", err)
	}
}

func TestCheckDisplayIsFast(t *testing.T) {
	// The point of the check is that it fails immediately rather than hanging,
	// so it must never block on anything.
	done := make(chan error, 1)
	go func() { done <- CheckDisplay() }()
	select {
	case <-done:
	case <-timeoutAfterASecond():
		t.Fatal("CheckDisplay blocked; it must answer immediately")
	}
}
