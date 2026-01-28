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
	"bufio"
	"fmt"
	"os"
	"runtime"
	"strings"

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
		fmt.Printf("  Origin:   {X: %d, Y: %d, W: %d, H: %d}\n",
			info.Origin.X, info.Origin.Y, info.Origin.W, info.Origin.H)
		fmt.Printf("  Size:     {W: %d, H: %d}\n", info.Size.W, info.Size.H)
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

	margin := 32

	for _, d := range displays {
		info := d.Info()
		w, h := info.Size.W, info.Size.H
		fmt.Printf("\nDisplay #%d (%dx%d) @ (%d,%d):\n",
			info.Index, w, h, info.Origin.X, info.Origin.Y)

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

	// Collect all output for potential saving
	var logBuilder strings.Builder
	logBuilder.WriteString("RobotGo Display Example\n")
	logBuilder.WriteString(fmt.Sprintf("Go version: %s\n", runtime.Version()))
	logBuilder.WriteString(fmt.Sprintf("OS/Arch: %s/%s\n", runtime.GOOS, runtime.GOARCH))
	logBuilder.WriteString(fmt.Sprintf("Total displays: %d\n", count))
	for _, d := range displays {
		info := d.Info()
		logBuilder.WriteString(fmt.Sprintf("Display #%d: ID=%d, IsMain=%v, Origin=(%d,%d,%d,%d), Size=%dx%d, Scale=%.2f\n",
			info.Index, info.ID, info.IsMain, info.Origin.X, info.Origin.Y, info.Origin.W, info.Origin.H, info.Size.W, info.Size.H, info.ScaleFactor))
	}

	fmt.Println("\nPress 's' to save log and exit, or 'e' to exit directly:")
	reader := bufio.NewReader(os.Stdin)
	for {
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(strings.ToLower(input))
		if input == "s" {
			if err := os.WriteFile("display_log.txt", []byte(logBuilder.String()), 0644); err != nil {
				fmt.Printf("Failed to save log: %v\n", err)
			} else {
				fmt.Println("Log saved to display_log.txt")
			}
			break
		} else if input == "e" {
			break
		}
		fmt.Println("Press 's' to save log and exit, or 'e' to exit directly:")
	}
}
