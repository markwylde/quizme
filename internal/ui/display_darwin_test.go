package ui

import (
	"strings"
	"testing"
)

func TestCheckDisplayOverSSHFromAnotherUser(t *testing.T) {
	// /dev/console is owned by whoever is sat at the machine. Over SSH as a
	// different user there is no desktop to draw on, and the command must say
	// so rather than hang waiting for a window that never appears.
	t.Setenv("SSH_CONNECTION", "10.0.0.1 22 10.0.0.2 22")
	err := CheckDisplay()
	if err == nil {
		t.Skip("this user owns the console, so an SSH session can still open a window")
	}
	if !strings.Contains(err.Error(), "no desktop session") {
		t.Errorf("error = %q, want it to explain there is no desktop session", err)
	}
}
