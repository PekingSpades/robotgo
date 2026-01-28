//#include "../base/os.h"

#include <ApplicationServices/ApplicationServices.h>

intptr scaleX();

double sys_scale(int32_t display_id) {
	CGDirectDisplayID displayID = (CGDirectDisplayID) display_id;
	if (displayID == -1) {
		displayID = CGMainDisplayID();
	}

	CGDisplayModeRef modeRef = CGDisplayCopyDisplayMode(displayID);
	double pixelWidth = CGDisplayModeGetPixelWidth(modeRef);
	double targetWidth = CGDisplayModeGetWidth(modeRef);

	return pixelWidth / targetWidth;
}

intptr scaleX(){
	return 0;
}

MMSizeInt32 getMainDisplaySize(void) {
	CGDirectDisplayID displayID = CGMainDisplayID();
	CGRect displayRect = CGDisplayBounds(displayID);

	CGSize size = displayRect.size;
	return MMSizeInt32Make((int32_t)size.width, (int32_t)size.height);
}

MMRectInt32 getScreenRect(int32_t display_id) {
	CGDirectDisplayID displayID = (CGDirectDisplayID) display_id;
	if (display_id == -1) {
		displayID = CGMainDisplayID();
	}
	CGRect displayRect = CGDisplayBounds(displayID);

	CGPoint point = displayRect.origin;
	CGSize size = displayRect.size;
	return MMRectInt32Make(
		(int32_t)point.x, (int32_t)point.y,
		(int32_t)size.width, (int32_t)size.height);
}

bool pointVisibleOnMainDisplay(MMPointInt32 point){
	MMSizeInt32 displaySize = getMainDisplaySize();
	return point.x < displaySize.w && point.y < displaySize.h;
}
