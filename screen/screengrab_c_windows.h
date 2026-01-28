#include "../base/bitmap_free_c.h"
#include <stdlib.h> /* malloc() */

#include <string.h>
#include "screen_c.h"

MMBitmapRef copyMMBitmapFromDisplayInRect(MMRectInt32 rect, int32_t display_id, int8_t isPid) {
	MMBitmapRef bitmap;
	void *data;
	HDC screen = NULL, screenMem = NULL;
	HBITMAP dib;
	BITMAPINFO bi;

	int32_t x = rect.origin.x, y = rect.origin.y;
	int32_t w = rect.size.w, h = rect.size.h;

	/* Initialize bitmap info. */
	bi.bmiHeader.biSize = sizeof(bi.bmiHeader);
	bi.bmiHeader.biWidth = (long) w;
	bi.bmiHeader.biHeight = -(long) h; /* Non-cartesian, please */
	bi.bmiHeader.biPlanes = 1;
	bi.bmiHeader.biBitCount = 32;
	bi.bmiHeader.biCompression = BI_RGB;
	bi.bmiHeader.biSizeImage = (DWORD)(4 * w * h);
	bi.bmiHeader.biXPelsPerMeter = 0;
	bi.bmiHeader.biYPelsPerMeter = 0;
	bi.bmiHeader.biClrUsed = 0;
	bi.bmiHeader.biClrImportant = 0;

	HWND hwnd;
	if (display_id == -1 || isPid == 0) {
	// 	screen = GetDC(NULL); /* Get entire screen */
		hwnd = GetDesktopWindow();
	} else {
		hwnd = (HWND) (uintptr) display_id;
	}
	screen = GetDC(hwnd);

	if (screen == NULL) { return NULL; }

	// Todo: Use DXGI
	screenMem = CreateCompatibleDC(screen);
	/* Get screen data in display device context. */
	dib = CreateDIBSection(screen, &bi, DIB_RGB_COLORS, &data, NULL, 0);

	/* Copy the data into a bitmap struct. */
	BOOL b = (screenMem == NULL) ||
		SelectObject(screenMem, dib) == NULL ||
	    !BitBlt(screenMem, (int)0, (int)0, (int)w, (int)h, screen, x, y, SRCCOPY);
	if (b) {
		/* Error copying data. */
		ReleaseDC(hwnd, screen);
		DeleteObject(dib);
		if (screenMem != NULL) { DeleteDC(screenMem); }

		return NULL;
	}

	bitmap = createMMBitmap_c(NULL, w, h, 4 * w, (uint8_t)bi.bmiHeader.biBitCount, 4);

	/* Copy the data to our pixel buffer. */
	if (bitmap != NULL) {
		bitmap->imageBuffer = malloc(bitmap->bytewidth * bitmap->height);
		memcpy(bitmap->imageBuffer, data, bitmap->bytewidth * bitmap->height);
	}

	ReleaseDC(hwnd, screen);
	DeleteObject(dib);
	DeleteDC(screenMem);

	return bitmap;
}
