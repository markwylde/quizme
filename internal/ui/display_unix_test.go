//go:build linux || freebsd || openbsd || netbsd || dragonfly

package ui

import "testing"

func TestCheckDisplayWithoutAServer(t *testing.T) {
	t.Setenv("DISPLAY", "")
	t.Setenv("WAYLAND_DISPLAY", "")
	if err := CheckDisplay(); err == nil {
		t.Error("CheckDisplay() = nil, want an error when no display server is set")
	}
}

func TestCheckDisplayWithWayland(t *testing.T) {
	t.Setenv("DISPLAY", "")
	t.Setenv("WAYLAND_DISPLAY", "wayland-0")
	if err := CheckDisplay(); err != nil {
		t.Errorf("CheckDisplay() = %v, want nil under Wayland", err)
	}
}
