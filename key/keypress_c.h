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

#include "../base/deadbeef_rand_c.h"
#include "../base/microsleep.h"
#include "keypress.h"
#include "keycode_c.h"

#include <ctype.h> /* For isupper() */
#if defined(IS_MACOSX)
	#include <ApplicationServices/ApplicationServices.h>
	#import <IOKit/hidsystem/IOHIDLib.h>
	#import <IOKit/hidsystem/ev_keymap.h>
#elif defined(USE_X11)
	#include <X11/extensions/XTest.h>
#endif

/*
 * Platform-specific helper functions
 */
#if defined(IS_WINDOWS)
	HWND GetHwndByPid(DWORD dwProcessId);

	HWND getHwnd(uintptr pid, int8_t isPid) {
		HWND hwnd = (HWND) pid;
		if (isPid == 0) {
			hwnd = GetHwndByPid(pid);
		}
		return hwnd;
	}

	/* Helper: check if a key is an extended key */
	static inline DWORD getExtendedKeyFlags(int key) {
		switch (key) {
			case VK_RCONTROL:
			case VK_SNAPSHOT: /* Print Screen */
			case VK_RMENU: /* Right Alt / Alt Gr */
			case VK_PAUSE: /* Pause / Break */
			case VK_HOME:
			case VK_UP:
			case VK_PRIOR: /* Page up */
			case VK_LEFT:
			case VK_RIGHT:
			case VK_END:
			case VK_DOWN:
			case VK_NEXT: /* Page Down */
			case VK_INSERT:
			case VK_DELETE:
			case VK_LWIN:
			case VK_RWIN:
			case VK_APPS: /* Application */
			case VK_VOLUME_MUTE:
			case VK_VOLUME_DOWN:
			case VK_VOLUME_UP:
			case VK_MEDIA_NEXT_TRACK:
			case VK_MEDIA_PREV_TRACK:
			case VK_MEDIA_STOP:
			case VK_MEDIA_PLAY_PAUSE:
			case VK_BROWSER_BACK:
			case VK_BROWSER_FORWARD:
			case VK_BROWSER_REFRESH:
			case VK_BROWSER_STOP:
			case VK_BROWSER_SEARCH:
			case VK_BROWSER_FAVORITES:
			case VK_BROWSER_HOME:
			case VK_LAUNCH_MAIL:
				return KEYEVENTF_EXTENDEDKEY;
			default:
				return 0;
		}
	}

	/* Helper: add a key input to an INPUT array */
	static inline void addKeyInput(INPUT *input, int key, DWORD flags) {
		input->type = INPUT_KEYBOARD;
		input->ki.wVk = key;
		input->ki.wScan = MapVirtualKey(key & 0xff, MAPVK_VK_TO_VSC);
		input->ki.dwFlags = flags | getExtendedKeyFlags(key);
		input->ki.time = 0;
		input->ki.dwExtraInfo = 0;
	}

	/* Send key event to a specific window via PostMessage */
	void keyEventToWindow(int key, DWORD flags, uintptr pid, int8_t isPid) {
		HWND hwnd = getHwnd(pid, isPid);
		int msg = (flags & KEYEVENTF_KEYUP) ? WM_KEYUP : WM_KEYDOWN;
		PostMessageW(hwnd, msg, key, 0);
	}

#elif defined(USE_X11)
	Display *XGetMainDisplay(void);
