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

#ifndef DISPLAY_C_MACOS_H
#define DISPLAY_C_MACOS_H

#include "../base/types.h"
#include <ApplicationServices/ApplicationServices.h>
#include <stdio.h>

// DisplayInfoC contains display information in C struct
typedef struct {
    uintptr handle;     // CGDirectDisplayID
    int32_t index;      // Display index
    int8_t  isMain;     // Is main display
    int32_t x, y;       // Virtual (scaled) coordinates for position
    int32_t w, h;       // Physical pixel size (not virtual)
    double  scale;      // Scale factor (pixel/virtual)
} DisplayInfoC;

// Get display count
static int32_t getDisplayCount() {
    uint32_t count = 0;
    if (CGGetActiveDisplayList(0, NULL, &count) == kCGErrorSuccess) {
        return (int32_t)count;
    }
    return 0;
}

// Get display info by CGDirectDisplayID
static DisplayInfoC getDisplayInfoById(CGDirectDisplayID displayID, int32_t index) {
    DisplayInfoC info = {0};

    // Get virtual bounds (for position in virtual desktop)
    CGRect bounds = CGDisplayBounds(displayID);

    info.handle = (uintptr)displayID;
    info.index = index;
    info.isMain = (displayID == CGMainDisplayID()) ? 1 : 0;

    // Position uses virtual coordinates (for locating display in virtual desktop)
    info.x = (int32_t)bounds.origin.x;
    info.y = (int32_t)bounds.origin.y;

    // Print bounds info
    printf("[Display %d] CGDisplayBounds: x=%.0f, y=%.0f, width=%.0f, height=%.0f\n",
           index, bounds.origin.x, bounds.origin.y, bounds.size.width, bounds.size.height);

    // Get display mode for physical pixel size and scale factor
    // Note: CGDisplayPixelsWide/High returns points (not pixels) despite its name
    CGDisplayModeRef mode = CGDisplayCopyDisplayMode(displayID);
    if (mode != NULL) {
        // Physical pixel resolution from display mode
        size_t pixelWidth = CGDisplayModeGetPixelWidth(mode);
        size_t pixelHeight = CGDisplayModeGetPixelHeight(mode);
        info.w = (int32_t)pixelWidth;
        info.h = (int32_t)pixelHeight;

        // Get virtual (logical points) width/height
        size_t virtualWidth = CGDisplayModeGetWidth(mode);
        size_t virtualHeight = CGDisplayModeGetHeight(mode);

        // Print all the values
        printf("[Display %d] CGDisplayModeGetPixelWidth:  %zu\n", index, pixelWidth);
        printf("[Display %d] CGDisplayModeGetPixelHeight: %zu\n", index, pixelHeight);
        printf("[Display %d] CGDisplayModeGetWidth:       %zu (logical points)\n", index, virtualWidth);
        printf("[Display %d] CGDisplayModeGetHeight:      %zu (logical points)\n", index, virtualHeight);

        // Calculate scale factor (physical pixels / logical points)
        if (virtualWidth > 0) {
            info.scale = (double)info.w / (double)virtualWidth;
        } else {
            info.scale = 1.0;
        }
        printf("[Display %d] Scale factor: %.2f\n", index, info.scale);

        CGDisplayModeRelease(mode);
    } else {
        // Fallback: CGDisplayPixelsWide/High returns points (logical), not physical pixels
        info.w = (int32_t)CGDisplayPixelsWide(displayID);
        info.h = (int32_t)CGDisplayPixelsHigh(displayID);
        info.scale = 1.0;
        printf("[Display %d] Fallback - CGDisplayPixelsWide: %d, CGDisplayPixelsHigh: %d\n",
               index, info.w, info.h);
    }

    return info;
}

// Get all displays info
static int32_t getAllDisplays(DisplayInfoC* displays, int32_t maxCount) {
    uint32_t count = 0;
    CGDirectDisplayID displayIDs[32];

    if (CGGetActiveDisplayList(32, displayIDs, &count) != kCGErrorSuccess) {
        return 0;
    }

    int32_t resultCount = (count < (uint32_t)maxCount) ? (int32_t)count : maxCount;

    for (int32_t i = 0; i < resultCount; i++) {
        displays[i] = getDisplayInfoById(displayIDs[i], i);
    }

    return resultCount;
}

// Get main display info
static DisplayInfoC getMainDisplay() {
    CGDirectDisplayID mainID = CGMainDisplayID();

    // Find main display index
    uint32_t count = 0;
    CGDirectDisplayID displayIDs[32];
    CGGetActiveDisplayList(32, displayIDs, &count);

    int32_t mainIndex = 0;
    for (uint32_t i = 0; i < count; i++) {
        if (displayIDs[i] == mainID) {
            mainIndex = (int32_t)i;
            break;
        }
    }

    return getDisplayInfoById(mainID, mainIndex);
}

// Get display at index
static DisplayInfoC getDisplayAt(int32_t index) {
    uint32_t count = 0;
    CGDirectDisplayID displayIDs[32];

    if (CGGetActiveDisplayList(32, displayIDs, &count) != kCGErrorSuccess) {
        DisplayInfoC empty = {0};
        return empty;
    }

    if (index >= 0 && index < (int32_t)count) {
        return getDisplayInfoById(displayIDs[index], index);
    }

    // Index out of range
    DisplayInfoC empty = {0};
    return empty;
}

#endif /* DISPLAY_C_MACOS_H */
