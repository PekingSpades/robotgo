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

#include "mouse.h"
#include "../base/deadbeef_rand.h"
#include "../base/microsleep.h"

#include <math.h> /* For floor() */
#if defined(IS_MACOSX)
	// #include </System/Library/Frameworks/ApplicationServices.framework/Headers/ApplicationServices.h>
	#include <ApplicationServices/ApplicationServices.h>
	// #include </System/Library/Frameworks/ApplicationServices.framework/Versions/A/Headers/ApplicationServices.h>
#elif defined(USE_X11)
	#include <X11/Xlib.h>
	#include <X11/extensions/XTest.h>
	#include <stdlib.h>
#endif

/* Some convenience macros for converting our enums to the system API types. */
#if defined(IS_MACOSX)
	CGEventType MMMouseDownToCGEventType(MMMouseButton button) {
		if (button == LEFT_BUTTON) {
			return kCGEventLeftMouseDown;
		}
		if (button == RIGHT_BUTTON) { 
			return kCGEventRightMouseDown;
		}
		return kCGEventOtherMouseDown;
	}

	CGEventType MMMouseUpToCGEventType(MMMouseButton button) {
		if (button == LEFT_BUTTON) { return kCGEventLeftMouseUp; }
		if (button == RIGHT_BUTTON) { return kCGEventRightMouseUp; }		
		return kCGEventOtherMouseUp;
	}

	CGEventType MMMouseDragToCGEventType(MMMouseButton button) {
		if (button == LEFT_BUTTON) { return kCGEventLeftMouseDragged; }
		if (button == RIGHT_BUTTON) { return kCGEventRightMouseDragged; }
		return kCGEventOtherMouseDragged;
	}

	CGEventType MMMouseToCGEventType(bool down, MMMouseButton button) {
		if (down) { return MMMouseDownToCGEventType(button); }
		return MMMouseUpToCGEventType(button);
	}

#elif defined(IS_WINDOWS)
 
	DWORD MMMouseUpToMEventF(MMMouseButton button) {
		if (button == LEFT_BUTTON) { return MOUSEEVENTF_LEFTUP; }
		if (button == RIGHT_BUTTON) { return MOUSEEVENTF_RIGHTUP; } 
		return MOUSEEVENTF_MIDDLEUP;
	}

	DWORD MMMouseDownToMEventF(MMMouseButton button) {
		if (button == LEFT_BUTTON) { return MOUSEEVENTF_LEFTDOWN; }
		if (button == RIGHT_BUTTON) { return MOUSEEVENTF_RIGHTDOWN; } 
		return MOUSEEVENTF_MIDDLEDOWN;
	}

	DWORD MMMouseToMEventF(bool down, MMMouseButton button) {
		if (down) { return MMMouseDownToMEventF(button); }
		return MMMouseUpToMEventF(button);
	}
#endif

#if defined(IS_MACOSX)
	/* Calculate the delta for a mouse move and add them to the event. */
	void calculateDeltas(CGEventRef *event, MMPointInt32 point) {
		/* The next few lines are a workaround for games not detecting mouse moves. */
		CGEventRef get = CGEventCreate(NULL);
		CGPoint mouse = CGEventGetLocation(get);

		// Calculate the deltas.
		int64_t deltaX = point.x - mouse.x;
		int64_t deltaY = point.y - mouse.y;

		CGEventSetIntegerValueField(*event, kCGMouseEventDeltaX, deltaX);
		CGEventSetIntegerValueField(*event, kCGMouseEventDeltaY, deltaY);

		CFRelease(get);
	}
#endif

/* Move the mouse using physical pixel coordinates (not affected by DPI settings).
 * On Windows: uses SetPhysicalCursorPos for consistent physical pixel positioning
 * On macOS: converts physical pixels to points using the provided scale factor
 * On X11: same as moveMouse (X11 already uses physical pixels)
 */