#elif defined(IS_MACOSX)
	static int SendTo(uintptr pid, CGEventRef event) {
		if (pid != 0) {
			CGEventPostToPid(pid, event);
		} else {
			CGEventPost(kCGHIDEventTap, event);
		}
		CFRelease(event);
		return 0;
	}

	static io_connect_t _getAuxiliaryKeyDriver(void) {
		static mach_port_t sEventDrvrRef = 0;
		mach_port_t masterPort, service, iter;
		kern_return_t kr;

		if (!sEventDrvrRef) {
			kr = IOMasterPort(bootstrap_port, &masterPort);
			assert(KERN_SUCCESS == kr);
			kr = IOServiceGetMatchingServices(masterPort, IOServiceMatching(kIOHIDSystemClass), &iter);
			assert(KERN_SUCCESS == kr);

			service = IOIteratorNext(iter);
			assert(service);

			kr = IOServiceOpen(service, mach_task_self(), kIOHIDParamConnectType, &sEventDrvrRef);
			assert(KERN_SUCCESS == kr);

			IOObjectRelease(service);
			IOObjectRelease(iter);
		}
		return sEventDrvrRef;
	}

	/* Helper: post media key event via IOKit HID */
	static int postMediaKeyEvent(MMKeyCode code, bool down) {
		NXEventData event;
		kern_return_t kr;
		IOGPoint loc = { 0, 0 };
		UInt32 evtInfo = code << 16 | (down ? NX_KEYDOWN : NX_KEYUP) << 8;

		bzero(&event, sizeof(NXEventData));
		event.compound.subType = NX_SUBTYPE_AUX_CONTROL_BUTTONS;
		event.compound.misc.L[0] = evtInfo;

		kr = IOHIDPostEvent(_getAuxiliaryKeyDriver(),
			NX_SYSDEFINED, loc, &event, kNXEventDataVersion, 0, FALSE);
		return (kr == KERN_SUCCESS) ? 0 : -1;
	}
#endif

/*
 * keyTap - Atomic key tap (press + release) with modifiers
 *
 * Press order:  Modifiers -> Main key
 * Release order: Main key -> Modifiers (LIFO)
 */
int keyTap(MMKeyCode code, MMKeyFlags flags) {
#if defined(IS_WINDOWS)
	INPUT inputs[10];
	int count = 0;

	/* Press: modifiers -> main key */
	if (flags & MOD_META) { addKeyInput(&inputs[count++], K_META, 0); }
	if (flags & MOD_ALT) { addKeyInput(&inputs[count++], K_ALT, 0); }
	if (flags & MOD_CONTROL) { addKeyInput(&inputs[count++], K_CONTROL, 0); }
	if (flags & MOD_SHIFT) { addKeyInput(&inputs[count++], K_SHIFT, 0); }
	addKeyInput(&inputs[count++], code, 0);

	/* Release: main key -> modifiers (LIFO) */
	addKeyInput(&inputs[count++], code, KEYEVENTF_KEYUP);
	if (flags & MOD_SHIFT) { addKeyInput(&inputs[count++], K_SHIFT, KEYEVENTF_KEYUP); }
	if (flags & MOD_CONTROL) { addKeyInput(&inputs[count++], K_CONTROL, KEYEVENTF_KEYUP); }
	if (flags & MOD_ALT) { addKeyInput(&inputs[count++], K_ALT, KEYEVENTF_KEYUP); }
	if (flags & MOD_META) { addKeyInput(&inputs[count++], K_META, KEYEVENTF_KEYUP); }

	return SendInput(count, inputs, sizeof(INPUT)) == count ? 0 : GetLastError();

#elif defined(USE_X11)
	Display *display = XGetMainDisplay();

	/* Press: modifiers -> main key */
	if (flags & MOD_META) {
		XTestFakeKeyEvent(display, XKeysymToKeycode(display, K_META), True, CurrentTime);
	}
	if (flags & MOD_ALT) {
		XTestFakeKeyEvent(display, XKeysymToKeycode(display, K_ALT), True, CurrentTime);
	}
	if (flags & MOD_CONTROL) {
		XTestFakeKeyEvent(display, XKeysymToKeycode(display, K_CONTROL), True, CurrentTime);
	}
	if (flags & MOD_SHIFT) {
		XTestFakeKeyEvent(display, XKeysymToKeycode(display, K_SHIFT), True, CurrentTime);
	}
	XTestFakeKeyEvent(display, XKeysymToKeycode(display, code), True, CurrentTime);

	/* Release: main key -> modifiers (LIFO) */
	XTestFakeKeyEvent(display, XKeysymToKeycode(display, code), False, CurrentTime);
	if (flags & MOD_SHIFT) {
		XTestFakeKeyEvent(display, XKeysymToKeycode(display, K_SHIFT), False, CurrentTime);
	}
	if (flags & MOD_CONTROL) {
		XTestFakeKeyEvent(display, XKeysymToKeycode(display, K_CONTROL), False, CurrentTime);
	}
	if (flags & MOD_ALT) {
		XTestFakeKeyEvent(display, XKeysymToKeycode(display, K_ALT), False, CurrentTime);
	}
	if (flags & MOD_META) {
		XTestFakeKeyEvent(display, XKeysymToKeycode(display, K_META), False, CurrentTime);
	}

	XSync(display, false);
	return 0;

#elif defined(IS_MACOSX)
	/* The media keys all have 1000 added to them to help us detect them. */
	if (code >= 1000) {
		code = code - 1000; /* Get the real keycode. */
		postMediaKeyEvent(code, true);
		microsleep(5.0);
		postMediaKeyEvent(code, false);
		return 0;
	}

	/* macOS: CGEventFlags makes it atomic - modifiers are set on the event itself */
	CGEventSourceRef source = CGEventSourceCreate(kCGEventSourceStateHIDSystemState);

	/* Press */
	CGEventRef keyDown = CGEventCreateKeyboardEvent(source, (CGKeyCode)code, true);
	if (keyDown == NULL) {
		CFRelease(source);
		return -1;
	}
	if (flags != 0) {
		CGEventSetFlags(keyDown, (CGEventFlags)flags);
	}
	CGEventPost(kCGHIDEventTap, keyDown);
	CFRelease(keyDown);

	/* Release */
	CGEventRef keyUp = CGEventCreateKeyboardEvent(source, (CGKeyCode)code, false);
	if (keyUp == NULL) {
		CFRelease(source);
		return -1;
	}
	if (flags != 0) {
		CGEventSetFlags(keyUp, (CGEventFlags)flags);
	}
	CGEventPost(kCGHIDEventTap, keyUp);
	CFRelease(keyUp);

	CFRelease(source);
	return 0;
#endif
}

