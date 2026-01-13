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

/*
#include "screen/screen_c.h"
#include "mouse/mouse_c.h"
*/
import "C"

import (
	"errors"
	"image"
	"time"

	"github.com/vcaesar/screenshot"
)

// Display errors
var (
	ErrNoDisplayFound = errors.New("no display found at coordinates")
	ErrInvalidDisplay = errors.New("invalid display")
	ErrOutOfBounds    = errors.New("coordinates outside display bounds")
	ErrCaptureFailed  = errors.New("screen capture failed")
	ErrInvalidRegion  = errors.New("invalid capture region")
)

// Display represents a physical display/monitor
type Display struct {
	id     uint32  // Platform-specific display ID (internal)
	X      int     // Physical pixel X origin in virtual screen (can be negative)
	Y      int     // Physical pixel Y origin in virtual screen (can be negative)
	Width  int     // Physical pixel width
	Height int     // Physical pixel height
	Scale  float64 // DPI scale factor (e.g., 1.0, 1.5, 2.0)
}

// Bounds returns the display bounds as a Rect
func (d Display) Bounds() Rect {
	return Rect{
		Point: Point{X: d.X, Y: d.Y},
		Size:  Size{W: d.Width, H: d.Height},
	}
}

// Contains checks if a physical pixel coordinate is within this display
func (d Display) Contains(x, y int) bool {
	return x >= d.X && x < d.X+d.Width &&
		y >= d.Y && y < d.Y+d.Height
}

// ToDisplayCoord converts virtual screen coordinates to display-relative coordinates
func (d Display) ToDisplayCoord(x, y int) (int, int) {
	return x - d.X, y - d.Y
}

// ToVirtualCoord converts display-relative coordinates to virtual screen coordinates
func (d Display) ToVirtualCoord(x, y int) (int, int) {
	return x + d.X, y + d.Y
}

// ID returns the internal platform-specific display ID
func (d Display) ID() uint32 {
	return d.id
}

// MainDisplay returns the primary display
func MainDisplay() (Display, error) {
	displays, err := Displays()
	if err != nil {
		return Display{}, err
	}
	if len(displays) == 0 {
		return Display{}, ErrNoDisplayFound
	}
	// Main display is always the first one (index 0)
	return displays[0], nil
}

// Displays returns all available displays
func Displays() ([]Display, error) {
	n := screenshot.NumActiveDisplays()
	if n == 0 {
		return nil, ErrNoDisplayFound
	}

	displays := make([]Display, n)
	for i := 0; i < n; i++ {
		bounds := screenshot.GetDisplayBounds(i)

		// Get physical pixel dimensions
		// On macOS, screenshot.GetDisplayBounds returns points, need to convert to pixels
		// On Windows/Linux, it returns physical pixels
		physicalWidth := bounds.Dx()
		physicalHeight := bounds.Dy()
		physicalX := bounds.Min.X
		physicalY := bounds.Min.Y

		// Get scale factor for this display
		scale := getDisplayScale(i)
		if scale == 0 {
			scale = 1.0
		}

		// On macOS, convert points to physical pixels
		if isMacOSPlatform {
			physicalWidth = int(float64(physicalWidth) * scale)
			physicalHeight = int(float64(physicalHeight) * scale)
			physicalX = int(float64(physicalX) * scale)
			physicalY = int(float64(physicalY) * scale)
		}

		displays[i] = Display{
			id:     uint32(i),
			X:      physicalX,
			Y:      physicalY,
			Width:  physicalWidth,
			Height: physicalHeight,
			Scale:  scale,
		}
	}

	return displays, nil
}

// DisplayAt returns the display containing the given virtual screen coordinate (physical pixels)
func DisplayAt(x, y int) (Display, error) {
	displays, err := Displays()
	if err != nil {
		return Display{}, err
	}

	for _, d := range displays {
		if d.Contains(x, y) {
			return d, nil
		}
	}

	return Display{}, ErrNoDisplayFound
}

// getDisplayScale returns the scale factor for display at index i
func getDisplayScale(i int) float64 {
	// Use the C function sys_scale
	return float64(C.sys_scale(C.int32_t(i)))
}

// Move moves the mouse to display-relative physical pixel coordinates.
// The coordinates (x, y) are relative to this display's origin.
func (d Display) Move(x, y int) error {
	// Convert display-relative to virtual screen coordinates
	vx, vy := d.ToVirtualCoord(x, y)

	// Call the physical pixel C function with the display's scale factor
	point := C.MMPointInt32Make(C.int32_t(vx), C.int32_t(vy))
	C.moveMousePhysical(point, C.double(d.Scale))

	if MouseSleep > 0 {
		time.Sleep(time.Duration(MouseSleep) * time.Millisecond)
	}

	return nil
}

// MoveSmooth smoothly moves the mouse to display-relative physical pixel coordinates.
// The coordinates (x, y) are relative to this display's origin.
// low and high control the speed range (milliseconds between moves).
func (d Display) MoveSmooth(x, y int, low, high float64) error {
	// Convert display-relative to virtual screen coordinates
	vx, vy := d.ToVirtualCoord(x, y)

	// Note: smoothlyMoveMouse uses the regular moveMouse internally,
	// which we need to update to support physical pixels.
	// For now, we call it directly - this will be updated in a later phase.
	point := C.MMPointInt32Make(C.int32_t(vx), C.int32_t(vy))
	C.smoothlyMoveMouse(point, C.double(low), C.double(high))

	if MouseSleep > 0 {
		time.Sleep(time.Duration(MouseSleep) * time.Millisecond)
	}

	return nil
}