void moveMousePhysical(MMPointInt32 point, double scale){
	#if defined(IS_MACOSX)
		// Convert physical pixels to points
		CGPoint cgPoint;
		cgPoint.x = (CGFloat)point.x / scale;
		cgPoint.y = (CGFloat)point.y / scale;

		CGEventSourceRef source = CGEventSourceCreate(kCGEventSourceStateHIDSystemState);
		CGEventRef move = CGEventCreateMouseEvent(source, kCGEventMouseMoved,
								cgPoint, kCGMouseButtonLeft);

		// Calculate deltas using point coordinates
		CGEventRef get = CGEventCreate(NULL);
		CGPoint mouse = CGEventGetLocation(get);
		int64_t deltaX = cgPoint.x - mouse.x;
		int64_t deltaY = cgPoint.y - mouse.y;
		CGEventSetIntegerValueField(move, kCGMouseEventDeltaX, deltaX);
		CGEventSetIntegerValueField(move, kCGMouseEventDeltaY, deltaY);
		CFRelease(get);

		CGEventPost(kCGHIDEventTap, move);
		CFRelease(move);
		CFRelease(source);
	#elif defined(USE_X11)
		// X11 already uses physical pixels
		Display *display = XGetMainDisplay();
		XWarpPointer(display, None, DefaultRootWindow(display), 0, 0, 0, 0, point.x, point.y);
		XSync(display, false);
	#elif defined(IS_WINDOWS)
		// Use SetPhysicalCursorPos for physical pixel positioning
		// This is not affected by DPI virtualization
		SetPhysicalCursorPos(point.x, point.y);
	#endif
}

void dragMouse(MMPointInt32 point, const MMMouseButton button){
	#if defined(IS_MACOSX)
		const CGEventType dragType = MMMouseDragToCGEventType(button);
		CGEventSourceRef source = CGEventSourceCreate(kCGEventSourceStateHIDSystemState);
		CGEventRef drag = CGEventCreateMouseEvent(source, dragType, 
								CGPointFromMMPointInt32(point), (CGMouseButton)button);

		calculateDeltas(&drag, point);

		CGEventPost(kCGHIDEventTap, drag);
		CFRelease(drag);
		CFRelease(source);
	#else
		moveMouse(point);
	#endif
}

/* Get mouse location in physical pixel coordinates.
 * On Windows: uses GetPhysicalCursorPos for consistent physical pixel positioning
 * On macOS: converts points to physical pixels using the display's scale factor
 * On X11: same as location (X11 already uses physical pixels)
 */
MMPointInt32 locationPhysical(double scale) {
	#if defined(IS_MACOSX)
		CGEventRef event = CGEventCreate(NULL);
		CGPoint pointInPoints = CGEventGetLocation(event);
		CFRelease(event);

		// Convert points to physical pixels
		return MMPointInt32Make(
			(int32_t)(pointInPoints.x * scale),
			(int32_t)(pointInPoints.y * scale)
		);
	#elif defined(USE_X11)
		// X11 already uses physical pixels
		int x, y;
		Window garb1, garb2;
		int garb_x, garb_y;
		unsigned int more_garbage;

		Display *display = XGetMainDisplay();
		XQueryPointer(display, XDefaultRootWindow(display), &garb1, &garb2, &x, &y,
						&garb_x, &garb_y, &more_garbage);

		return MMPointInt32Make(x, y);
	#elif defined(IS_WINDOWS)
		POINT point;
		// Use GetPhysicalCursorPos for physical pixel positioning
		// This is not affected by DPI virtualization
		GetPhysicalCursorPos(&point);
		return MMPointInt32FromPOINT(point);
	#endif
}

/* Get mouse location in physical pixel coordinates with auto-detected scale.
 * On macOS: determines which display the cursor is on and uses that display's scale
 * On Windows/X11: same as locationPhysical (scale doesn't matter)
 */
MMPointInt32 locationPhysicalAuto() {
	#if defined(IS_MACOSX)
		CGEventRef event = CGEventCreate(NULL);
		CGPoint pointInPoints = CGEventGetLocation(event);
		CFRelease(event);

		// Find which display contains this point
		CGDirectDisplayID displayID;
		uint32_t count;
		CGGetDisplaysWithPoint(pointInPoints, 1, &displayID, &count);

		if (count == 0) {
			// Default to main display if no display found
			displayID = CGMainDisplayID();
		}

		// Get scale factor for that display
		CGDisplayModeRef modeRef = CGDisplayCopyDisplayMode(displayID);
		double pixelWidth = CGDisplayModeGetPixelWidth(modeRef);
		double targetWidth = CGDisplayModeGetWidth(modeRef);
		CGDisplayModeRelease(modeRef);

		double scale = pixelWidth / targetWidth;

		// Convert points to physical pixels
		return MMPointInt32Make(
			(int32_t)(pointInPoints.x * scale),
			(int32_t)(pointInPoints.y * scale)
		);
	#elif defined(USE_X11)
		// X11 already uses physical pixels
		int x, y;
		Window garb1, garb2;
		int garb_x, garb_y;
		unsigned int more_garbage;

		Display *display = XGetMainDisplay();
		XQueryPointer(display, XDefaultRootWindow(display), &garb1, &garb2, &x, &y,
						&garb_x, &garb_y, &more_garbage);

		return MMPointInt32Make(x, y);
	#elif defined(IS_WINDOWS)
		POINT point;
		GetPhysicalCursorPos(&point);
		return MMPointInt32FromPOINT(point);
	#endif
}

