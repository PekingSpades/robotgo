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

package main

import (
	"fmt"
	"runtime"

	"github.com/go-vgo/robotgo"
)

func main() {
	fmt.Println("========================================")
	fmt.Println("RobotGo Display Example")
	fmt.Printf("Go version: %s\n", runtime.Version())
	fmt.Printf("OS/Arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)
	fmt.Println("========================================")

	// Print all displays info
	count := robotgo.DisplayCount()
	fmt.Printf("\nTotal displays: %d\n", count)
	fmt.Println("----------------------------------------")

	displays := robotgo.AllDisplays()
	for _, d := range displays {
		info := d.Info()
		fmt.Printf("Display #%d\n", info.Index)
		fmt.Printf("  ID:       %d\n", info.ID)
		fmt.Printf("  IsMain:   %v\n", info.IsMain)
		fmt.Printf("  Position: (%d, %d)\n", info.Bounds.X, info.Bounds.Y)
		fmt.Printf("  Size:     %d x %d\n", info.Bounds.W, info.Bounds.H)
		fmt.Printf("  Scale:    %.2f\n", info.ScaleFactor)
		fmt.Println("----------------------------------------")
	}

	// Print main display info
	main := robotgo.MainDisplay()
	fmt.Printf("MainDisplay: Index=%d, Size=%dx%d, Scale=%.2f\n",
		main.Index(), main.Width(), main.Height(), main.Scale())

	// Move mouse to each display's corners and center (with 10px margin from edges)
	fmt.Println("\n========================================")
	fmt.Println("Moving mouse to each display's corners")
	fmt.Println("========================================")

	margin := 10

	for _, d := range displays {
		info := d.Info()
		w, h := info.Bounds.W, info.Bounds.H
		fmt.Printf("\nDisplay #%d (%dx%d) @ (%d,%d):\n",
			info.Index, w, h, info.Bounds.X, info.Bounds.Y)

		// Define positions with 10px margin from edges
		positions := []struct {
			name string
			x, y int
		}{
			{"top-left", margin, margin},
			{"top-right", w - 1 - margin, margin},
			{"bottom-left", margin, h - 1 - margin},
			{"bottom-right", w - 1 - margin, h - 1 - margin},
			{"center", w / 2, h / 2},
		}

		for _, pos := range positions {
			d.Move(pos.x, pos.y)
			robotgo.MilliSleep(800)

			absX, absY := robotgo.Location()
			relX, relY, ok := d.MouseLocation()
			if ok {
				fmt.Printf("  %-12s: target=(%4d,%4d) abs=(%5d,%5d) rel=(%4d,%4d)\n",
					pos.name, pos.x, pos.y, absX, absY, relX, relY)
			} else {
				fmt.Printf("  %-12s: target=(%4d,%4d) abs=(%5d,%5d) (mouse not on this display)\n",
					pos.name, pos.x, pos.y, absX, absY)
			}
			robotgo.MilliSleep(200)
		}
	}

	fmt.Println("\n========================================")
	fmt.Println("Done!")
	fmt.Println("========================================")
}
