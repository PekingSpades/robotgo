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

/*
#cgo windows LDFLAGS: -lgdi32 -luser32

#include "screen/display_c.h"
*/
import "C"

import "image"

// MainDisplay returns the main display.
func MainDisplay() *Display {
	info := C.getMainDisplay()
	return &Display{
		id:     int(info.handle),
		index:  int(info.index),
		isMain: info.isMain != 0,
		bounds: Rect{
			Point: Point{X: int(info.x), Y: int(info.y)},
			Size:  Size{W: int(info.w), H: int(info.h)},
		},
		virtualOrigin: Point{X: int(info.vx), Y: int(info.vy)},
		scale:         float64(info.scale),
	}
}

// AllDisplays returns all displays.
func AllDisplays() []*Display {
	count := int(C.getDisplayCount())
	if count <= 0 {
		return nil
	}

	var cDisplays [32]C.DisplayInfoC
	actualCount := int(C.getAllDisplays(&cDisplays[0], C.int32_t(32)))

	displays := make([]*Display, actualCount)
	for i := 0; i < actualCount; i++ {
		info := cDisplays[i]
		displays[i] = &Display{
			id:     int(info.handle),
			index:  int(info.index),
			isMain: info.isMain != 0,
			bounds: Rect{
				Point: Point{X: int(info.x), Y: int(info.y)},
				Size:  Size{W: int(info.w), H: int(info.h)},
			},
			virtualOrigin: Point{X: int(info.vx), Y: int(info.vy)},
			scale:         float64(info.scale),
		}
	}

	return displays
}

// DisplayAt returns the display at the specified index.
// Returns nil if the index is invalid.
func DisplayAt(index int) *Display {
	if index < 0 {
		return nil
	}

	info := C.getDisplayAt(C.int32_t(index))
	if info.w == 0 && info.h == 0 {
		return nil
	}

	return &Display{
		id:     int(info.handle),
		index:  int(info.index),
		isMain: info.isMain != 0,
		bounds: Rect{
			Point: Point{X: int(info.x), Y: int(info.y)},
			Size:  Size{W: int(info.w), H: int(info.h)},
		},
		virtualOrigin: Point{X: int(info.vx), Y: int(info.vy)},
		scale:         float64(info.scale),
	}
}

// DisplayCount returns the number of displays.
func DisplayCount() int {
	return int(C.getDisplayCount())
}

// virtualWidth returns the virtual (logical) width of the display.
func (d *Display) virtualWidth() int {
	if d.scale > 0 {
		return int(float64(d.bounds.W) / d.scale)
	}
	return d.bounds.W
}

// virtualHeight returns the virtual (logical) height of the display.
func (d *Display) virtualHeight() int {
	if d.scale > 0 {
		return int(float64(d.bounds.H) / d.scale)
	}
	return d.bounds.H
}

// ToAbsolute converts physical pixel coordinates relative to this display
// to absolute coordinates suitable for mouse APIs.
// When DPI awareness is active: returns physical absolute coordinates
// When DPI awareness fails: returns virtual absolute coordinates
func (d *Display) ToAbsolute(physX, physY int) (absX, absY int) {
	if dpiAware {
		// DPI aware: use physical coordinates directly
		return d.bounds.X + physX, d.bounds.Y + physY
	}
	// DPI unaware: convert physical to virtual coordinates
	// Mouse APIs use virtual coordinates when DPI awareness fails
	if d.scale > 0 {
		absX = d.virtualOrigin.X + int(float64(physX)/d.scale)
		absY = d.virtualOrigin.Y + int(float64(physY)/d.scale)
	} else {
		absX = d.virtualOrigin.X + physX
		absY = d.virtualOrigin.Y + physY
	}
	return
}

