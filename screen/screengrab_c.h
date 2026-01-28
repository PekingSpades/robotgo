#if defined(IS_MACOSX)
	#include "screengrab_c_macos.h"
#elif defined(USE_X11)
	#include "screengrab_c_x11.h"
#elif defined(IS_WINDOWS)
	#include "screengrab_c_windows.h"
#endif
