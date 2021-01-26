// Copyright (C) 2018-2021 Andrew Colin Kissa <andrew@datopdog.io>
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this file,
// You can obtain one at http://mozilla.org/MPL/2.0/.

/*
Package protocol Golang Clamd client
Clamd - Golang clamd client
*/
package protocol

import (
	"testing"
)

type CommandTestKey struct {
	in  Command
	out string
}

type RequiresParamsTestKey struct {
	in  Command
	out bool
}

var NonExistant Command = 20

var TestCommands = []CommandTestKey{
	{Ping, "PING"},
	{Version, "VERSION"},
	{Reload, "RELOAD"},
	{Shutdown, "SHUTDOWN"},
	{Scan, "SCAN"},
	{ContScan, "CONTSCAN"},
	{MultiScan, "MULTISCAN"},
	{Instream, "INSTREAM"},
	{Fildes, "FILDES"},
	{Stats, "STATS"},
	{IDSession, "IDSESSION"},
	{EndSession, "END"},
	{VersionCmds, "VERSIONCOMMANDS"},
	{NonExistant, ""},
}

var TestCommandRequiresParams = []RequiresParamsTestKey{
	{Ping, false},
	{Version, false},
	{Reload, false},
	{Shutdown, false},
	{Scan, true},
	{ContScan, true},
	{MultiScan, true},
	{Instream, true},
	{Fildes, true},
	{Stats, false},
	{IDSession, false},
	{EndSession, false},
	{VersionCmds, false},
}

func TestCommand(t *testing.T) {
	for _, tt := range TestCommands {
		if s := tt.in.String(); s != tt.out {
			t.Errorf("%q.String() = %q, want %q", tt.in, s, tt.out)
		}
	}
}

func TestCommandRequiresParam(t *testing.T) {
	for _, tt := range TestCommandRequiresParams {
		if b := tt.in.RequiresParam(); b != tt.out {
			t.Errorf("%q.RequiresParam() = %t, want %t", tt.in, b, tt.out)
		}
	}
}
