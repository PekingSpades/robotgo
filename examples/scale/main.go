package main

import (
	"fmt"

	"github.com/go-vgo/robotgo"
)

func main() {
	// syscall.NewLazyDLL("user32.dll").NewProc("SetProcessDPIAware").Call()

	width, height := robotgo.GetScaleSize()
	fmt.Println("get scale screen size: ", width, height)

	bitmap := robotgo.CaptureScreen(0, 0, width, height)
	defer robotgo.FreeBitmap(bitmap)
	// bitmap.Save(bitmap, "test.png")
	robotgo.Save(robotgo.ToImage(bitmap), "test.png")

	// Use the new Display API for physical pixel operations
	display, err := robotgo.MainDisplay()
	if err != nil {
		fmt.Println("error getting main display:", err)
		return
	}
	fmt.Println("Main display:", display.Width, "x", display.Height, "scale:", display.Scale)

	robotgo.Move(10, 10)
	robotgo.MoveSmooth(100, 100, 1.0, 3.0)

	x, y, _ := robotgo.Location()
	fmt.Println("Location:", x, y)

	// List all displays
	displays, _ := robotgo.Displays()
	for i, d := range displays {
		fmt.Printf("Display %d: %dx%d at (%d,%d) scale=%.2f\n",
			i, d.Width, d.Height, d.X, d.Y, d.Scale)
	}

	// Get pixel color using new Display API
	color, err := display.GetPixelColor(100, 100)
	if err == nil {
		fmt.Println("Pixel color at (100, 100):", color)
	}
}