/*
 * keyToggle - Atomic key toggle (press or release) with modifiers
 *
 * down=true:  Modifiers -> Main key (press order)
 * down=false: Main key -> Modifiers (release order, LIFO)
 */
int keyToggle(MMKeyCode code, const bool down, MMKeyFlags flags) {
#if defined(IS_WINDOWS)
	INPUT inputs[5];
	int count = 0;
	const DWORD dwFlags = down ? 0 : KEYEVENTF_KEYUP;

	if (down) {
		/* Press: modifiers -> main key */
		if (flags & MOD_META) { addKeyInput(&inputs[count++], K_META, dwFlags); }
		if (flags & MOD_ALT) { addKeyInput(&inputs[count++], K_ALT, dwFlags); }
		if (flags & MOD_CONTROL) { addKeyInput(&inputs[count++], K_CONTROL, dwFlags); }
		if (flags & MOD_SHIFT) { addKeyInput(&inputs[count++], K_SHIFT, dwFlags); }
		addKeyInput(&inputs[count++], code, dwFlags);
	} else {
		/* Release: main key -> modifiers (LIFO) */
		addKeyInput(&inputs[count++], code, dwFlags);
		if (flags & MOD_SHIFT) { addKeyInput(&inputs[count++], K_SHIFT, dwFlags); }
		if (flags & MOD_CONTROL) { addKeyInput(&inputs[count++], K_CONTROL, dwFlags); }
		if (flags & MOD_ALT) { addKeyInput(&inputs[count++], K_ALT, dwFlags); }
		if (flags & MOD_META) { addKeyInput(&inputs[count++], K_META, dwFlags); }
	}

	return SendInput(count, inputs, sizeof(INPUT)) == count ? 0 : GetLastError();

#elif defined(USE_X11)
	Display *display = XGetMainDisplay();
	const Bool is_press = down ? True : False;

	if (down) {
		/* Press: modifiers -> main key */
		if (flags & MOD_META) {
			XTestFakeKeyEvent(display, XKeysymToKeycode(display, K_META), is_press, CurrentTime);
		}
		if (flags & MOD_ALT) {
			XTestFakeKeyEvent(display, XKeysymToKeycode(display, K_ALT), is_press, CurrentTime);
		}
		if (flags & MOD_CONTROL) {
			XTestFakeKeyEvent(display, XKeysymToKeycode(display, K_CONTROL), is_press, CurrentTime);
		}
		if (flags & MOD_SHIFT) {
			XTestFakeKeyEvent(display, XKeysymToKeycode(display, K_SHIFT), is_press, CurrentTime);
		}
		XTestFakeKeyEvent(display, XKeysymToKeycode(display, code), is_press, CurrentTime);
	} else {
		/* Release: main key -> modifiers (LIFO) */
		XTestFakeKeyEvent(display, XKeysymToKeycode(display, code), is_press, CurrentTime);
		if (flags & MOD_SHIFT) {
			XTestFakeKeyEvent(display, XKeysymToKeycode(display, K_SHIFT), is_press, CurrentTime);
		}
		if (flags & MOD_CONTROL) {
			XTestFakeKeyEvent(display, XKeysymToKeycode(display, K_CONTROL), is_press, CurrentTime);
		}
		if (flags & MOD_ALT) {
			XTestFakeKeyEvent(display, XKeysymToKeycode(display, K_ALT), is_press, CurrentTime);
		}
		if (flags & MOD_META) {
			XTestFakeKeyEvent(display, XKeysymToKeycode(display, K_META), is_press, CurrentTime);
		}
	}

	XSync(display, false);
	return 0;

#elif defined(IS_MACOSX)
	/* The media keys all have 1000 added to them to help us detect them. */
	if (code >= 1000) {
		code = code - 1000; /* Get the real keycode. */
		return postMediaKeyEvent(code, down);
	}

	/* macOS: CGEventFlags makes it atomic */
	CGEventSourceRef source = CGEventSourceCreate(kCGEventSourceStateHIDSystemState);
	CGEventRef keyEvent = CGEventCreateKeyboardEvent(source, (CGKeyCode)code, down);

	if (keyEvent == NULL) {
		CFRelease(source);
		return -1;
	}

	CGEventSetType(keyEvent, down ? kCGEventKeyDown : kCGEventKeyUp);
	if (flags != 0) {
		CGEventSetFlags(keyEvent, (CGEventFlags)flags);
	}

	CGEventPost(kCGHIDEventTap, keyEvent);
	CFRelease(keyEvent);
	CFRelease(source);
	return 0;
#endif
}

