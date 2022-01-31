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
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"path"
	"strings"
	"testing"
	"testing/iotest"
	"time"
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

func getaddr() (address, network string, err error) {
	address = os.Getenv("CLAMD_ADDRESS")
	if address == "" {
		address = "/opt/local/var/run/clamav/clamd.socket"
		if _, err = os.Stat(address); os.IsNotExist(err) {
			address = "/var/run/clamav/clamd.ctl"
		}
		if _, err = os.Stat(address); os.IsNotExist(err) {
			address = "/run/clamav/clamd.ctl"
		}
		if _, err = os.Stat(address); os.IsNotExist(err) {
			return
		}
	}
	if strings.HasPrefix(address, "/") {
		network = "unix"
	} else {
		network = "tcp4"
	}
	return
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

func TestBasics(t *testing.T) {
	// Test Non existent socket
	var expected string

	testSock := "/tmp/.dumx.sock"
	_, e := NewClient("unix", testSock)
	if e == nil {
		t.Fatalf("An error should be returned as sock does not exist")
	}
	expected = fmt.Sprintf(unixSockErr, testSock)
	if e.Error() != expected {
		t.Errorf("Expected %q want %q", expected, e)
	}

	// Test defaults
	_, e = NewClient("", "")
	if e == nil {
		t.Fatalf("An error should be returned as sock does not exist")
	}
	expected = fmt.Sprintf(unixSockErr, defaultSock)
	if e.Error() != expected {
		t.Errorf("Got %q want %q", expected, e)
	}

	// Test udp
	proto := "udp"
	_, e = NewClient(proto, "127.1.1.1:3310")
	if e == nil {
		t.Fatalf("Expected an error got nil")
	}
	expected = fmt.Sprintf(unsupportedProtoErr, proto)
	if e.Error() != expected {
		t.Errorf("Got %q want %q", expected, e)
	}

	// Test tcp
	network := "tcp"
	address := "127.1.1.1:3310"
	c, e := NewClient(network, address)
	if e != nil {
		t.Fatalf("An error should not be returned")
	}
	if c.network != network {
		t.Errorf("Got %q want %q", c.network, network)
	}
	if c.address != address {
		t.Errorf("Got %q want %q", c.address, address)
	}
	// Test Fildes
	ctx := context.Background()
	if _, e = c.Fildes(ctx, "/tmp"); e == nil {
		t.Fatalf("An error should be returned")
	}
	if e.Error() != fldesErr {
		t.Errorf("Got %q want %q", e, fldesErr)
	}

}

func TestSettings(t *testing.T) {
	var e error
	var c *Client
	network := "tcp"
	address := "127.1.1.1:3310"
	if c, e = NewClient(network, address); e != nil {
		t.Fatalf("An error should not be returned")
	}
	if c.connTimeout != defaultTimeout {
		t.Errorf("The default conn timeout should be set")
	}
	if c.connSleep != defaultSleep {
		t.Errorf("The default conn sleep should be set")
	}
	if c.connRetries != 0 {
		t.Errorf("The default conn retries should be set")
	}
	expected := 2 * time.Second
	c.SetConnTimeout(expected)
	if c.connTimeout != expected {
		t.Errorf("Calling c.SetConnTimeout(%q) failed", expected)
	}
	c.SetCmdTimeout(expected)
	if c.cmdTimeout != expected {
		t.Errorf("Calling c.SetCmdTimeout(%q) failed", expected)
	}
	c.SetConnSleep(expected)
	if c.connSleep != expected {
		t.Errorf("Calling c.SetConnSleep(%q) failed", expected)
	}
	c.SetConnRetries(2)
	if c.connRetries != 2 {
		t.Errorf("Calling c.SetConnRetries(%q) failed", 2)
	}
	c.SetConnRetries(-2)
	if c.connRetries != 0 {
		t.Errorf("Preventing negative values in c.SetConnRetries(%q) failed", -2)
	}
}

func TestMethodsErrors(t *testing.T) {
	var e error
	var c *Client
	network := "tcp"
	address := "127.1.1.1:3310"
	if c, e = NewClient(network, address); e != nil {
		t.Fatalf("An error should not be returned")
	}
	c.SetConnTimeout(500 * time.Microsecond)
	ctx := context.Background()
	// c.SetConnRetries(1)
	if _, e = c.Ping(ctx); e == nil {
		t.Fatalf("An error should be returned")
	}
	if _, ok := e.(*net.OpError); !ok {
		t.Errorf("Expected *net.OpError want %q", e)
	}

	if _, e = c.Version(ctx); e == nil {
		t.Fatalf("An error should be returned")
	}
	if _, ok := e.(*net.OpError); !ok {
		t.Errorf("Expected *net.OpError want %q", e)
	}

	if _, e = c.Reload(ctx); e == nil {
		t.Fatalf("An error should be returned")
	}
	if _, ok := e.(*net.OpError); !ok {
		t.Errorf("Expected *net.OpError want %q", e)
	}

	if e = c.Shutdown(ctx); e == nil {
		t.Fatalf("An error should be returned")
	}
	if _, ok := e.(*net.OpError); !ok {
		t.Errorf("Expected *net.OpError want %q", e)
	}

	if _, e = c.Stats(ctx); e == nil {
		t.Fatalf("An error should be returned")
	}
	if _, ok := e.(*net.OpError); !ok {
		t.Errorf("Expected *net.OpError want %q", e)
	}

	if _, e = c.VersionCmds(ctx); e == nil {
		t.Fatalf("An error should be returned")
	}
	if _, ok := e.(*net.OpError); !ok {
		t.Errorf("Expected *net.OpError want %q", e)
	}

	if _, e = c.Scan(ctx, "/tmp/bxx.syx"); e == nil {
		t.Fatalf("An error should be returned")
	}
	if _, ok := e.(*net.OpError); !ok {
		t.Errorf("Expected *net.OpError want %q", e)
	}

	if _, e = c.ContScan(ctx, "/tmp/bxx.syx"); e == nil {
		t.Fatalf("An error should be returned")
	}
	if _, ok := e.(*net.OpError); !ok {
		t.Errorf("Expected *net.OpError want %q", e)
	}

	if _, e = c.MultiScan(ctx, "/tmp/bxx.syx"); e == nil {
		t.Fatalf("An error should be returned")
	}
	if _, ok := e.(*net.OpError); !ok {
		t.Errorf("Expected *net.OpError want %q", e)
	}

	expected := "stat /tmp/bxx.syx: no such file or directory"
	if _, e = c.InStream(ctx, "/tmp/bxx.syx"); e == nil {
		t.Fatalf("An error should be returned")
	}
	if e.Error() != expected {
		t.Errorf("Got %q want %q", e, expected)
	}

	if _, e = c.Fildes(ctx, "/tmp/bxx.syx"); e == nil {
		t.Fatalf("An error should be returned")
	}
	if e.Error() != expected {
		t.Errorf("Got %q want %q", e, expected)
	}

	if address, network, e = getaddr(); e != nil {
		t.Fatalf("An error should not be returned")
	}

	if c, e = NewClient(network, address); e != nil {
		t.Fatalf("An error should not be returned")
	}
	c.SetConnTimeout(500 * time.Microsecond)

	if _, e = c.ScanReader(ctx, iotest.ErrReader(io.ErrUnexpectedEOF)); e != io.ErrUnexpectedEOF {
		t.Errorf("Expected ErrUnexpectedEOF got %q", e)
	}
}

func TestMethods(t *testing.T) {
	var e error
	var b bool
	var c *Client
	var f *os.File
	var result []*Response
	var vcmds []string
	var network, address, rs, dir string

	if address, network, e = getaddr(); e != nil {
		t.Fatalf("An error should not be returned")
	}

	fn := "./examples/eicar.txt"
	zfn := "./examples/eicar.tar.bz2"
	dir, e = ioutil.TempDir("", "")
	if e != nil {
		t.Errorf("Temp directory creation failed")
	}
	defer os.RemoveAll(dir)
	if e = os.Chmod(dir, 0755); e != nil {
		t.Errorf("Temp directory chmod failed")
	}
	tfn := path.Join(dir, "eicar.txt")
	tzfn := path.Join(dir, "eicar.tar.bz2")
	if e = copyFile(fn, tfn, 0644); e != nil {
		t.Errorf("Copy %s => %s failed: %t", fn, tfn, e)
	}
	if e = copyFile(zfn, tzfn, 0644); e != nil {
		t.Errorf("Copy %s => %s failed: %t", fn, tzfn, e)
	}

	if c, e = NewClient(network, address); e != nil {
		t.Fatalf("An error should not be returned")
	}
	ctx := context.Background()
	if b, e = c.Ping(ctx); e != nil {
		t.Fatalf("An error should not be returned")
	}
	if !b {
		t.Errorf("Expected %t got %t", true, b)
	}

	if rs, e = c.Version(ctx); e != nil {
		t.Fatalf("An error should not be returned")
	}
	if !strings.HasPrefix(rs, "Clam") {
		t.Errorf("Expected version starting with Clam, got %q", rs)
	}

	if rs, e = c.Stats(ctx); e != nil {
		t.Fatalf("An error should not be returned")
	}
	if !strings.HasPrefix(rs, "POOLS:") {
		t.Errorf("Expected version starting with POOLS:, got %q", rs)
	}

	if vcmds, e = c.VersionCmds(ctx); e != nil {
		t.Fatalf("An error should not be returned")
	}
	if len(vcmds) == 0 {
		t.Fatalf("Expected a slice of strings:, got %q", vcmds)
	}
	if vcmds[0] != "SCAN" {
		t.Errorf("Expected SCAN:, got %q", vcmds[0])
	}

	if result, e = c.Scan(ctx, tfn); e != nil {
		t.Fatalf("Expected nil got %q", e)
	}
	l := len(result)
	if l == 0 {
		t.Errorf("Expected a slice of Response objects:, got %v", result)
	} else if l > 1 {
		t.Errorf("Expected a slice of Response 1 object:, got %d", l)
	} else {
		mb := result[0]
		if mb.Filename != tfn {
			t.Errorf("Expected %q, got %q", tfn, mb.Filename)
		}
		if mb.Signature != "Eicar-Signature" {
			t.Errorf("Expected %q, got %q", "Eicar-Signature", mb.Signature)
		}
	}

	if f, e = os.Open(tfn); e != nil {
		t.Fatalf("Expected nil got %q", e)
	}
	defer f.Close()
	if result, e = c.ScanReader(ctx, f); e != nil {
		t.Errorf("Expected nil got %q", e)
	}
	l = len(result)
	if l == 0 {
		t.Errorf("Expected a slice of Response objects:, got %v", result)
	} else if l > 1 {
		t.Errorf("Expected a slice of Response 1 object:, got %d", l)
	} else {
		mb := result[0]
		if mb.Filename != "stream" {
			t.Errorf("Expected %q, got %q", "stream", mb.Filename)
		}
		if mb.Signature != "Eicar-Signature" {
			t.Errorf("Expected %q, got %q", "Eicar-Signature", mb.Signature)
		}
	}

	if result, e = c.ContScan(ctx, path.Dir(tfn)); e != nil {
		t.Fatalf("Expected nil got %q", e)
	}
	l = len(result)
	if l == 0 {
		t.Errorf("Expected a slice of Response objects:, got %v", result)
	} else if l > 2 {
		t.Errorf("Expected a slice of Response 2 objects:, got %d", l)
	} else {
		mb := result[0]
		if mb.Filename != tfn {
			t.Errorf("Expected %q, got %q", tfn, mb.Filename)
		}
		if mb.Signature != "Eicar-Signature" {
			t.Errorf("Expected %q, got %q", "Eicar-Signature", mb.Signature)
		}
		mb = result[1]
		if mb.Signature != "Eicar-Signature" {
			t.Errorf("Expected %q, got %q", "Eicar-Signature", mb.Signature)
		}
	}

	if result, e = c.MultiScan(ctx, tfn); e != nil {
		t.Fatalf("Expected nil got %q", e)
	}
	l = len(result)
	if l == 0 {
		t.Errorf("Expected a slice of Response objects:, got %v", result)
	} else if l > 1 {
		t.Errorf("Expected a slice of Response 1 object:, got %q", l)
	} else {
		mb := result[0]
		if mb.Filename != tfn {
			t.Errorf("Expected %q, got %q", tfn, mb.Filename)
		}
		if mb.Signature != "Eicar-Signature" {
			t.Errorf("Expected %q, got %q", "Eicar-Signature", mb.Signature)
		}
	}

	if result, e = c.InStream(ctx, fn); e != nil {
		t.Fatalf("An error should not be returned")
	}
	l = len(result)
	if l == 0 {
		t.Errorf("Expected a slice of Response objects:, got %v", result)
	} else if l > 1 {
		t.Errorf("Expected a slice of Response 1 object:, got %q", l)
	} else {
		mb := result[0]
		if mb.Filename != "stream" {
			t.Errorf("Expected %q, got %q", "stream", mb.Filename)
		}
		if mb.Signature != "Eicar-Signature" {
			t.Errorf("Expected %q, got %q", "Eicar-Signature", mb.Signature)
		}
	}

	if network == "unix" {
		if result, e = c.Fildes(ctx, fn); e != nil {
			t.Fatalf("An error should not be returned")
		}
		l := len(result)
		if l == 0 {
			t.Errorf("Expected a slice of Response objects:, got %v", result)
		} else if l > 1 {
			t.Errorf("Expected a slice of Response 1 object:, got %q", l)
		} else {
			mb := result[0]
			if !strings.HasPrefix(mb.Filename, "fd[") {
				t.Errorf("Expected name starting with fd[, got %q", mb.Filename)
			}
			if mb.Signature != "Eicar-Signature" {
				t.Errorf("Expected %q, got %q", "Eicar-Signature", mb.Signature)
			}
		}
	}
	if b, e = c.Reload(ctx); e != nil {
		t.Fatalf("An error should not be returned")
	}
	if !b {
		t.Errorf("Expected true, got %t", b)
	}

	t.Run("read with early EOF", func(t *testing.T) {
		// Most readers only return io.EOF, when nothing has been read. Some readers
		// however return io.EOF already when data is returned. This case needs to work
		// as well.

		if f, e = os.Open(tfn); e != nil {
			t.Fatalf("Expected nil got %q", e)
		}
		defer f.Close()

		stat, err := f.Stat()
		if err != nil {
			t.Fatalf("Expected nil got %q", e)
		}

		if result, e = c.ScanReader(ctx, &eofReader{f, stat.Size()}); e != nil {
			t.Errorf("Expected nil got %q", e)
		}
		l = len(result)
		if l == 0 {
			t.Errorf("Expected a slice of Response objects:, got %v", result)
		} else if l > 1 {
			t.Errorf("Expected a slice of Response 1 object:, got %d", l)
		} else {
			mb := result[0]
			if mb.Filename != "stream" {
				t.Errorf("Expected %q, got %q", "stream", mb.Filename)
			}
			if mb.Signature != "Eicar-Signature" {
				t.Errorf("Expected %q, got %q", "Eicar-Signature", mb.Signature)
			}
		}
	})
}

func copyFile(src, dst string, mode os.FileMode) error {
	var err error
	var srcfd *os.File
	var dstfd *os.File

	if srcfd, err = os.Open(src); err != nil {
		return err
	}
	defer srcfd.Close()

	if dstfd, err = os.Create(dst); err != nil {
		return err
	}
	defer dstfd.Close()

	if _, err = io.Copy(dstfd, srcfd); err != nil {
		return err
	}
	return os.Chmod(dst, mode)
}

type eofReader struct {
	source io.Reader
	size   int64
}

func (e *eofReader) Read(p []byte) (n int, err error) {
	if n, err = e.source.Read(p); err != nil {
		return
	}

	e.size -= int64(n)
	if e.size <= 0 {
		err = io.EOF
	}
	return
}
