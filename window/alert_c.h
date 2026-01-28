#if defined(IS_MACOSX)
	#include "alert_c_macos.h"
#elif defined(USE_X11)
	#include "alert_c_x11.h"
#else
	#include "alert_c_windows.h"
#endif
