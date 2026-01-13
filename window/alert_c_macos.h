// #include "os.h"
#include <CoreFoundation/CoreFoundation.h>

CFStringRef CFStringCreateWithUTF8String(const char *title) {
	if (title == NULL) { return NULL; }
	return CFStringCreateWithCString(NULL, title, kCFStringEncodingUTF8);
}

int showAlert(const char *title, const char *msg,
		const char *defaultButton, const char *cancelButton) {
	CFStringRef alertHeader = CFStringCreateWithUTF8String(title);
	CFStringRef alertMessage = CFStringCreateWithUTF8String(msg);
	CFStringRef defaultButtonTitle = CFStringCreateWithUTF8String(defaultButton);
	CFStringRef cancelButtonTitle = CFStringCreateWithUTF8String(cancelButton);
	CFOptionFlags responseFlags;

	SInt32 err = CFUserNotificationDisplayAlert(
		0.0, kCFUserNotificationNoteAlertLevel, NULL, NULL, NULL, alertHeader, alertMessage,
		defaultButtonTitle, cancelButtonTitle, NULL, &responseFlags);

	if (alertHeader != NULL) CFRelease(alertHeader);
	if (alertMessage != NULL) CFRelease(alertMessage);
	if (defaultButtonTitle != NULL) CFRelease(defaultButtonTitle);
	if (cancelButtonTitle != NULL) CFRelease(cancelButtonTitle);

	if (err != 0) { return -1; }
	return (responseFlags == kCFUserNotificationDefaultResponse) ? 0 : 1;
}
