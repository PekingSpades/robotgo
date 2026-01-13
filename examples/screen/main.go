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
	"strconv"

	"github.com/go-vgo/robotgo"
)

func bitmap() {
	// Get main display
	display, err := robotgo.MainDisplay()
	if err != nil {
		fmt.Println("Error getting main display:", err)
		return
	}

	// Capture entire screen using Display API
	bit, err := display.CaptureAll()
	if err != nil {
		fmt.Println("Capture error:", err)
		return
	}
	defer robotgo.FreeBitmap(bit)
	fmt.Println("bitmap...", robotgo.ToBitmap(bit).Width)

	// Save using Display API
	err = display.SaveCaptureAll("saveCapture.png")
	if err != nil {
		fmt.Println("SaveCapture error:", err)
	}

	// Capture region using Display API
	img, err := display.CaptureImage(0, 0, display.Width, display.Height)
	if err != nil {
		fmt.Println("CaptureImage error:", err)
	} else {
		robotgo.Save(img, "save.png")
	}

	// Use the new Display API for multi-monitor capture
	displays, _ := robotgo.Displays()
	for i, d := range displays {
		// Capture entire display
		img1, err := d.CaptureAllImage()
		if err != nil {
			fmt.Println("Capture error:", err)
			continue
		}
		path1 := "save_" + strconv.Itoa(i)
		robotgo.Save(img1, path1+".png")
		robotgo.SaveJpeg(img1, path1+".jpeg", 50)

		// Capture a region using Display.CaptureImage
		img2, err := d.CaptureImage(10, 10, 20, 20)
		if err != nil {
			fmt.Println("Capture error:", err)
			continue
		}
		path2 := "test_" + strconv.Itoa(i)
		robotgo.Save(img2, path2+".png")
		robotgo.SaveJpeg(img2, path2+".jpeg", 50)

		// Capture using bounds from Display
		img3, err := d.CaptureAllImage()
		if err != nil {
			fmt.Println("Capture error:", err)
			continue
		}
		robotgo.Save(img3, path2+"_1.png")
	}
}

func color() {
	// Get main display for pixel color operations
	display, err := robotgo.MainDisplay()
	if err != nil {
		fmt.Println("Error getting display:", err)
		return
	}

	// gets the pixel color at 100, 200 using Display API
	color, err := display.GetPixelColor(100, 200)
	if err != nil {
		fmt.Println("GetPixelColor error:", err)
	} else {
		fmt.Println("color----", color, "-----------------")
	}

	// Use global GetPxColor (still available for virtual screen coords)
	clo := robotgo.GetPxColor(100, 200)
	fmt.Println("color...", clo)
	clostr := robotgo.PadHex(clo)
	fmt.Println("color...", clostr)

	rgb := robotgo.RgbToHex(255, 100, 200)
	rgbstr := robotgo.PadHex(robotgo.U32ToHex(rgb))
	fmt.Println("rgb...", rgbstr)

	hex := robotgo.HexToRgb(uint32(rgb))
	fmt.Println("hex...", hex)
	hexh := robotgo.PadHex(robotgo.U8ToHex(hex))
	fmt.Println("HexToRgb...", hexh)

	// gets the pixel color at 10, 20.
	color2, _ := display.GetPixelColor(10, 20)
	fmt.Println("color---", color2)
}

func screen() {
	////////////////////////////////////////////////////////////////////////////////
	// Read the screen
	////////////////////////////////////////////////////////////////////////////////

	bitmap()

	// Use the new Display API for screen info
	display, err := robotgo.MainDisplay()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println("Main display size:", display.Width, "x", display.Height)
	fmt.Println("Main display scale:", display.Scale)

	// List all displays
	displays, _ := robotgo.Displays()
	for i, d := range displays {
		fmt.Printf("Display %d: %dx%d at (%d,%d) scale=%.2f\n",
			i, d.Width, d.Height, d.X, d.Y, d.Scale)
	}

	color()
}

func main() {
	screen()
}
