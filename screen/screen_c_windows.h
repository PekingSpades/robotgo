//#include "../base/os.h"

intptr scaleX();

double sys_scale(int32_t display_id) {
	double s = scaleX() / 96.0;
	return s;
}

intptr scaleX(){
	// Get desktop dc
	HDC desktopDc = GetDC(NULL);
	// Get native resolution
	intptr horizontalDPI = GetDeviceCaps(desktopDc, LOGPIXELSX);
	return horizontalDPI;
}

MMSizeInt32 getMainDisplaySize(void) {
	return MMSizeInt32Make(
						(int32_t)GetSystemMetrics(SM_CXSCREEN),
		                (int32_t)GetSystemMetrics(SM_CYSCREEN));
}

MMRectInt32 getScreenRect(int32_t display_id) {
	if (GetSystemMetrics(SM_CMONITORS) == 1
			|| display_id == -1 || display_id == 0) {
		return MMRectInt32Make(
						(int32_t)0,
						(int32_t)0,
			 			(int32_t)GetSystemMetrics(SM_CXSCREEN),
		                (int32_t)GetSystemMetrics(SM_CYSCREEN));
	} else {
		return MMRectInt32Make(
			 			(int32_t)GetSystemMetrics(SM_XVIRTUALSCREEN),
						(int32_t)GetSystemMetrics(SM_YVIRTUALSCREEN),
						(int32_t)GetSystemMetrics(SM_CXVIRTUALSCREEN),
		                (int32_t)GetSystemMetrics(SM_CYVIRTUALSCREEN));
	}
}

bool pointVisibleOnMainDisplay(MMPointInt32 point){
	MMSizeInt32 displaySize = getMainDisplaySize();
	return point.x < displaySize.w && point.y < displaySize.h;
}