/*
 * keyTapPid - Key tap to a specific process (non-atomic, uses PostMessage on Windows)
 */
int keyTapPid(MMKeyCode code, MMKeyFlags flags, uintptr pid) {
#if defined(IS_MACOSX)
	/* macOS: supports PID natively */
	CGEventSourceRef source = CGEventSourceCreate(kCGEventSourceStateHIDSystemState);

	CGEventRef keyDown = CGEventCreateKeyboardEvent(source, (CGKeyCode)code, true);
	if (keyDown == NULL) {
		CFRelease(source);
		return -1;
	}
	if (flags != 0) {
		CGEventSetFlags(keyDown, (CGEventFlags)flags);
	}
	CGEventPostToPid(pid, keyDown);
	CFRelease(keyDown);

	CGEventRef keyUp = CGEventCreateKeyboardEvent(source, (CGKeyCode)code, false);
	if (keyUp == NULL) {
		CFRelease(source);
		return -1;
	}
	if (flags != 0) {
		CGEventSetFlags(keyUp, (CGEventFlags)flags);
	}
	CGEventPostToPid(pid, keyUp);
	CFRelease(keyUp);

	CFRelease(source);
	return 0;

#elif defined(IS_WINDOWS)
	/* Windows: use PostMessage (non-atomic) */
	if (flags & MOD_META) { keyEventToWindow(K_META, 0, pid, 0); }
	if (flags & MOD_ALT) { keyEventToWindow(K_ALT, 0, pid, 0); }
	if (flags & MOD_CONTROL) { keyEventToWindow(K_CONTROL, 0, pid, 0); }
	if (flags & MOD_SHIFT) { keyEventToWindow(K_SHIFT, 0, pid, 0); }
	keyEventToWindow(code, 0, pid, 0);

	keyEventToWindow(code, KEYEVENTF_KEYUP, pid, 0);
	if (flags & MOD_SHIFT) { keyEventToWindow(K_SHIFT, KEYEVENTF_KEYUP, pid, 0); }
	if (flags & MOD_CONTROL) { keyEventToWindow(K_CONTROL, KEYEVENTF_KEYUP, pid, 0); }
	if (flags & MOD_ALT) { keyEventToWindow(K_ALT, KEYEVENTF_KEYUP, pid, 0); }
	if (flags & MOD_META) { keyEventToWindow(K_META, KEYEVENTF_KEYUP, pid, 0); }
	return 0;

#elif defined(USE_X11)
	/* X11: no direct PID support, fall back to global */
	return keyTap(code, flags);
#endif
}

