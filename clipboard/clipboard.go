// Copyright 2013 @atotto. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Package clipboard read/write on clipboard
*/
package clipboard

import (
	"errors"

	"golang.design/x/clipboard"
)

// ErrUnsupported is returned when clipboard is not available
var ErrUnsupported = errors.New("clipboard: unsupported platform or not initialized")

// Unsupported might be set true during clipboard init,
// to help callers decide whether or not to
// offer clipboard options.
var Unsupported bool

func init() {
	err := clipboard.Init()
	if err != nil {
		Unsupported = true
	}
}

// ReInit re-initializes the clipboard, useful for retrying after init failure
func ReInit() error {
	err := clipboard.Init()
	if err != nil {
		Unsupported = true
		return err
	}
	Unsupported = false
	return nil
}

// ReadAll read string from clipboard
func ReadAll() (string, error) {
	if Unsupported {
		return "", ErrUnsupported
	}
	data := clipboard.Read(clipboard.FmtText)
	return string(data), nil
}

// WriteAll write string to clipboard
func WriteAll(text string) error {
	if Unsupported {
		return ErrUnsupported
	}
	clipboard.Write(clipboard.FmtText, []byte(text))
	return nil
}
