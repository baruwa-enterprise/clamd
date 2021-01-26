// Copyright (C) 2018-2021 Andrew Colin Kissa <andrew@datopdog.io>
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this file,
// You can obtain one at http://mozilla.org/MPL/2.0/.

/*
Package clamd Golang Clamd client
Clamd - Golang clamd client
*/
package main

// GitCommit is the git commit that was compiled.
// This will be filled in by the compiler.
var GitCommit string

// Version is the main version number that is being run at the moment.
const Version = "0.0.1"

// VersionPrerelease is a pre-release marker for the version.
// If this is "" (empty string) then it means that it is a final release.
// Otherwise, this is a pre-release such as "dev" (in development)
var VersionPrerelease = ""

// BuildDate is the build date
var BuildDate = ""