/*
 * keyTogglePid - Key toggle to a specific process
 */
int keyTogglePid(MMKeyCode code, const bool down, MMKeyFlags flags, uintptr pid) {
#if defined(IS_MACOSX)
	CGEventSourceRef source = CGEventSourceCreate(kCGEventSourceStateHIDSystemState);
	CGEventRef keyEvent = CGEventCreateKeyboardEvent(source, (CGKeyCode)code, down);

	if (keyEvent == NULL) {
		CFRelease(source);
		return -1;
	}

	CGEventSetType(keyEvent, down ? kCGEventKeyDown : kCGEventKeyUp);
	if (flags != 0) {
		CGEventSetFlags(keyEvent, (CGEventFlags)flags);
	}

	CGEventPostToPid(pid, keyEvent);
	CFRelease(keyEvent);
	CFRelease(source);
	return 0;

#elif defined(IS_WINDOWS)
	DWORD dwFlags = down ? 0 : KEYEVENTF_KEYUP;

	if (down) {
		if (flags & MOD_META) { keyEventToWindow(K_META, dwFlags, pid, 0); }
		if (flags & MOD_ALT) { keyEventToWindow(K_ALT, dwFlags, pid, 0); }
		if (flags & MOD_CONTROL) { keyEventToWindow(K_CONTROL, dwFlags, pid, 0); }
		if (flags & MOD_SHIFT) { keyEventToWindow(K_SHIFT, dwFlags, pid, 0); }
		keyEventToWindow(code, dwFlags, pid, 0);
	} else {
		keyEventToWindow(code, dwFlags, pid, 0);
		if (flags & MOD_SHIFT) { keyEventToWindow(K_SHIFT, dwFlags, pid, 0); }
		if (flags & MOD_CONTROL) { keyEventToWindow(K_CONTROL, dwFlags, pid, 0); }
		if (flags & MOD_ALT) { keyEventToWindow(K_ALT, dwFlags, pid, 0); }
		if (flags & MOD_META) { keyEventToWindow(K_META, dwFlags, pid, 0); }
	}
	return 0;

#elif defined(USE_X11)
	return keyToggle(code, down, flags);
#endif
}

/*
 * Legacy functions for compatibility
 */
#if defined(USE_X11)
	bool toUpper(char c) {
		if (isupper(c)) {
			return true;
		}
		char *special = "~!@#$%^&*()_+{}|:\"<>?";
		while (*special) {
			if (*special == c) {
				return true;
			}
			special++;
		}
		return false;
	}
