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

//go:build darwin
// +build darwin

package robotgo

/*
#cgo darwin CFLAGS: -x objective-c -Wno-deprecated-declarations
#cgo darwin LDFLAGS: -framework Cocoa -framework CoreFoundation -framework IOKit
#cgo darwin LDFLAGS: -framework Carbon -framework OpenGL
#cgo darwin LDFLAGS: -weak_framework ScreenCaptureKit

#include "screen/display_c.h"
*/
import "C"

import "image"

// MainDisplay returns the main display.
func MainDisplay() *Display {
	info := C.getMainDisplay()
	scale := float64(info.scale)
	physW, physH := int(info.w), int(info.h)

	// Calculate logical size for origin
	logicalW, logicalH := physW, physH
	if scale > 0 {
		logicalW = int(float64(physW) / scale)
		logicalH = int(float64(physH) / scale)
	}

	return &Display{
		id:     int(info.handle),
		index:  int(info.index),
		isMain: info.isMain != 0,
		origin: Rect{
			Point: Point{X: int(info.x), Y: int(info.y)},
			Size:  Size{W: logicalW, H: logicalH},
		},
		size:  Size{W: physW, H: physH},
		scale: scale,
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

		displays[i] = &Display{
			id:     int(info.handle),
			index:  int(info.index),
			isMain: info.isMain != 0,
			origin: Rect{
				Point: Point{X: int(info.x), Y: int(info.y)},
				Size:  Size{W: logicalW, H: logicalH},
			},
			size:  Size{W: physW, H: physH},
			scale: scale,
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

	return &Display{
		id:     int(info.handle),
		index:  int(info.index),
		isMain: info.isMain != 0,
		origin: Rect{
			Point: Point{X: int(info.x), Y: int(info.y)},
			Size:  Size{W: logicalW, H: logicalH},
		},
		size:  Size{W: physW, H: physH},
		scale: scale,
	}
}

// DisplayCount returns the number of displays.
func DisplayCount() int {
	return int(C.getDisplayCount())
}

// ToAbsolute converts physical pixel coordinates relative to this display
// to virtual absolute coordinates (for use with macOS APIs).
// Input: physical pixel coordinates (0,0 is top-left of this display)
// Output: virtual absolute coordinates in the virtual desktop
func (d *Display) ToAbsolute(physX, physY int) (virtAbsX, virtAbsY int) {
	if d.scale > 0 {
		// Convert physical to virtual relative, then add virtual origin
		virtAbsX = d.origin.X + int(float64(physX)/d.scale)
		virtAbsY = d.origin.Y + int(float64(physY)/d.scale)
	} else {
		virtAbsX = d.origin.X + physX
		virtAbsY = d.origin.Y + physY
	}
	return
}

// ToRelative converts virtual absolute coordinates to physical pixel coordinates
// relative to this display.
// Input: virtual absolute coordinates from the virtual desktop
// Output: physical pixel coordinates relative to this display
// Returns (0, 0, false) if the coordinate is not within this display.
func (d *Display) ToRelative(virtAbsX, virtAbsY int) (physX, physY int, ok bool) {
	if !d.Contains(virtAbsX, virtAbsY) {
		return 0, 0, false
	}
	// Calculate virtual relative coordinates
	virtRelX := virtAbsX - d.origin.X
	virtRelY := virtAbsY - d.origin.Y
	// Convert to physical coordinates
	if d.scale > 0 {
		physX = int(float64(virtRelX) * d.scale)
		physY = int(float64(virtRelY) * d.scale)
	} else {
		physX = virtRelX
		physY = virtRelY
	}
	return physX, physY, true
}

// Contains checks if the specified virtual absolute coordinate is within this display.
// Input: virtual absolute coordinates
func (d *Display) Contains(virtAbsX, virtAbsY int) bool {
	return virtAbsX >= d.origin.X && virtAbsX < d.origin.X+d.origin.W &&
		virtAbsY >= d.origin.Y && virtAbsY < d.origin.Y+d.origin.H
}

// Move moves the mouse to the specified physical pixel coordinates relative to this display.
// The coordinates are converted to virtual coordinates before calling the macOS API.
func (d *Display) Move(physX, physY int) {
	virtAbsX, virtAbsY := d.ToAbsolute(physX, physY)
	Move(virtAbsX, virtAbsY)
}

// MoveSmooth smoothly moves the mouse to the specified physical pixel coordinates
// relative to this display.
func (d *Display) MoveSmooth(physX, physY int, args ...interface{}) bool {
	virtAbsX, virtAbsY := d.ToAbsolute(physX, physY)
	return MoveSmooth(virtAbsX, virtAbsY, args...)
}

// Drag drags the mouse from one position to another on this display.
// Coordinates are in physical pixels relative to this display.
func (d *Display) Drag(fromX, fromY, toX, toY int, button string) {
	d.DragSmooth(fromX, fromY, toX, toY, button)
}

// DragSmooth drags the mouse smoothly from one position to another on this display.
// Coordinates are in physical pixels relative to this display.
func (d *Display) DragSmooth(fromX, fromY, toX, toY int, button string, args ...interface{}) {
	d.Move(fromX, fromY)
	Toggle(button)
	MilliSleep(50)
	virtAbsX, virtAbsY := d.ToAbsolute(toX, toY)
	dragSmooth(virtAbsX, virtAbsY, button, args...)
	Toggle(button, "up")
}

// DragTo drags the mouse from the current position to the specified position on this display.
// Coordinates are in physical pixels relative to this display.
func (d *Display) DragTo(physX, physY int, button string) {
	Toggle(button)
	MilliSleep(50)
	virtAbsX, virtAbsY := d.ToAbsolute(physX, physY)
	dragSmooth(virtAbsX, virtAbsY, button)
	Toggle(button, "up")
}

// Capture captures the entire display.
func (d *Display) Capture() (*image.RGBA, error) {
	return d.CaptureRect(0, 0, d.size.W, d.size.H)
}

// CaptureRect captures a rectangular region of this display.
// Coordinates and size are in physical pixels relative to this display.
func (d *Display) CaptureRect(physX, physY, w, h int) (*image.RGBA, error) {
	virtAbsX, virtAbsY := d.ToAbsolute(physX, physY)
	// Convert size from physical to virtual for the capture API
	virtW := w
	virtH := h
	if d.scale > 0 {
		virtW = int(float64(w) / d.scale)
		virtH = int(float64(h) / d.scale)
	}
	return Capture(virtAbsX, virtAbsY, virtW, virtH)
}

// GetPixelColor gets the pixel color at the specified physical pixel coordinates
// relative to this display.
// Returns the color as a hex string "RRGGBB".
func (d *Display) GetPixelColor(physX, physY int) string {
	virtAbsX, virtAbsY := d.ToAbsolute(physX, physY)
	return GetPixelColor(virtAbsX, virtAbsY)
}

// MouseLocation gets the mouse location in physical pixels relative to this display.
// Returns (-1, -1, false) if the mouse is not on this display.
func (d *Display) MouseLocation() (physX, physY int, ok bool) {
	virtAbsX, virtAbsY := Location()
	return d.ToRelative(virtAbsX, virtAbsY)
}

// ContainsMouse checks if the mouse is on this display.
func (d *Display) ContainsMouse() bool {
	virtAbsX, virtAbsY := Location()
	return d.Contains(virtAbsX, virtAbsY)
}
