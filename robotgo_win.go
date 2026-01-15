// Copyright (c) 2016-2025 AtomAI, All rights reserved.
//
// See the COPYRIGHT file at the top-level directory of this distribution and at
// https://github.com/go-vgo/robotgo/blob/master/LICENSE
//
// Licensed under the Apache License, Version 2.0 <LICENSE-APACHE or
// http://www.apache.org/licenses/LICENSE-2.0>
//
// This file may not be copied, modified, or distributed
// except according to those terms.

//go:build windows
// +build windows

package robotgo

import (
	"syscall"
	"unsafe"

	// "github.com/lxn/win"
	"github.com/tailscale/win"
)

// DPI awareness constants
const (
	DPI_AWARENESS_CONTEXT_PER_MONITOR_AWARE_V2 = ^uintptr(3) // -4
	DPI_AWARENESS_CONTEXT_UNAWARE              = ^uintptr(0) // -1
	PROCESS_PER_MONITOR_DPI_AWARE              = 2

	// Error codes
	E_ACCESSDENIED = 0x80070005
)

// dpiAware indicates whether DPI awareness was successfully set.
// If false, the process is DPI unaware and uses logical coordinates.
var dpiAware bool

func init() {
	dpiAware = initDPIAwareness()
}

// IsDPIAware returns whether the process is DPI aware.
// If false, all coordinate APIs use logical (scaled) coordinates.
func IsDPIAware() bool {
	return dpiAware
}

// initDPIAwareness sets the process DPI awareness to Per-Monitor V2.
// This ensures GetPhysicalCursorPos and other APIs return true physical coordinates.
// Returns true if DPI awareness was successfully set, false otherwise.
func initDPIAwareness() bool {
	user32 := syscall.NewLazyDLL("user32.dll")
	shcore := syscall.NewLazyDLL("shcore.dll")

	// Check if already DPI aware
	checkDPIAware := func() bool {
		// Try GetThreadDpiAwarenessContext (Windows 10 1607+)
		if getProc := user32.NewProc("GetThreadDpiAwarenessContext"); getProc.Find() == nil {
			ctx, _, _ := getProc.Call()
			if ctx != 0 && ctx != DPI_AWARENESS_CONTEXT_UNAWARE {
				return true
			}
		}
		// Try GetProcessDpiAwareness (Windows 8.1+)
		if getProc := shcore.NewProc("GetProcessDpiAwareness"); getProc.Find() == nil {
			var awareness uint32
			ret, _, _ := getProc.Call(0, uintptr(unsafe.Pointer(&awareness)))
			if ret == 0 && awareness >= 1 { // PROCESS_SYSTEM_DPI_AWARE or higher
				return true
			}
		}
		return false
	}

	// First check: already DPI aware?
	if checkDPIAware() {
		return true
	}

	// Try SetProcessDpiAwarenessContext (Windows 10 1703+)
	if proc := user32.NewProc("SetProcessDpiAwarenessContext"); proc.Find() == nil {
		proc.Call(DPI_AWARENESS_CONTEXT_PER_MONITOR_AWARE_V2)
		if checkDPIAware() {
			return true
		}
	}

	// Try SetProcessDpiAwareness (Windows 8.1+)
	if proc := shcore.NewProc("SetProcessDpiAwareness"); proc.Find() == nil {
		proc.Call(PROCESS_PER_MONITOR_DPI_AWARE)
		if checkDPIAware() {
			return true
		}
	}

	// Try SetProcessDPIAware (Vista+)
	if proc := user32.NewProc("SetProcessDPIAware"); proc.Find() == nil {
		proc.Call()
		if checkDPIAware() {
			return true
		}
	}

	return false
}

// FindWindow find window hwnd by name
func FindWindow(name string) win.HWND {
	u1, _ := syscall.UTF16PtrFromString(name)
	hwnd := win.FindWindow(nil, u1)
	return hwnd
}

// GetHWND get foreground window hwnd
func GetHWND() win.HWND {
	hwnd := win.GetForegroundWindow()
	return hwnd
}

// SendInput send n input event
func SendInput(nInputs uint32, pInputs unsafe.Pointer, cbSize int32) uint32 {
	return win.SendInput(nInputs, pInputs, cbSize)
}

// SendMsg send a message with hwnd
func SendMsg(hwnd win.HWND, msg uint32, wParam, lParam uintptr) uintptr {
	return win.SendMessage(hwnd, msg, wParam, lParam)
}

// SetActiveWindow set window active with hwnd
func SetActiveWindow(hwnd win.HWND) win.HWND {
	return win.SetActiveWindow(hwnd)
}

// SetFocus set window focus with hwnd
func SetFocus(hwnd win.HWND) win.HWND {
	return win.SetFocus(hwnd)
}

// SetForeg set the window into the foreground by hwnd
func SetForeg(hwnd win.HWND) bool {
	return win.SetForegroundWindow(hwnd)
}

// GetMain get the main display hwnd
func GetMain() win.HWND {
	return win.GetActiveWindow()
}

// GetMainId get the main display id
func GetMainId() int {
	return int(GetMain())
}

// GetDesktopWindow get the desktop window hwnd id
func GetDesktopWindow() win.HWND {
	return win.GetDesktopWindow()
}

// GetMainDPI get the display dpi
func GetMainDPI() int {
	return int(GetDPI(GetHWND()))
}

// GetDPI get the window dpi
func GetDPI(hwnd win.HWND) uint32 {
	return win.GetDpiForWindow(hwnd)
}

// GetSysDPI get the system metrics dpi
func GetSysDPI(idx int32, dpi uint32) int32 {
	return win.GetSystemMetricsForDpi(idx, dpi)
}
