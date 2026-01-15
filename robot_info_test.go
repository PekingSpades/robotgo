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
	"log"
	"runtime"
	"testing"
	"time"

	"github.com/go-vgo/robotgo"
	"github.com/vcaesar/tt"
)

func TestGetVer(t *testing.T) {
	fmt.Println("go version: ", runtime.Version())
	ver := robotgo.GetVersion()

	tt.Expect(t, robotgo.Version, ver)
}

func TestGetScreenSize(t *testing.T) {
	x, y := robotgo.GetScreenSize()
	log.Println("Get screen size: ", x, y)

	rect := robotgo.GetScreenRect()
	fmt.Println("Get screen rect: ", rect)

	x, y = robotgo.Location()
	fmt.Println("Get location: ", x, y)
}

func TestGetDisplayScale(t *testing.T) {
	display := robotgo.MainDisplay()
	s := display.Scale()
	log.Println("MainDisplay Scale: ", s)

	count := robotgo.DisplayCount()
	for i := 0; i < count; i++ {
		d := robotgo.DisplayAt(i)
		if d != nil {
			log.Printf("Display %d Scale: %f\n", i, d.Scale())
		}
	}
}

func TestAllDisplaysInfo(t *testing.T) {
	count := robotgo.DisplayCount()
	fmt.Printf("Total displays: %d\n", count)
	fmt.Println("----------------------------------------")

	displays := robotgo.AllDisplays()
	for _, d := range displays {
		info := d.Info()
		fmt.Printf("Display #%d\n", info.Index)
		fmt.Printf("  ID:       %d\n", info.ID)
		fmt.Printf("  IsMain:   %v\n", info.IsMain)
		fmt.Printf("  Position: (%d, %d)\n", info.Origin.X, info.Origin.Y)
		fmt.Printf("  Size:     %d x %d\n", info.Size.W, info.Size.H)
		fmt.Printf("  Scale:    %.2f\n", info.ScaleFactor)
		fmt.Println("----------------------------------------")
	}

	// Also test MainDisplay
	main := robotgo.MainDisplay()
	fmt.Printf("MainDisplay: Index=%d, Size=%dx%d, Scale=%.2f\n",
		main.Index(), main.Width(), main.Height(), main.Scale())
}

func TestDisplayMoveCorners(t *testing.T) {
	displays := robotgo.AllDisplays()
	fmt.Printf("Testing mouse movement on %d display(s)\n", len(displays))

	for _, d := range displays {
		info := d.Info()
		w, h := info.Size.W, info.Size.H
		fmt.Printf("\nDisplay #%d (%dx%d) @ (%d,%d):\n", info.Index, w, h, info.Origin.X, info.Origin.Y)

		// Define corners and center positions (relative to display)
		positions := []struct {
			name string
			x, y int
		}{
			{"top-left", 0, 0},
			{"top-right", w - 1, 0},
			{"bottom-left", 0, h - 1},
			{"bottom-right", w - 1, h - 1},
			{"center", w / 2, h / 2},
		}

		for _, pos := range positions {
			d.Move(pos.x, pos.y)
			robotgo.MilliSleep(1000)
			absX, absY := robotgo.Location()
			relX, relY, _ := d.MouseLocation()
			fmt.Printf("  %s: target=(%d,%d) abs=(%d,%d) rel=(%d,%d)\n",
				pos.name, pos.x, pos.y, absX, absY, relX, relY)
			robotgo.MilliSleep(1000)
		}
	}

	fmt.Println("\nDone!")
}

func TestGetTitle(t *testing.T) {
	// just exercise the function, it used to crash with a segfault + "Maximum
	// number of clients reached"
	for i := 0; i < 128; i++ {
		robotgo.GetTitle()
	}
}

func TestMousePositionMonitor(t *testing.T) {
	// Print screen info
	fmt.Println("========== Screen Info ==========")
	count := robotgo.DisplayCount()
	fmt.Printf("Total displays: %d\n", count)

	displays := robotgo.AllDisplays()
	for _, d := range displays {
		info := d.Info()
		fmt.Printf("Display #%d: ID=%d, IsMain=%v, Position=(%d,%d), Size=%dx%d, Scale=%.2f\n",
			info.Index, info.ID, info.IsMain,
			info.Origin.X, info.Origin.Y,
			info.Size.W, info.Size.H,
			info.ScaleFactor)
	}
	fmt.Println("=================================")

	// Print mouse position every 1 seconds
	fmt.Println("\nMonitoring mouse position (every 1 seconds)...")
	fmt.Println("Press Ctrl+C to stop")

	for i := 0; i < 120; i++ {
		x, y := robotgo.Location()
		fmt.Printf("[%s] Mouse position: (%d, %d)\n", time.Now().Format("15:04:05"), x, y)
		time.Sleep(1 * time.Second)
	}
}
