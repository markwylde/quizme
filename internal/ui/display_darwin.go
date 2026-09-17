package ui

import (
	"errors"
	"os"
	"syscall"
)

// CheckDisplay reports whether a desktop session is available to open a window
// in. A questionnaire that cannot be presented must fail immediately: hanging
// forever in CI or over a bare SSH connection is far worse than an error.
func CheckDisplay() error {
	if os.Getenv("SSH_CONNECTION") == "" && os.Getenv("SSH_TTY") == "" {
		return nil
	}
	// Over SSH there is only a desktop to draw on if this user also owns the
	// machine's console.
	info, err := os.Stat("/dev/console")
	if err != nil {
		return errors.New("no desktop session is available over this connection")
	}
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok || int(stat.Uid) != os.Getuid() {
		return errors.New("no desktop session is available over this connection; run quizme on the machine you are logged into")
	}
	return nil
}
