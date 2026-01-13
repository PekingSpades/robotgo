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

#ifndef DISPLAY_C_WINDOWS_H
#define DISPLAY_C_WINDOWS_H

#include "../base/types.h"
#include <windows.h>

// DisplayInfoC contains display information in C struct
typedef struct {
    uintptr handle;     // HMONITOR handle
    int32_t index;      // Display index
    int8_t  isMain;     // Is main display
    int32_t x, y, w, h; // Physical coordinates and size
    double  scale;      // Scale factor (physical/logical)
} DisplayInfoC;

// EnumDisplayContext is the enumeration context
typedef struct {
    int32_t       targetIndex;   // Target index (-1 means enumerate all)
    int32_t       currentIndex;  // Current index
    DisplayInfoC* displays;      // Output array
    int32_t       maxCount;      // Max count
    int32_t       foundCount;    // Found count
} EnumDisplayContext;

// Get monitor real physical size (bypassing DPI virtualization)
// Reference: screenshot library's getMonitorRealSize implementation
static int getMonitorRealSize(HMONITOR hMonitor, RECT* outRect) {
    // Step 1: Get device name via GetMonitorInfoW
    MONITORINFOEXW info = {0};
    info.cbSize = sizeof(info);

    if (!GetMonitorInfoW(hMonitor, (MONITORINFO*)&info)) {
        return 0;  // Failed, caller should use logical coordinates as fallback
    }

    // Step 2: Get physical resolution via EnumDisplaySettingsW
    DEVMODEW devMode = {0};
    devMode.dmSize = sizeof(devMode);

    if (!EnumDisplaySettingsW(info.szDevice, ENUM_CURRENT_SETTINGS, &devMode)) {
        return 0;  // Failed
    }

    // Step 3: Build physical coordinate rect
    outRect->left   = devMode.dmPosition.x;
    outRect->top    = devMode.dmPosition.y;
    outRect->right  = devMode.dmPosition.x + (LONG)devMode.dmPelsWidth;
    outRect->bottom = devMode.dmPosition.y + (LONG)devMode.dmPelsHeight;

    return 1;  // Success
}

// Monitor enumeration callback for counting
static BOOL CALLBACK countMonitorCallback(HMONITOR hMonitor, HDC hdcMonitor, LPRECT lprcMonitor, LPARAM dwData) {
    int32_t *count = (int32_t*)dwData;
    (*count)++;
    return TRUE;
}

// Monitor enumeration callback for getting info
static BOOL CALLBACK MonitorInfoEnumProc(HMONITOR hMonitor, HDC hdcMonitor, LPRECT lprcMonitor, LPARAM dwData) {
    EnumDisplayContext* ctx = (EnumDisplayContext*)dwData;

    if (ctx->currentIndex >= ctx->maxCount) {
        return FALSE; // Stop enumeration
    }

    // Get physical size
    RECT physRect;
    int hasPhysical = getMonitorRealSize(hMonitor, &physRect);

    // Get monitor info (to check if main display)
    MONITORINFO mi = {0};
    mi.cbSize = sizeof(mi);
    GetMonitorInfoW(hMonitor, &mi);

    DisplayInfoC* info = &ctx->displays[ctx->currentIndex];
    info->handle = (uintptr)hMonitor;
    info->index = ctx->currentIndex;
    info->isMain = (mi.dwFlags & MONITORINFOF_PRIMARY) ? 1 : 0;

    if (hasPhysical) {
        // Use physical coordinates
        info->x = physRect.left;
        info->y = physRect.top;
        info->w = physRect.right - physRect.left;
        info->h = physRect.bottom - physRect.top;
        // Calculate scale factor
        int32_t logicalW = lprcMonitor->right - lprcMonitor->left;
        if (logicalW > 0) {
            info->scale = (double)info->w / (double)logicalW;
        } else {
            info->scale = 1.0;
        }
    } else {
        // Fallback: use logical coordinates
        info->x = lprcMonitor->left;
        info->y = lprcMonitor->top;
        info->w = lprcMonitor->right - lprcMonitor->left;
        info->h = lprcMonitor->bottom - lprcMonitor->top;
        info->scale = 1.0;
    }

    ctx->currentIndex++;
    ctx->foundCount++;
    return TRUE;  // Continue enumeration
}

// Get display count
static int32_t getDisplayCount() {
    int32_t count = 0;
    EnumDisplayMonitors(NULL, NULL, countMonitorCallback, (LPARAM)&count);
    return count;
}

// Get all displays info
static int32_t getAllDisplays(DisplayInfoC* displays, int32_t maxCount) {
    EnumDisplayContext ctx = {0};
    ctx.targetIndex = -1;
    ctx.currentIndex = 0;
    ctx.displays = displays;
    ctx.maxCount = maxCount;
    ctx.foundCount = 0;

    EnumDisplayMonitors(NULL, NULL, MonitorInfoEnumProc, (LPARAM)&ctx);
    return ctx.foundCount;
}

// Get main display info
static DisplayInfoC getMainDisplay() {
    DisplayInfoC displays[32];
    int32_t count = getAllDisplays(displays, 32);

    for (int32_t i = 0; i < count; i++) {
        if (displays[i].isMain) {
            return displays[i];
        }
    }

    // Fallback: return first display
    if (count > 0) {
        return displays[0];
    }

    // No display found
    DisplayInfoC empty = {0};
    return empty;
}

// Get display at index
static DisplayInfoC getDisplayAt(int32_t index) {
    DisplayInfoC displays[32];
    int32_t count = getAllDisplays(displays, 32);

    if (index >= 0 && index < count) {
        return displays[index];
    }

    // Index out of range
    DisplayInfoC empty = {0};
    return empty;
}

#endif /* DISPLAY_C_WINDOWS_H */
