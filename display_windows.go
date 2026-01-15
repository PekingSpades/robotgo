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

// WindowsPlatformInfo contains Windows-specific display information.
// This struct stores physical coordinates that are always available
// regardless of DPI awareness mode.
type WindowsPlatformInfo struct {
	// PhysicalOrigin is the top-left corner in physical pixel coordinates.
	// This is always the physical position, regardless of DPI awareness mode.
	// Use this for screen capture operations which always use physical coordinates.
	PhysicalOrigin Point
}

// Platform returns the platform name.
func (w *WindowsPlatformInfo) Platform() string {
	return "windows"
}

// MainDisplay returns the main display.
func MainDisplay() *Display {
	info := C.getMainDisplay()
	scale := float64(info.scale)
	physW, physH := int(info.w), int(info.h)

	// Calculate logical size
	logicalW, logicalH := physW, physH
	if scale > 0 {
		logicalW = int(float64(physW) / scale)
		logicalH = int(float64(physH) / scale)
	}

	// Always store physical origin for capture operations
	physicalOrigin := Point{X: int(info.x), Y: int(info.y)}

	var origin Rect
	if dpiAware {
		// DPI aware: use physical coordinates and size
		origin = Rect{
			Point: Point{X: int(info.x), Y: int(info.y)},
			Size:  Size{W: physW, H: physH},
		}
	} else {
		// DPI unaware: use virtual coordinates and logical size
		origin = Rect{
			Point: Point{X: int(info.vx), Y: int(info.vy)},
			Size:  Size{W: logicalW, H: logicalH},
		}
	}

	return &Display{
		id:     int(info.handle),
		index:  int(info.index),
		isMain: info.isMain != 0,
		origin: origin,
		size:   Size{W: physW, H: physH},
		scale:  scale,
		platform: &WindowsPlatformInfo{
			PhysicalOrigin: physicalOrigin,
		},
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
		scale := float64(info.scale)
		physW, physH := int(info.w), int(info.h)

		logicalW, logicalH := physW, physH
		if scale > 0 {
			logicalW = int(float64(physW) / scale)
			logicalH = int(float64(physH) / scale)
		}

		// Always store physical origin for capture operations
		physicalOrigin := Point{X: int(info.x), Y: int(info.y)}

		var origin Rect
		if dpiAware {
			origin = Rect{
				Point: Point{X: int(info.x), Y: int(info.y)},
				Size:  Size{W: physW, H: physH},
			}
		} else {
			origin = Rect{
				Point: Point{X: int(info.vx), Y: int(info.vy)},
				Size:  Size{W: logicalW, H: logicalH},
			}
		}

		displays[i] = &Display{
			id:     int(info.handle),
			index:  int(info.index),
			isMain: info.isMain != 0,
			origin: origin,
			size:   Size{W: physW, H: physH},
			scale:  scale,
			platform: &WindowsPlatformInfo{
				PhysicalOrigin: physicalOrigin,
			},
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

	scale := float64(info.scale)
	physW, physH := int(info.w), int(info.h)

	logicalW, logicalH := physW, physH
	if scale > 0 {
		logicalW = int(float64(physW) / scale)
		logicalH = int(float64(physH) / scale)
	}

	// Always store physical origin for capture operations
	physicalOrigin := Point{X: int(info.x), Y: int(info.y)}

	var origin Rect
	if dpiAware {
		origin = Rect{
			Point: Point{X: int(info.x), Y: int(info.y)},
			Size:  Size{W: physW, H: physH},
		}
	} else {
		origin = Rect{
			Point: Point{X: int(info.vx), Y: int(info.vy)},
			Size:  Size{W: logicalW, H: logicalH},
		}
	}

	return &Display{
		id:     int(info.handle),
		index:  int(info.index),
		isMain: info.isMain != 0,
		origin: origin,
		size:   Size{W: physW, H: physH},
		scale:  scale,
		platform: &WindowsPlatformInfo{
			PhysicalOrigin: physicalOrigin,
		},
	}
}

// DisplayCount returns the number of displays.
func DisplayCount() int {
	return int(C.getDisplayCount())
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
	return absX >= d.origin.X && absX < d.origin.X+d.origin.W &&
		absY >= d.origin.Y && absY < d.origin.Y+d.origin.H
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
// Always uses physical coordinates for capture, regardless of DPI awareness mode.
func (d *Display) CaptureRect(physX, physY, w, h int) (*image.RGBA, error) {
	pi, ok := d.platform.(*WindowsPlatformInfo)
	if !ok {
		return nil, ErrInvalidPlatformInfo
	}
	absX := pi.PhysicalOrigin.X + physX
	absY := pi.PhysicalOrigin.Y + physY
	return Capture(absX, absY, w, h)
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