#endif

void toggleKey(char c, const bool down, MMKeyFlags flags, uintptr pid) {
	MMKeyCode keyCode = keyCodeForChar(c);

	#if defined(USE_X11)
		if (toUpper(c) && !(flags & MOD_SHIFT)) {
			flags |= MOD_SHIFT;
		}
	#else
		if (isupper(c) && !(flags & MOD_SHIFT)) {
			flags |= MOD_SHIFT;
		}
	#endif

	#if defined(IS_WINDOWS)
		int modifiers = keyCode >> 8;
		if ((modifiers & 1) != 0) { flags |= MOD_SHIFT; }
		if ((modifiers & 2) != 0) { flags |= MOD_CONTROL; }
		if ((modifiers & 4) != 0) { flags |= MOD_ALT; }
		keyCode = keyCode & 0xff;
	#endif

	if (pid != 0) {
		keyTogglePid(keyCode, down, flags, pid);
	} else {
		keyToggle(keyCode, down, flags);
	}
}

#if defined(IS_MACOSX)
	void toggleUnicode(UniChar ch, const bool down, uintptr pid) {
		CGEventSourceRef source = CGEventSourceCreate(kCGEventSourceStateHIDSystemState);
		CGEventRef keyEvent = CGEventCreateKeyboardEvent(source, 0, down);
		if (keyEvent == NULL) {
			fputs("Could not create keyboard event.\n", stderr);
			CFRelease(source);
			return;
		}

		CGEventKeyboardSetUnicodeString(keyEvent, 1, &ch);
		SendTo(pid, keyEvent);
		CFRelease(source);
	}
#else
	#define toggleUniKey(c, down) toggleKey(c, down, MOD_NONE, 0)
#endif

void unicodeType(const unsigned value, uintptr pid, int8_t isPid) {
#if defined(IS_MACOSX)
	UniChar ch = (UniChar)value;
	toggleUnicode(ch, true, pid);
	microsleep(5.0);
	toggleUnicode(ch, false, pid);

#elif defined(IS_WINDOWS)
	if (pid != 0) {
		HWND hwnd = getHwnd(pid, isPid);
		PostMessageW(hwnd, WM_CHAR, value, 0);
		return;
	}

	INPUT input[2];
	memset(input, 0, sizeof(input));

	input[0].type = INPUT_KEYBOARD;
	input[0].ki.wVk = 0;
	input[0].ki.wScan = value;
	input[0].ki.dwFlags = 0x4; // KEYEVENTF_UNICODE

	input[1].type = INPUT_KEYBOARD;
	input[1].ki.wVk = 0;
	input[1].ki.wScan = value;
	input[1].ki.dwFlags = KEYEVENTF_KEYUP | 0x4;

	SendInput(2, input, sizeof(INPUT));

#elif defined(USE_X11)
	toggleUniKey(value, true);
	microsleep(5.0);
	toggleUniKey(value, false);
#endif
}

#if defined(USE_X11)
	int input_utf(const char *utf) {
		Display *dpy = XOpenDisplay(NULL);
		KeySym sym = XStringToKeysym(utf);

		int min, max, numcodes;
		XDisplayKeycodes(dpy, &min, &max);
		KeySym *keysym;
		keysym = XGetKeyboardMapping(dpy, min, max-min+1, &numcodes);
		keysym[(max-min-1)*numcodes] = sym;
		XChangeKeyboardMapping(dpy, min, numcodes, keysym, (max-min));
		XFree(keysym);
		XFlush(dpy);

		KeyCode code = XKeysymToKeycode(dpy, sym);
		XTestFakeKeyEvent(dpy, code, True, 1);
		XTestFakeKeyEvent(dpy, code, False, 1);

		XFlush(dpy);
		XCloseDisplay(dpy);
		return 0;
	}
#else
	int input_utf(const char *utf) {
		return 0;
	}
#endif