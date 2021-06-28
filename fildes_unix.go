//+build !windows

// Copyright (C) 2018-2021 Andrew Colin Kissa <andrew@datopdog.io>
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
	"net"
	"net/textproto"
	"os"
	"syscall"

	"github.com/baruwa-enterprise/clamd/protocol"
)

func (c *Client) fildesScan(tc *textproto.Conn, conn net.Conn, p string) (err error) {
	var f *os.File
	var vf *os.File

	fmt.Fprintf(tc.W, "n%s\n", protocol.Fildes)
	tc.W.Flush()

	if f, err = os.Open(p); err != nil {
		return
	}
	defer f.Close()

	s := conn.(*net.UnixConn)
	if vf, err = s.File(); err != nil {
		return
	}
	sock := int(vf.Fd())
	defer vf.Close()

	fds := []int{int(f.Fd())}
	rights := syscall.UnixRights(fds...)
	if err = syscall.Sendmsg(sock, nil, rights, nil, 0); err != nil {
		return
	}

	return
}
