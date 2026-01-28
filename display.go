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

package robotgo

import (
	"errors"
)

// ErrCaptureScreen is returned when screen capture fails.
var ErrCaptureScreen = errors.New("capture screen failed")

// ErrInvalidPlatformInfo is returned when platform info is missing or invalid.
var ErrInvalidPlatformInfo = errors.New("invalid platform info")

// PlatformInfo is an interface for platform-specific display information.
// Each platform can implement this interface to store custom data.
type PlatformInfo interface {
	// Platform returns the platform name (e.g., "windows", "darwin", "linux")
	Platform() string
}

// Display represents a physical display/monitor.
//
// Coordinate semantics:
//   - origin: Bounds in platform's coordinate system (position + logical size)
//   - size: Physical pixel resolution
//   - Move/Capture/etc: Accept physical pixel coordinates relative to this display
type Display struct {
	id       int          // Platform-specific display ID
	index    int          // Display index (0 is main display)
	isMain   bool         // Whether this is the main display
	origin   Rect         // Bounds in platform's coordinate system (position + logical size)
	size     Size         // Physical pixel size
	scale    float64      // Scale factor (physical pixels / virtual points)
	platform PlatformInfo // Platform-specific information (nil if not set)
}

// DisplayInfo contains detailed information about a display.
type DisplayInfo struct {
	ID          int     // Platform-specific ID
	Index       int     // Index
	IsMain      bool    // Whether this is the main display
	Origin      Rect    // Bounds in platform's coordinate system
	Size        Size    // Physical pixel size
	ScaleFactor float64 // Scale factor
}

// ID returns the platform-specific display identifier.
//   - Windows: HMONITOR handle
//   - macOS: CGDirectDisplayID
//   - Linux: Xinerama screen index
func (d *Display) ID() int {
	return d.id
}

// Index returns the display index (0 is the main display).
func (d *Display) Index() int {
	return d.index
}

// IsMain returns whether this is the main display.
func (d *Display) IsMain() bool {
	return d.isMain
}

// Origin returns the display bounds in platform's coordinate system.
// X/Y: Position, W/H: Logical size (for mouse operations)
func (d *Display) Origin() Rect {
	return d.origin
}

// Size returns the display physical pixel size.
func (d *Display) Size() Size {
	return d.size
}

// Width returns the display physical pixel width.
func (d *Display) Width() int {
	return d.size.W
}

// Height returns the display physical pixel height.
func (d *Display) Height() int {
	return d.size.H
}

// Scale returns the scale factor (physical pixels / virtual points).
//   - Windows: DPI / 96.0
//   - macOS: PixelWidth / VirtualWidth
//   - Linux: Xft.dpi / 96.0
func (d *Display) Scale() float64 {
	return d.scale
}

// GetPlatformInfo returns the platform-specific information.
// Returns nil if no platform-specific info is set.
// Use type assertion to access platform-specific fields:
//
//	if info, ok := d.GetPlatformInfo().(*WindowsPlatformInfo); ok {
//	    // access Windows-specific fields
//	}
func (d *Display) GetPlatformInfo() PlatformInfo {
	return d.platform
}

// Info returns detailed display information.
func (d *Display) Info() DisplayInfo {
	return DisplayInfo{
		ID:          d.id,
		Index:       d.index,
		IsMain:      d.isMain,
		Origin:      d.origin,
		Size:        d.size,
		ScaleFactor: d.scale,
	}
}

// Note: The following methods are platform-specific and defined in:
// - display_darwin.go (macOS): Converts physical coordinates to virtual before calling APIs
// - display_windows.go (Windows): Uses physical coordinates directly
// - display_linux.go (Linux): Uses physical coordinates directly
//
// Platform-specific methods:
// - ToAbsolute(x, y int) (absX, absY int)
// - ToRelative(absX, absY int) (x, y int, ok bool)
// - Contains(absX, absY int) bool
// - Move(x, y int)
// - MoveSmooth(x, y int, args ...interface{}) bool
// - Drag(fromX, fromY, toX, toY int, button string)
// - DragTo(x, y int, button string)
// - Capture() (*image.RGBA, error)
// - CaptureRect(x, y, w, h int) (*image.RGBA, error)
// - GetPixelColor(x, y int) string
// - MouseLocation() (x, y int, ok bool)
// - ContainsMouse() bool
