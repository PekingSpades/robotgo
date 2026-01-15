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
	"image"
	"image/color"
	"runtime"

	"github.com/go-vgo/robotgo"
	"golang.org/x/image/draw"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

func main() {
	fmt.Println("========================================")
	fmt.Println("RobotGo Capture Example")
	fmt.Printf("Go version: %s\n", runtime.Version())
	fmt.Printf("OS/Arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)
	fmt.Println("========================================")

	displays := robotgo.AllDisplays()
	count := len(displays)
	fmt.Printf("\nTotal displays: %d\n", count)
	fmt.Println("----------------------------------------")

	// Print all displays info
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

	// Calculate the bounding box for all displays (using Origin: position + logical size)
	var minX, minY, maxX, maxY int
	for i, d := range displays {
		info := d.Info()
		// Use Origin for position and logical size
		x := info.Origin.X
		y := info.Origin.Y
		w := info.Origin.W
		h := info.Origin.H

		if i == 0 {
			minX, minY = x, y
			maxX, maxY = x+w, y+h
		} else {
			if x < minX {
				minX = x
			}
			if y < minY {
				minY = y
			}
			if x+w > maxX {
				maxX = x + w
			}
			if y+h > maxY {
				maxY = y + h
			}
		}
	}

	overviewWidth := maxX - minX
	overviewHeight := maxY - minY
	fmt.Printf("\nOverview canvas: %dx%d (offset: %d, %d)\n",
		overviewWidth, overviewHeight, minX, minY)

	// Add margin for axis labels
	axisMargin := 50
	canvasWidth := overviewWidth + axisMargin
	canvasHeight := overviewHeight + axisMargin

	// Create overview image with margin for axes
	overview := image.NewRGBA(image.Rect(0, 0, canvasWidth, canvasHeight))

	// Fill with dark gray background
	bgColor := color.RGBA{R: 40, G: 40, B: 40, A: 255}
	draw.Draw(overview, overview.Bounds(), &image.Uniform{bgColor}, image.Point{}, draw.Src)

	// Draw coordinate axes
	drawAxes(overview, axisMargin, canvasWidth, canvasHeight, minX, minY, maxX, maxY)

	// Capture each display and save individually
	fmt.Println("\nCapturing displays...")
	for _, d := range displays {
		info := d.Info()
		idx := info.Index

		// Capture the display
		img, err := d.Capture()
		if err != nil {
			fmt.Printf("  Display #%d: capture failed: %v\n", idx, err)
			continue
		}

		// Save individual screenshot
		filename := fmt.Sprintf("display_%d.png", idx)
		if err := robotgo.Save(img, filename); err != nil {
			fmt.Printf("  Display #%d: save failed: %v\n", idx, err)
			continue
		}
		fmt.Printf("  Display #%d: saved to %s (%dx%d)\n",
			idx, filename, img.Bounds().Dx(), img.Bounds().Dy())

		// Calculate position in overview (relative to minX, minY, plus axis margin)
		// Use Origin for position and size (logical coordinates)
		posX := info.Origin.X - minX + axisMargin
		posY := info.Origin.Y - minY + axisMargin
		destW := info.Origin.W
		destH := info.Origin.H

		// Scale and draw the screenshot onto the overview
		destRect := image.Rect(posX, posY, posX+destW, posY+destH)
		draw.CatmullRom.Scale(overview, destRect, img, img.Bounds(), draw.Over, nil)

		// Draw Origin coordinates at top-left corner
		drawOriginLabel(overview, posX, posY, info.Origin.X, info.Origin.Y)

		// Draw info text at bottom-right corner
		drawDisplayInfo(overview, posX, posY, destW, destH, info)
	}

	// Save the overview image
	overviewFile := "display_overview.png"
	if err := robotgo.Save(overview, overviewFile); err != nil {
		fmt.Printf("\nFailed to save overview: %v\n", err)
	} else {
		fmt.Printf("\nOverview saved to %s (%dx%d)\n",
			overviewFile, canvasWidth, canvasHeight)
	}

	fmt.Println("\n========================================")
	fmt.Println("Done!")
	fmt.Println("========================================")
}

// drawDisplayInfo draws display information at the bottom-right corner of a display area
func drawDisplayInfo(img *image.RGBA, displayX, displayY, displayW, displayH int, info robotgo.DisplayInfo) {
	lines := []string{
		fmt.Sprintf("Display #%d", info.Index),
		fmt.Sprintf("Resolution: %dx%d", info.Size.W, info.Size.H),
		fmt.Sprintf("Origin: (%d, %d, %d, %d)", info.Origin.X, info.Origin.Y, info.Origin.W, info.Origin.H),
		fmt.Sprintf("Size: %dx%d", info.Size.W, info.Size.H),
		fmt.Sprintf("Scale: %.2f", info.ScaleFactor),
	}

	// Use basicfont for simple text rendering
	face := basicfont.Face7x13

	lineHeight := 16
	padding := 10
	maxWidth := 0

	// Calculate max text width
	for _, line := range lines {
		w := font.MeasureString(face, line).Ceil()
		if w > maxWidth {
			maxWidth = w
		}
	}

	// Calculate text box dimensions
	boxWidth := maxWidth + padding*2
	boxHeight := len(lines)*lineHeight + padding*2

	// Position at bottom-right corner of the display area (using logical size)
	boxX := displayX + displayW - boxWidth - padding
	boxY := displayY + displayH - boxHeight - padding

	// Draw semi-transparent background
	for y := boxY; y < boxY+boxHeight; y++ {
		for x := boxX; x < boxX+boxWidth; x++ {
			if x >= 0 && y >= 0 && x < img.Bounds().Dx() && y < img.Bounds().Dy() {
				img.Set(x, y, color.RGBA{R: 0, G: 0, B: 0, A: 180})
			}
		}
	}

	// Draw text
	textColor := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	for i, line := range lines {
		x := boxX + padding
		y := boxY + padding + (i+1)*lineHeight - 3 // baseline adjustment

		drawText(img, x, y, line, face, textColor)
	}
}

// drawText draws a string on an image at the specified position
func drawText(img *image.RGBA, x, y int, text string, face font.Face, col color.Color) {
	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(col),
		Face: face,
		Dot:  fixed.Point26_6{X: fixed.I(x), Y: fixed.I(y)},
	}
	d.DrawString(text)
}