// Capture captures a region of this display.
// The coordinates (x, y, w, h) are in physical pixels relative to this display's origin.
// Returns a CBitmap that must be freed with FreeBitmap when done.
func (d Display) Capture(x, y, w, h int) (CBitmap, error) {
	// Validate region
	if w <= 0 || h <= 0 {
		return nil, ErrInvalidRegion
	}

	// Convert display-relative to virtual screen coordinates
	vx, vy := d.ToVirtualCoord(x, y)

	// Convert physical pixels to points for macOS screenshot library
	captureX, captureY, captureW, captureH := vx, vy, w, h
	if isMacOSPlatform {
		captureX = int(float64(vx) / d.Scale)
		captureY = int(float64(vy) / d.Scale)
		captureW = int(float64(w) / d.Scale)
		captureH = int(float64(h) / d.Scale)
	}

	// Use the C capture function
	bit := C.capture_screen(
		C.int32_t(captureX),
		C.int32_t(captureY),
		C.int32_t(captureW),
		C.int32_t(captureH),
		C.int32_t(d.id),
		C.int8_t(0),
	)

	if bit == nil {
		return nil, ErrCaptureFailed
	}

	return CBitmap(bit), nil
}

// CaptureImage captures a region of this display and returns an image.Image.
// The coordinates (x, y, w, h) are in physical pixels relative to this display's origin.
func (d Display) CaptureImage(x, y, w, h int) (image.Image, error) {
	bit, err := d.Capture(x, y, w, h)
	if err != nil {
		return nil, err
	}
	defer FreeBitmap(bit)

	return ToImage(bit), nil
}

// CaptureRGBA captures a region of this display and returns an *image.RGBA.
// The coordinates (x, y, w, h) are in physical pixels relative to this display's origin.
func (d Display) CaptureRGBA(x, y, w, h int) (*image.RGBA, error) {
	// Convert display-relative to virtual screen coordinates
	vx, vy := d.ToVirtualCoord(x, y)

	// Convert physical pixels to points for macOS screenshot library
	captureX, captureY, captureW, captureH := vx, vy, w, h
	if isMacOSPlatform {
		captureX = int(float64(vx) / d.Scale)
		captureY = int(float64(vy) / d.Scale)
		captureW = int(float64(w) / d.Scale)
		captureH = int(float64(h) / d.Scale)
	}

	return screenshot.Capture(captureX, captureY, captureW, captureH)
}

// CaptureAll captures the entire display.
// Returns a CBitmap that must be freed with FreeBitmap when done.
func (d Display) CaptureAll() (CBitmap, error) {
	return d.Capture(0, 0, d.Width, d.Height)
}

// CaptureAllImage captures the entire display and returns an image.Image.
func (d Display) CaptureAllImage() (image.Image, error) {
	return d.CaptureImage(0, 0, d.Width, d.Height)
}

// CaptureAllRGBA captures the entire display and returns an *image.RGBA.
func (d Display) CaptureAllRGBA() (*image.RGBA, error) {
	return d.CaptureRGBA(0, 0, d.Width, d.Height)
}

// SaveCapture captures a region and saves it to a file.
// The coordinates (x, y, w, h) are in physical pixels relative to this display's origin.
func (d Display) SaveCapture(path string, x, y, w, h int) error {
	img, err := d.CaptureImage(x, y, w, h)
	if err != nil {
		return err
	}
	return Save(img, path)
}

// SaveCaptureAll captures the entire display and saves it to a file.
func (d Display) SaveCaptureAll(path string) error {
	return d.SaveCapture(path, 0, 0, d.Width, d.Height)
}

// GetPixelColor gets the pixel color at display-relative coordinates.
// The coordinates (x, y) are in physical pixels relative to this display's origin.
// Returns the color as a hex string (e.g., "ff0000" for red).
func (d Display) GetPixelColor(x, y int) (string, error) {
	hex, err := d.GetPixelHex(x, y)
	if err != nil {
		return "", err
	}
	return PadHexs(hex), nil
}

// GetPixelHex gets the pixel color at display-relative coordinates as CHex.
// The coordinates (x, y) are in physical pixels relative to this display's origin.
func (d Display) GetPixelHex(x, y int) (CHex, error) {
	// Validate coordinates are within display bounds
	if x < 0 || y < 0 || x >= d.Width || y >= d.Height {
		return 0, ErrOutOfBounds
	}

	// Convert display-relative to virtual screen coordinates
	vx, vy := d.ToVirtualCoord(x, y)

	// Convert physical pixels to points for macOS
	captureX, captureY := vx, vy
	if isMacOSPlatform {
		captureX = int(float64(vx) / d.Scale)
		captureY = int(float64(vy) / d.Scale)
	}

	color := C.get_px_color(C.int32_t(captureX), C.int32_t(captureY), C.int32_t(d.id))
	return CHex(color), nil
}
