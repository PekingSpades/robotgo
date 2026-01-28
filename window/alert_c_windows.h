// #include "os.h"

int showAlert(const char *title, const char *msg,
		const char *defaultButton, const char *cancelButton) {
	/* TODO: Display custom buttons instead of the pre-defined "OK" and "Cancel". */
	int response = MessageBox(NULL, msg, title,
							(cancelButton == NULL) ? MB_OK : MB_OKCANCEL );
	return (response == IDOK) ? 0 : 1;
}
