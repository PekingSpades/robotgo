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
	"image"

	"github.com/vcaesar/screenshot"
)

// Capture capture the screenshot, use the CaptureImg default
//
// Deprecated: Use Display.CaptureRGBA() instead for physical pixel coordinates.
func Capture(args ...int) (*image.RGBA, error) {
	display, err := MainDisplay()
	if err != nil {
		return nil, err
	}

	var x, y, w, h int
	if len(args) > 3 {
		x, y, w, h = args[0], args[1], args[2], args[3]
	} else {
		x, y, w, h = 0, 0, display.Width, display.Height
	}

	// Convert physical pixels to points for screenshot library on macOS
	captureX, captureY, captureW, captureH := x, y, w, h
	if isMacOSPlatform {
		captureX = int(float64(x) / display.Scale)
		captureY = int(float64(y) / display.Scale)
		captureW = int(float64(w) / display.Scale)
		captureH = int(float64(h) / display.Scale)
	}

	return screenshot.Capture(captureX, captureY, captureW, captureH)
}

// SaveCapture capture screen and save the screenshot to image
//
// Deprecated: Use Display.SaveCapture() instead for physical pixel coordinates.
func SaveCapture(path string, args ...int) error {
	img, err := CaptureImg(args...)
	if err != nil {
		return err
	}

	return Save(img, path)
}
