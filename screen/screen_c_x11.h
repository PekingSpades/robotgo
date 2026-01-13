//#include "../base/os.h"

#include <X11/Xlib.h>
#include <X11/Xresource.h>
// #include "../base/xdisplay_c.h"

intptr scaleX();

double sys_scale(int32_t display_id) {
	Display *dpy = XOpenDisplay(NULL);

	int scr = 0; /* Screen number */
	double xres = ((((double) DisplayWidth(dpy, scr)) * 25.4) /
		((double) DisplayWidthMM(dpy, scr)));

	char *rms = XResourceManagerString(dpy);
	if (rms) {
		XrmDatabase db = XrmGetStringDatabase(rms);
		if (db) {
			XrmValue value;
			char *type = NULL;

			if (XrmGetResource(db, "Xft.dpi", "String", &type, &value)) {
				if (value.addr) {
					xres = atof(value.addr);
				}
			}

			XrmDestroyDatabase(db);
		}
	}
	XCloseDisplay (dpy);

	return xres / 96.0;
}

intptr scaleX(){
	return 0;
}

MMSizeInt32 getMainDisplaySize(void) {
	Display *display = XGetMainDisplay();
	const int screen = DefaultScreen(display);

	return MMSizeInt32Make(
						(int32_t)DisplayWidth(display, screen),
	                	(int32_t)DisplayHeight(display, screen));
}

MMRectInt32 getScreenRect(int32_t display_id) {
	Display *display = XGetMainDisplay();
	const int screen = DefaultScreen(display);

	return MMRectInt32Make(
					(int32_t)0, (int32_t)0,
					(int32_t)DisplayWidth(display, screen),
	                (int32_t)DisplayHeight(display, screen));
}

bool pointVisibleOnMainDisplay(MMPointInt32 point){
	MMSizeInt32 displaySize = getMainDisplaySize();
	return point.x < displaySize.w && point.y < displaySize.h;
}
