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

package robotgo_test

import (
	"fmt"
	"runtime"
	"testing"

	"github.com/go-vgo/robotgo"
	"github.com/vcaesar/tt"
)

func TestGetVer(t *testing.T) {
	fmt.Println("go version: ", runtime.Version())
	ver := robotgo.GetVersion()

	tt.Expect(t, robotgo.Version, ver)
}

func TestDisplay(t *testing.T) {
	// Test the new Display API
	display, err := robotgo.MainDisplay()
	tt.Nil(t, err)
	fmt.Println("Main display: ", display.Width, "x", display.Height, "scale:", display.Scale)

	// Test all displays
	displays, err := robotgo.Displays()
	tt.Nil(t, err)
	for i, d := range displays {
		fmt.Printf("Display %d: %dx%d at (%d,%d) scale=%.2f\n",
			i, d.Width, d.Height, d.X, d.Y, d.Scale)
	}

	x, y, _ := robotgo.Location()
	fmt.Println("Get location: ", x, y)
}

func TestGetTitle(t *testing.T) {
	// just exercise the function, it used to crash with a segfault + "Maximum
	// number of clients reached"
	for i := 0; i < 128; i++ {
		robotgo.GetTitle()
	}
}
