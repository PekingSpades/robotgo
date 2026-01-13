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

//go:build linux
// +build linux

package robotgo

/*
#cgo linux CFLAGS: -I/usr/src
#cgo linux LDFLAGS: -L/usr/src -lm -lX11 -lXtst -lXinerama

#include "screen/goScreen.h"
#include "screen/display_c.h"
#include "mouse/mouse_c.h"
*/
import "C"

import (
	"image"
)

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
		scale: float64(info.scale),
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
			scale: float64(info.scale),
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
		scale: float64(info.scale),
	}
}

// DisplayCount returns the number of displays.
func DisplayCount() int {
	return int(C.getDisplayCount())
}

// Move moves the mouse to the specified coordinates relative to this display.
// On Linux, coordinates are in physical pixels.
func (d *Display) Move(x, y int) {
	absX, absY := d.ToAbsolute(x, y)
	cx := C.int32_t(absX)
	cy := C.int32_t(absY)
	C.moveMouse(C.MMPointInt32Make(cx, cy))
	MilliSleep(MouseSleep)
}

// MoveSmooth smoothly moves the mouse to the specified coordinates relative to this display.
func (d *Display) MoveSmooth(x, y int, args ...interface{}) bool {
	absX, absY := d.ToAbsolute(x, y)

	var (
		mouseDelay = 1
		low        C.double
		high       C.double
	)

	if len(args) > 2 {
		mouseDelay = args[2].(int)
	}

	if len(args) > 1 {
		low = C.double(args[0].(float64))
		high = C.double(args[1].(float64))
	} else {
		low = 1.0
		high = 3.0
	}

	cbool := C.smoothlyMoveMouse(C.MMPointInt32Make(C.int32_t(absX), C.int32_t(absY)), low, high)
	MilliSleep(MouseSleep + mouseDelay)

	return bool(cbool)
}

// Drag drags the mouse from one position to another on this display.
func (d *Display) Drag(fromX, fromY, toX, toY int, button string) {
	d.Move(fromX, fromY)
	Toggle(button)
	MilliSleep(50)
	d.MoveSmooth(toX, toY)
	Toggle(button, "up")
}

// DragTo drags the mouse from the current position to the specified position on this display.
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
// Coordinates are relative to this display.
func (d *Display) CaptureRect(x, y, w, h int) (*image.RGBA, error) {
	absX, absY := d.ToAbsolute(x, y)

	bit := C.capture_screen(
		C.int32_t(absX),
		C.int32_t(absY),
		C.int32_t(w),
		C.int32_t(h),
		C.int32_t(-1),
		C.int8_t(0),
	)

	if bit == nil {
		return nil, ErrCaptureScreen
	}
	defer FreeBitmap(CBitmap(bit))

	return ToRGBA(CBitmap(bit)), nil
}

// GetPixelColor gets the pixel color at the specified coordinates relative to this display.
// Returns the color as a hex string "RRGGBB".
func (d *Display) GetPixelColor(x, y int) string {
	absX, absY := d.ToAbsolute(x, y)
	return GetPixelColor(absX, absY)
}

// MouseLocation gets the mouse location relative to this display.
// Returns (-1, -1, false) if the mouse is not on this display.
func (d *Display) MouseLocation() (x, y int, ok bool) {
	absX, absY := Location()
	return d.ToRelative(absX, absY)
}

// ContainsMouse checks if the mouse is on this display.
func (d *Display) ContainsMouse() bool {
	absX, absY := Location()
	return d.Contains(absX, absY)
}
