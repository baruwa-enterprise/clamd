// Copyright (C) 2018 Andrew Colin Kissa <andrew@datopdog.io>
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this file,
// You can obtain one at http://mozilla.org/MPL/2.0/.

/*
Package clamd Golang Clamd client
Clamd - Golang clamd client
*/
package clamd

import (
	"fmt"
	"strings"
	"testing"
)

type checkErrorTestKey struct {
	in  string
	out error
}

var s = "Could not open file /.xxxx ERROR"
var errNf = fmt.Errorf("%s", strings.TrimRight(s, " ERROR"))
var TestcheckErrors = []checkErrorTestKey{
	{"This is a test", nil},
	{s, errNf},
}

func TestCheckErrors(t *testing.T) {
	for _, tt := range TestcheckErrors {
		if e := checkError(tt.in); e != tt.out {
			if e != nil && e.Error() != tt.out.Error() {
				t.Errorf("%q = checkError(%q), want %q", tt.out, tt.in, tt.out)
			}
		}
	}
}