// drawAxes draws X and Y coordinate axes with labels
func drawAxes(img *image.RGBA, margin, canvasW, canvasH, minX, minY, maxX, maxY int) {
	face := basicfont.Face7x13
	axisColor := color.RGBA{R: 200, G: 200, B: 200, A: 255}
	tickColor := color.RGBA{R: 150, G: 150, B: 150, A: 255}

	// Draw Y axis (left side)
	for y := margin; y < canvasH; y++ {
		img.Set(margin-1, y, axisColor)
		img.Set(margin, y, axisColor)
	}

	// Draw X axis (top side)
	for x := margin; x < canvasW; x++ {
		img.Set(x, margin-1, axisColor)
		img.Set(x, margin, axisColor)
	}

	// Draw tick marks and labels on X axis
	contentW := maxX - minX
	tickInterval := calculateTickInterval(contentW)
	startTick := (minX / tickInterval) * tickInterval
	if startTick < minX {
		startTick += tickInterval
	}

	for tick := startTick; tick <= maxX; tick += tickInterval {
		x := tick - minX + margin
		// Draw tick mark
		for ty := margin - 5; ty < margin; ty++ {
			img.Set(x, ty, tickColor)
		}
		// Draw label
		label := fmt.Sprintf("%d", tick)
		labelW := font.MeasureString(face, label).Ceil()
		drawText(img, x-labelW/2, margin-10, label, face, axisColor)
	}

	// Draw tick marks and labels on Y axis
	contentH := maxY - minY
	tickIntervalY := calculateTickInterval(contentH)
	startTickY := (minY / tickIntervalY) * tickIntervalY
	if startTickY < minY {
		startTickY += tickIntervalY
	}

	for tick := startTickY; tick <= maxY; tick += tickIntervalY {
		y := tick - minY + margin
		// Draw tick mark
		for tx := margin - 5; tx < margin; tx++ {
			img.Set(tx, y, tickColor)
		}
		// Draw label
		label := fmt.Sprintf("%d", tick)
		labelW := font.MeasureString(face, label).Ceil()
		drawText(img, margin-labelW-8, y+4, label, face, axisColor)
	}

	// Draw origin label
	originLabel := fmt.Sprintf("(%d,%d)", minX, minY)
	drawText(img, 2, margin-10, originLabel, face, axisColor)
}

// calculateTickInterval calculates a nice tick interval for axis labels
func calculateTickInterval(size int) int {
	if size <= 500 {
		return 100
	} else if size <= 1000 {
		return 200
	} else if size <= 2000 {
		return 500
	} else if size <= 5000 {
		return 1000
	}
	return 2000
}

// drawOriginLabel draws the Origin coordinates at top-left corner of a display
func drawOriginLabel(img *image.RGBA, x, y, originX, originY int) {
	face := basicfont.Face7x13
	label := fmt.Sprintf("(%d, %d)", originX, originY)

	padding := 5
	labelW := font.MeasureString(face, label).Ceil()
	boxW := labelW + padding*2
	boxH := 16 + padding

	// Draw semi-transparent background
	for py := y; py < y+boxH; py++ {
		for px := x; px < x+boxW; px++ {
			if px >= 0 && py >= 0 && px < img.Bounds().Dx() && py < img.Bounds().Dy() {
				img.Set(px, py, color.RGBA{R: 0, G: 0, B: 0, A: 180})
			}
		}
	}

	// Draw text
	textColor := color.RGBA{R: 255, G: 255, B: 0, A: 255} // Yellow for visibility
	drawText(img, x+padding, y+14, label, face, textColor)
}