// ToRelative converts absolute coordinates from mouse APIs to physical pixel
// coordinates relative to this display.
// Returns (0, 0, false) if the coordinate is not within this display.
func (d *Display) ToRelative(absX, absY int) (physX, physY int, ok bool) {
	if !d.Contains(absX, absY) {
		return 0, 0, false
	}
	if dpiAware {
		// DPI aware: coordinates are physical
		return absX - d.bounds.X, absY - d.bounds.Y, true
	}
	// DPI unaware: coordinates are virtual, convert to physical
	virtRelX := absX - d.virtualOrigin.X
	virtRelY := absY - d.virtualOrigin.Y
	if d.scale > 0 {
		physX = int(float64(virtRelX) * d.scale)
		physY = int(float64(virtRelY) * d.scale)
	} else {
		physX = virtRelX
		physY = virtRelY
	}
	return physX, physY, true
}

// Contains checks if the specified absolute coordinate is within this display.
// The coordinate system depends on DPI awareness state.
func (d *Display) Contains(absX, absY int) bool {
	if dpiAware {
		// DPI aware: use physical coordinates
		return absX >= d.bounds.X && absX < d.bounds.X+d.bounds.W &&
			absY >= d.bounds.Y && absY < d.bounds.Y+d.bounds.H
	}
	// DPI unaware: use virtual coordinates
	virtW := d.virtualWidth()
	virtH := d.virtualHeight()
	return absX >= d.virtualOrigin.X && absX < d.virtualOrigin.X+virtW &&
		absY >= d.virtualOrigin.Y && absY < d.virtualOrigin.Y+virtH
}

// Move moves the mouse to the specified coordinates relative to this display.
// Coordinates are in physical pixels.
func (d *Display) Move(x, y int) {
	absX, absY := d.ToAbsolute(x, y)
	Move(absX, absY)
}

// MoveSmooth smoothly moves the mouse to the specified coordinates relative to this display.
// Coordinates are in physical pixels.
func (d *Display) MoveSmooth(x, y int, args ...interface{}) bool {
	absX, absY := d.ToAbsolute(x, y)
	return MoveSmooth(absX, absY, args...)
}

// Drag drags the mouse from one position to another on this display.
// Coordinates are in physical pixels relative to this display.
func (d *Display) Drag(fromX, fromY, toX, toY int, button string) {
	d.Move(fromX, fromY)
	Toggle(button)
	MilliSleep(50)
	d.MoveSmooth(toX, toY)
	Toggle(button, "up")
}

// DragTo drags the mouse from the current position to the specified position on this display.
// Coordinates are in physical pixels relative to this display.
func (d *Display) DragTo(x, y int, button string) {
	Toggle(button)
	MilliSleep(50)
	d.MoveSmooth(x, y)
	Toggle(button, "up")
}

// Capture captures the entire display.
func (d *Display) Capture() (*image.RGBA, error) {
	return d.CaptureRect(0, 0, d.bounds.W, d.bounds.H)
}

// CaptureRect captures a rectangular region of this display.
// Coordinates and size are in physical pixels relative to this display.
func (d *Display) CaptureRect(physX, physY, w, h int) (*image.RGBA, error) {
	absX, absY := d.ToAbsolute(physX, physY)
	if dpiAware {
		return Capture(absX, absY, w, h)
	}
	// DPI unaware: convert physical size to virtual size
	virtW := w
	virtH := h
	if d.scale > 0 {
		virtW = int(float64(w) / d.scale)
		virtH = int(float64(h) / d.scale)
	}
	return Capture(absX, absY, virtW, virtH)
}

// GetPixelColor gets the pixel color at the specified coordinates relative to this display.
// Returns the color as a hex string "RRGGBB".
// Coordinates are in physical pixels.
func (d *Display) GetPixelColor(x, y int) string {
	absX, absY := d.ToAbsolute(x, y)
	return GetPixelColor(absX, absY)
}

// MouseLocation gets the mouse location relative to this display.
// Returns (-1, -1, false) if the mouse is not on this display.
// Coordinates are in physical pixels.
func (d *Display) MouseLocation() (x, y int, ok bool) {
	absX, absY := Location()
	return d.ToRelative(absX, absY)
}

// ContainsMouse checks if the mouse is on this display.
func (d *Display) ContainsMouse() bool {
	absX, absY := Location()
	return d.Contains(absX, absY)
}
