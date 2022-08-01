// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package env

import (
	"io"
)

// Additional routines compiled into the package only during testing.

var DefaultUsage = Usage

// ResetForTesting clears all env state and sets the usage function as directed.
// After calling ResetForTesting, parse errors in env handling will not
// exit the program.
func ResetForTesting(usage func()) {
	Environ = NewEnvSet("", ContinueOnError)
	Environ.SetOutput(io.Discard)
	Environ.Usage = environUsage
	Usage = usage
}
