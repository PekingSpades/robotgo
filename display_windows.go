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
	origin := Point{X: int(info.x), Y: int(info.y)}
	if !dpiAware {
		// When DPI unaware, use virtual coordinates as origin
		origin = Point{X: int(info.vx), Y: int(info.vy)}
	}
	return &Display{
		id:     int(info.handle),
		index:  int(info.index),
		isMain: info.isMain != 0,
		origin: origin,
		size:   Size{W: int(info.w), H: int(info.h)},
		scale:  float64(info.scale),
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
		origin := Point{X: int(info.x), Y: int(info.y)}
		if !dpiAware {
			origin = Point{X: int(info.vx), Y: int(info.vy)}
		}
		displays[i] = &Display{
			id:     int(info.handle),
			index:  int(info.index),
			isMain: info.isMain != 0,
			origin: origin,
			size:   Size{W: int(info.w), H: int(info.h)},
			scale:  float64(info.scale),
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

	origin := Point{X: int(info.x), Y: int(info.y)}
	if !dpiAware {
		origin = Point{X: int(info.vx), Y: int(info.vy)}
	}
	return &Display{
		id:     int(info.handle),
		index:  int(info.index),
		isMain: info.isMain != 0,
		origin: origin,
		size:   Size{W: int(info.w), H: int(info.h)},
		scale:  float64(info.scale),
	}
}

// DisplayCount returns the number of displays.
func DisplayCount() int {
	return int(C.getDisplayCount())
}

// virtualWidth returns the virtual (logical) width of the display.
func (d *Display) virtualWidth() int {
	if d.scale > 0 {
		return int(float64(d.size.W) / d.scale)
	}
	return d.size.W
}

// virtualHeight returns the virtual (logical) height of the display.
func (d *Display) virtualHeight() int {
	if d.scale > 0 {
		return int(float64(d.size.H) / d.scale)
	}
	return d.size.H
}

// ToAbsolute converts physical pixel coordinates relative to this display
// to absolute coordinates suitable for mouse APIs.
func (d *Display) ToAbsolute(physX, physY int) (absX, absY int) {
	if dpiAware {
		// DPI aware: origin is physical, coordinates are physical
		return d.origin.X + physX, d.origin.Y + physY
	}
	// DPI unaware: origin is virtual, convert physical to virtual
	if d.scale > 0 {
		absX = d.origin.X + int(float64(physX)/d.scale)
		absY = d.origin.Y + int(float64(physY)/d.scale)
	} else {
		absX = d.origin.X + physX
		absY = d.origin.Y + physY
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
		return absX - d.origin.X, absY - d.origin.Y, true
	}
	// DPI unaware: coordinates are virtual, convert to physical
	virtRelX := absX - d.origin.X
	virtRelY := absY - d.origin.Y
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
func (d *Display) Contains(absX, absY int) bool {
	if dpiAware {
		// DPI aware: use physical size
		return absX >= d.origin.X && absX < d.origin.X+d.size.W &&
			absY >= d.origin.Y && absY < d.origin.Y+d.size.H
	}
	// DPI unaware: use virtual size
	virtW := d.virtualWidth()
	virtH := d.virtualHeight()
	return absX >= d.origin.X && absX < d.origin.X+virtW &&
		absY >= d.origin.Y && absY < d.origin.Y+virtH
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
	return d.CaptureRect(0, 0, d.size.W, d.size.H)
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
