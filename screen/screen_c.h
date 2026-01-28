#if defined(IS_MACOSX)
	#include "screen_c_macos.h"
#elif defined(USE_X11)
	#include "screen_c_x11.h"
#elif defined(IS_WINDOWS)
	#include "screen_c_windows.h"
#endif
