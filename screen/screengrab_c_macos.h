#include "../base/bitmap_free_c.h"
#include <stdlib.h> /* malloc() */

#include <OpenGL/OpenGL.h>
#include <OpenGL/gl.h>
#include <ApplicationServices/ApplicationServices.h>
#if __has_include(<ScreenCaptureKit/ScreenCaptureKit.h>)
	#include <ScreenCaptureKit/ScreenCaptureKit.h>
	#define HAS_SCREENCAPTUREKIT 1
#endif
#include "screen_c.h"

#if defined(HAS_SCREENCAPTUREKIT)
	// ScreenCaptureKit API - requires macOS 14.4+
	static CGImageRef capture15(CGDirectDisplayID id, CGRect diIntersectDisplayLocal, CGColorSpaceRef colorSpace) API_AVAILABLE(macos(14.4)) {
		dispatch_semaphore_t semaphore = dispatch_semaphore_create(0);
		__block CGImageRef image1 = nil;
		[SCShareableContent getShareableContentWithCompletionHandler:^(SCShareableContent* content, NSError* error) {
			@autoreleasepool {
				if (error) {
					dispatch_semaphore_signal(semaphore);
					return;
				}

				SCDisplay* target = nil;
				for (SCDisplay *display in content.displays) {
					if (display.displayID == id) {
						target = display;
						break;
					}
				}
				if (!target) {
					dispatch_semaphore_signal(semaphore);
					return;
				}

				SCContentFilter* filter = [[SCContentFilter alloc] initWithDisplay:target excludingWindows:@[]];
				SCStreamConfiguration* config = [[SCStreamConfiguration alloc] init];
				config.queueDepth = 5;
				config.sourceRect = diIntersectDisplayLocal;
				config.width = diIntersectDisplayLocal.size.width * sys_scale(id);
				config.height = diIntersectDisplayLocal.size.height * sys_scale(id);
				config.scalesToFit = false;
				config.captureResolution = 1;

				[SCScreenshotManager captureImageWithFilter:filter
					configuration:config
					completionHandler:^(CGImageRef img, NSError* error) {
						if (!error) {
							image1 = CGImageCreateCopyWithColorSpace(img, colorSpace);
						}
						dispatch_semaphore_signal(semaphore);
				}];
			}
		}];

		dispatch_semaphore_wait(semaphore, DISPATCH_TIME_FOREVER);
		dispatch_release(semaphore);
		return image1;
	}
#endif

MMBitmapRef copyMMBitmapFromDisplayInRect(MMRectInt32 rect, int32_t display_id, int8_t isPid) {
	MMBitmapRef bitmap = NULL;
	uint8_t *buffer = NULL;
	size_t bufferSize = 0;

	CGDirectDisplayID displayID = (CGDirectDisplayID) display_id;
	if (displayID == -1 || displayID == 0) {
		displayID = CGMainDisplayID();
	}

	MMPointInt32 o = rect.origin; MMSizeInt32 s = rect.size;

	// Runtime version check instead of compile-time macro
	CGImageRef image = NULL;
#ifdef HAS_SCREENCAPTUREKIT
	if (@available(macOS 14.4, *)) {
		// macOS 14.4+: use ScreenCaptureKit (CGDisplayCreateImageForRect is deprecated)
		CGColorSpaceRef color = CGColorSpaceCreateWithName(kCGColorSpaceSRGB);
		image = capture15(displayID, CGRectMake(o.x, o.y, s.w, s.h), color);
		CGColorSpaceRelease(color);
	}
#if __MAC_OS_X_VERSION_MIN_REQUIRED < 150000
	else {
		// Deployment target < macOS 15: legacy API available
		image = CGDisplayCreateImageForRect(displayID, CGRectMake(o.x, o.y, s.w, s.h));
	}
#endif
#else
	// SDK without ScreenCaptureKit: use legacy API
	image = CGDisplayCreateImageForRect(displayID, CGRectMake(o.x, o.y, s.w, s.h));
#endif
	if (!image) { return NULL; }

	CFDataRef imageData = CGDataProviderCopyData(CGImageGetDataProvider(image));
	if (!imageData) { return NULL; }

	bufferSize = CFDataGetLength(imageData);
	buffer = malloc(bufferSize);
	CFDataGetBytes(imageData, CFRangeMake(0, bufferSize), buffer);

	bitmap = createMMBitmap_c(buffer,
			CGImageGetWidth(image), CGImageGetHeight(image), CGImageGetBytesPerRow(image),
			CGImageGetBitsPerPixel(image), CGImageGetBitsPerPixel(image) / 8);

	CFRelease(imageData);
	CGImageRelease(image);

	return bitmap;
}