/* Press down a button, or release it. */
int toggleMouseErr(bool down, MMMouseButton button) {
	#if defined(IS_MACOSX)
		const CGPoint currentPos = CGPointFromMMPointInt32(location());
		const CGEventType mouseType = MMMouseToCGEventType(down, button);
		CGEventSourceRef source = CGEventSourceCreate(kCGEventSourceStateHIDSystemState);
		CGEventRef event = CGEventCreateMouseEvent(source, mouseType, currentPos, (CGMouseButton)button);

		if (event == NULL) {
			CFRelease(source);
			return (int)kCGErrorCannotComplete;
		}

		CGEventPost(kCGHIDEventTap, event);
		CFRelease(event);
		CFRelease(source);

		return 0;
	#elif defined(USE_X11)
		Display *display = XGetMainDisplay();
		Status status = XTestFakeButtonEvent(display, button, down ? True : False, CurrentTime);
		XSync(display, false);

		return status ? 0 : 1;
	#elif defined(IS_WINDOWS)
		// mouse_event(MMMouseToMEventF(down, button), 0, 0, 0, 0);
		INPUT mouseInput;

		mouseInput.type = INPUT_MOUSE;
		mouseInput.mi.dx = 0;
		mouseInput.mi.dy = 0;
		mouseInput.mi.dwFlags = MMMouseToMEventF(down, button);
		mouseInput.mi.time = 0;
		mouseInput.mi.dwExtraInfo = 0;
		mouseInput.mi.mouseData = 0;
		UINT sent = SendInput(1, &mouseInput, sizeof(mouseInput));
		return sent == 1 ? 0 : (int)GetLastError();
	#endif
}

/* Multi-click function supporting any click count (1=single, 2=double, 3=triple, etc.) */
int multiClickErr(MMMouseButton button, int clickCount){
	if (clickCount < 1) {
		return 0;
	}

	#if defined(IS_MACOSX)
		const CGPoint currentPos = CGPointFromMMPointInt32(location());
		const CGEventType mouseTypeDown = MMMouseToCGEventType(true, button);
		const CGEventType mouseTypeUP = MMMouseToCGEventType(false, button);

		CGEventSourceRef source = CGEventSourceCreate(kCGEventSourceStateHIDSystemState);
		CGEventRef event = CGEventCreateMouseEvent(source, mouseTypeDown, currentPos, (CGMouseButton)button);

		if (event == NULL) {
			CFRelease(source);
			return (int)kCGErrorCannotComplete;
		}

		CGEventSetIntegerValueField(event, kCGMouseEventClickState, clickCount);
		CGEventPost(kCGHIDEventTap, event);

		CGEventSetType(event, mouseTypeUP);
		CGEventPost(kCGHIDEventTap, event);

		CFRelease(event);
		CFRelease(source);

		return 0;
	#else
		int i;
		for (i = 0; i < clickCount; i++) {
			int err = toggleMouseErr(true, button);
			if (err != 0) {
				return err;
			}
			microsleep(5.0);
			err = toggleMouseErr(false, button);
			if (err != 0) {
				return err;
			}
			if (i < clickCount - 1) {
				microsleep(200);
			}
		}
		return 0;
	#endif
}

