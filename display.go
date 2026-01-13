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

import "errors"

// ErrCaptureScreen is returned when screen capture fails.
var ErrCaptureScreen = errors.New("capture screen failed")

// Display represents a physical display/monitor.
//
// Coordinate semantics differ by platform:
//   - Windows: Physical pixel coordinates
//   - macOS: Virtual (scaled) coordinates
//   - Linux: Physical pixel coordinates
type Display struct {
	id     int     // Platform-specific display ID
	index  int     // Display index (0 is main display)
	isMain bool    // Whether this is the main display
	bounds Rect    // Display bounds (relative to virtual screen)
	scale  float64 // Scale factor (for informational purposes only)
}

// DisplayInfo contains detailed information about a display.
type DisplayInfo struct {
	ID          int     // Platform-specific ID
	Index       int     // Index
	IsMain      bool    // Whether this is the main display
	Bounds      Rect    // Bounds
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

// Bounds returns the display bounds (relative to virtual screen coordinate system).
func (d *Display) Bounds() Rect {
	return d.bounds
}

// Size returns the display size.
func (d *Display) Size() Size {
	return d.bounds.Size
}

// Width returns the display width.
func (d *Display) Width() int {
	return d.bounds.W
}

// Height returns the display height.
func (d *Display) Height() int {
	return d.bounds.H
}

// Scale returns the scale factor (for informational purposes only).
//   - Windows: DPI / 96.0
//   - macOS: PixelWidth / VirtualWidth
//   - Linux: Xft.dpi / 96.0
func (d *Display) Scale() float64 {
	return d.scale
}

// Info returns detailed display information.
func (d *Display) Info() DisplayInfo {
	return DisplayInfo{
		ID:          d.id,
		Index:       d.index,
		IsMain:      d.isMain,
		Bounds:      d.bounds,
		ScaleFactor: d.scale,
	}
}

// ToAbsolute converts coordinates relative to this display to absolute coordinates.
func (d *Display) ToAbsolute(x, y int) (absX, absY int) {
	return d.bounds.X + x, d.bounds.Y + y
}

// ToRelative converts absolute coordinates to coordinates relative to this display.
// Returns (0, 0, false) if the coordinate is not within this display.
func (d *Display) ToRelative(absX, absY int) (x, y int, ok bool) {
	if !d.Contains(absX, absY) {
		return 0, 0, false
	}
	return absX - d.bounds.X, absY - d.bounds.Y, true
}

// Contains checks if the specified absolute coordinate is within this display.
func (d *Display) Contains(absX, absY int) bool {
	return absX >= d.bounds.X && absX < d.bounds.X+d.bounds.W &&
		absY >= d.bounds.Y && absY < d.bounds.Y+d.bounds.H
}
