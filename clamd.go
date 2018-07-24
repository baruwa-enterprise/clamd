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
	"io"
	"net"
	"net/textproto"
	"os"
	"strings"
	"time"

	"github.com/baruwa-enterprise/clamd/protocol"
)

const (
	defaultTimeout = 15 * time.Second
	defaultSleep   = 1 * time.Second
)

// A Client represents a Clamd client.
type Client struct {
	network     string
	address     string
	connTimeout time.Duration
	connRetries int
	connSleep   time.Duration
	cmdTimeout  time.Duration
}

// SetConnTimeout sets the connection timeout
func (c *Client) SetConnTimeout(t time.Duration) {
	c.connTimeout = t
}

// SetCmdTimeout sets the cmd timeout
func (c *Client) SetCmdTimeout(t time.Duration) {
	c.cmdTimeout = t
}

// SetConnRetries sets the number of times
// connection is retried
func (c *Client) SetConnRetries(s int) {
	if s < 0 {
		s = 0
	}
	c.connRetries = s
}

// SetConnSleep sets the connection retry sleep
// duration in seconds
func (c *Client) SetConnSleep(s time.Duration) {
	c.connSleep = s
}

// Ping sends a ping to the server
func (c *Client) Ping() (b bool, err error) {
	var r string
	if r, err = c.basicCmd(protocol.Ping); err != nil {
		return
	}

	if err = checkError(r); err != nil {
		return
	}

	b = r == "PONG"

	return
}

// Version returns the server version
func (c *Client) Version() (v string, err error) {
	v, err = c.basicCmd(protocol.Version)

	if err = checkError(v); err != nil {
		return
	}

	return
}

// Reload the server
func (c *Client) Reload() (b bool, err error) {
	var r string
	if r, err = c.basicCmd(protocol.Reload); err != nil {
		return
	}

	if err = checkError(r); err != nil {
		return
	}

	b = r == "RELOADING"

	return
}

// Shutdown stops the server
func (c *Client) Shutdown() (err error) {
	_, err = c.basicCmd(protocol.Shutdown)
	return
}

// Scan a file or directory
func (c *Client) Scan() {
}

// ContScan a file or directory
func (c *Client) ContScan() {
}

// MultiScan a file or directory
func (c *Client) MultiScan() {

}

// InStream scan a stream
func (c *Client) InStream() {

}

// Fildes scan a FD
func (c *Client) Fildes() {

}

// Stats returns server stats
func (c *Client) Stats() (s string, err error) {
	s, err = c.basicCmd(protocol.Stats)

	if err = checkError(s); err != nil {
		return
	}

	return
}

// IDSession starts a session
func (c *Client) IDSession() {

}

// End closes a session
func (c *Client) End() {

}

// VersionCmds returns the supported cmds
func (c *Client) VersionCmds() {

}

func (c *Client) dial() (conn net.Conn, err error) {
	d := &net.Dialer{}

	if c.connTimeout > 0 {
		d.Timeout = c.connTimeout
	}

	for i := 0; i <= c.connRetries; i++ {
		conn, err = d.Dial(c.network, c.address)
		if e, ok := err.(net.Error); ok && e.Timeout() {
			time.Sleep(c.connSleep)
			continue
		}
		break
	}
	return
}

func (c *Client) basicCmd(cmd protocol.Command) (r string, err error) {
	var conn net.Conn
	var l []byte
	var b strings.Builder
	var tc *textproto.Conn

	conn, err = c.dial()
	if err != nil {
		return
	}

	if c.cmdTimeout > 0 {
		conn.SetDeadline(time.Now().Add(c.cmdTimeout))
	}

	tc = textproto.NewConn(conn)
	defer tc.Close()

	id := tc.Next()
	tc.StartRequest(id)
	fmt.Fprintf(tc.W, "n%s\n", cmd)
	tc.W.Flush()
	tc.EndRequest(id)

	tc.StartResponse(id)
	defer tc.EndResponse(id)

	if cmd == protocol.Shutdown {
		return
	}

	for {
		l, err = tc.R.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			break
		}
		fmt.Fprintf(&b, "%s", l)
	}

	r = strings.TrimRight(b.String(), "\n")

	return
}

// NewClient returns a new Clamd client.
func NewClient(network, address string) (c *Client, err error) {
	if network == "" && address == "" {
		network = "unix"
		address = "/var/run/clamav/clamd.sock"
	}

	if network == "unix" || network == "unixpacket" {
		if _, err = os.Stat(address); os.IsNotExist(err) {
			err = fmt.Errorf("The unix socket: %s does not exist", address)
			return
		}
	}

	c = &Client{
		network:     network,
		address:     address,
		connTimeout: defaultTimeout,
		connSleep:   defaultSleep,
	}
	return
}

func checkError(s string) (err error) {
	if strings.HasSuffix(s, "ERROR") {
		err = fmt.Errorf("%s", strings.TrimRight(s, " ERROR"))
		return
	}

	return
}
