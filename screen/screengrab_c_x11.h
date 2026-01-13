#include "../base/bitmap_free_c.h"
#include <stdlib.h> /* malloc() */

#include <X11/Xlib.h>
#include <X11/Xutil.h>
#include "../base/xdisplay_c.h"
#include "screen_c.h"

MMBitmapRef copyMMBitmapFromDisplayInRect(MMRectInt32 rect, int32_t display_id, int8_t isPid) {
	MMBitmapRef bitmap;
	Display *display;
	if (display_id == -1) {
		display = XOpenDisplay(NULL);
	} else {
		display = XGetMainDisplay();
	}

	MMPointInt32 o = rect.origin; MMSizeInt32 s = rect.size;
	XImage *image = XGetImage(display, XDefaultRootWindow(display),
							(int)o.x, (int)o.y, (unsigned int)s.w, (unsigned int)s.h,
							AllPlanes, ZPixmap);
	XCloseDisplay(display);
	if (image == NULL) { return NULL; }

	bitmap = createMMBitmap_c((uint8_t *)image->data,
				s.w, s.h, (size_t)image->bytes_per_line,
				(uint8_t)image->bits_per_pixel, (uint8_t)image->bits_per_pixel / 8);
	image->data = NULL; /* Steal ownership of bitmap data so we don't have to copy it. */
	XDestroyImage(image);

	return bitmap;
}