/* Function used to scroll the screen in the required direction. */
void scrollMouseXY(int x, int y) {
	#if defined(IS_WINDOWS)
		// Fix for #97, C89 needs variables declared on top of functions (mouseScrollInput)
		INPUT mouseScrollInputH;
		INPUT mouseScrollInputV;
	#endif

	#if defined(IS_MACOSX)
		CGEventSourceRef source = CGEventSourceCreate(kCGEventSourceStateHIDSystemState);
		CGEventRef event = CGEventCreateScrollWheelEvent(source, kCGScrollEventUnitPixel, 2, y, x);
		CGEventPost(kCGHIDEventTap, event);

		CFRelease(event);
		CFRelease(source);
	#elif defined(USE_X11)
		int ydir = 4; /* Button 4 is up, 5 is down. */
		int xdir = 6;
		Display *display = XGetMainDisplay();

		if (y < 0) { ydir = 5; }
		if (x < 0) { xdir = 7; }

		int xi; int yi;
		for (xi = 0; xi < abs(x); xi++) {
			XTestFakeButtonEvent(display, xdir, 1, CurrentTime);
			XTestFakeButtonEvent(display, xdir, 0, CurrentTime);
		}
		for (yi = 0; yi < abs(y); yi++) {
			XTestFakeButtonEvent(display, ydir, 1, CurrentTime);
			XTestFakeButtonEvent(display, ydir, 0, CurrentTime);
		}

		XSync(display, false);
	#elif defined(IS_WINDOWS)
		mouseScrollInputH.type = INPUT_MOUSE;
		mouseScrollInputH.mi.dx = 0;
		mouseScrollInputH.mi.dy = 0;
		mouseScrollInputH.mi.dwFlags = MOUSEEVENTF_WHEEL;
		mouseScrollInputH.mi.time = 0;
		mouseScrollInputH.mi.dwExtraInfo = 0;
		mouseScrollInputH.mi.mouseData = WHEEL_DELTA * x;

		mouseScrollInputV.type = INPUT_MOUSE;
		mouseScrollInputV.mi.dx = 0;
		mouseScrollInputV.mi.dy = 0;
		mouseScrollInputV.mi.dwFlags = MOUSEEVENTF_WHEEL;
		mouseScrollInputV.mi.time = 0;
		mouseScrollInputV.mi.dwExtraInfo = 0;
		mouseScrollInputV.mi.mouseData = WHEEL_DELTA * y;

		SendInput(1, &mouseScrollInputH, sizeof(mouseScrollInputH));
		SendInput(1, &mouseScrollInputV, sizeof(mouseScrollInputV));
	#endif
}

/* A crude, fast hypot() approximation to get around the fact that hypot() is not a standard ANSI C function. */
#if !defined(M_SQRT2)
	#define M_SQRT2 1.4142135623730950488016887 /* Fix for MSVC. */
#endif

static double crude_hypot(double x, double y){
	double big = fabs(x); /* max(|x|, |y|) */
	double small = fabs(y); /* min(|x|, |y|) */

	if (big > small) {
		double temp = big;
		big = small;
		small = temp;
	}

	return ((M_SQRT2 - 1.0) * small) + big;
}

bool smoothlyMoveMouse(MMPointInt32 endPoint, double lowSpeed, double highSpeed){
	MMPointInt32 pos = location();
	// MMSizeInt32 screenSize = getMainDisplaySize();
	double velo_x = 0.0, velo_y = 0.0;
	double distance;

	while ((distance =crude_hypot((double)pos.x - endPoint.x, (double)pos.y - endPoint.y)) > 1.0) {
		double gravity = DEADBEEF_UNIFORM(5.0, 500.0);
		// double gravity = DEADBEEF_UNIFORM(lowSpeed, highSpeed);
		double veloDistance;
		velo_x += (gravity * ((double)endPoint.x - pos.x)) / distance;
		velo_y += (gravity * ((double)endPoint.y - pos.y)) / distance;

		/* Normalize velocity to get a unit vector of length 1. */
		veloDistance = crude_hypot(velo_x, velo_y);
		velo_x /= veloDistance;
		velo_y /= veloDistance;

		pos.x += floor(velo_x + 0.5);
		pos.y += floor(velo_y + 0.5);

		/* Make sure we are in the screen boundaries! (Strange things will happen if we are not.) */
		// if (pos.x >= screenSize.w || pos.y >= screenSize.h) {
		// 	return false;
		// }
		moveMouse(pos);

		/* Wait 1 - 3 milliseconds. */
		microsleep(DEADBEEF_UNIFORM(lowSpeed, highSpeed));
		// microsleep(DEADBEEF_UNIFORM(1.0, 3.0));
	}

	return true;
}
