//go:build linux || freebsd || openbsd || netbsd || dragonfly

package ui

import (
	"errors"
	"os"
)

// CheckDisplay reports whether a desktop session is available to open a window
// in. A questionnaire that cannot be presented must fail immediately: hanging
// forever in CI or over a bare SSH connection is far worse than an error.
func CheckDisplay() error {
	if os.Getenv("DISPLAY") != "" || os.Getenv("WAYLAND_DISPLAY") != "" {
		return nil
	}
	return errors.New("no desktop session is available: neither DISPLAY nor WAYLAND_DISPLAY is set")
}
