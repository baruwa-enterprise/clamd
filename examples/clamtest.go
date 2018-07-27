package main

import (
	"errors"
	"fmt"
	"go/build"
	"log"
	"os"
	"path"
	"strings"
	"time"

	"github.com/baruwa-enterprise/clamd"
	flag "github.com/spf13/pflag"
)

var (
	cfg     *Config
	cmdName string
)

// Config holds the configuration
type Config struct {
	Address string
	Port    int
}

func parseAddr(a string, p int) (n string, h string) {
	if strings.HasPrefix(a, "/") {
		n = "unix"
		h = a
	} else {
		n = "tcp"
		if strings.Contains(a, ":") {
			h = fmt.Sprintf("[%s]:%d", a, p)
		} else {
			h = fmt.Sprintf("%s:%d", a, p)
		}
	}
	return
}

func init() {
	cfg = &Config{}
	cmdName = path.Base(os.Args[0])
	flag.StringVarP(&cfg.Address, "host", "H", "192.168.1.14",
		`Specify Clamd host to connect to.`)
	flag.IntVarP(&cfg.Port, "port", "p", 3310,
		`In TCP/IP mode, connect to spamd server listening on given port`)
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [options]\n", cmdName)
	fmt.Fprint(os.Stderr, "\nOptions:\n")
	flag.PrintDefaults()
}

func main() {
	flag.Usage = usage
	flag.ErrHelp = errors.New("")
	flag.CommandLine.SortFlags = false
	flag.Parse()
	network, address := parseAddr(cfg.Address, cfg.Port)
	ch := make(chan bool)
	go func() {
		defer func() {
			ch <- true
		}()

		c, e := clamd.NewClient(network, address)
		if e != nil {
			log.Println(e)
			return
		}
		c.SetConnTimeout(5 * time.Second)
		r, e := c.Ping()
		if e != nil {
			log.Println(e)
			return
		}
		fmt.Println("PING", r)
	}()
	go func() {
		defer func() {
			ch <- true
		}()

		c, e := clamd.NewClient(network, address)
		if e != nil {
			log.Println(e)
			return
		}
		c.SetConnTimeout(5 * time.Second)
		s, e := c.Stats()
		if e != nil {
			log.Println(e)
			return
		}
		fmt.Println("STATS", s)
	}()
	go func() {
		defer func() {
			ch <- true
		}()

		c, e := clamd.NewClient(network, address)
		if e != nil {
			log.Println(e)
			return
		}
		c.SetConnTimeout(5 * time.Second)
		s, e := c.Version()
		if e != nil {
			log.Println(e)
			return
		}
		fmt.Println("VERSION", s)
	}()
	go func() {
		defer func() {
			ch <- true
		}()

		c, e := clamd.NewClient(network, address)
		if e != nil {
			log.Println(e)
			return
		}
		c.SetConnTimeout(5 * time.Second)
		s, e := c.VersionCmds()
		if e != nil {
			log.Println(e)
			return
		}
		fmt.Println("VERSIONCOMMANDS", s)
	}()
	<-ch
	if network == "unix" {
		c, e := clamd.NewClient(network, address)
		if e != nil {
			log.Println(e)
			return
		}
		c.SetConnTimeout(5 * time.Second)
		gopath := os.Getenv("GOPATH")
		if gopath == "" {
			gopath = build.Default.GOPATH
		}
		fn := path.Join(gopath, "src/github.com/baruwa-enterprise/clamd/examples/eicar.txt")
		s, e := c.InStream(fn)
		if e != nil {
			log.Println("ERROR:", e)
			return
		}
		fmt.Println("INSTREAM", "fn=>", s[0].Filename, "sig=>", s[0].Signature, "status=>", s[0].Status)
		fmt.Println("RAW=>", s[0].Raw)
	}
	// fildes
	if network == "unix" {
		c, e := clamd.NewClient(network, address)
		if e != nil {
			log.Println(e)
			return
		}
		c.SetConnTimeout(5 * time.Second)
		gopath := os.Getenv("GOPATH")
		if gopath == "" {
			gopath = build.Default.GOPATH
		}
		fn := path.Join(gopath, "src/github.com/baruwa-enterprise/clamd/examples/eicar.txt")
		s, e := c.Fildes(fn)
		if e != nil {
			log.Println("ERROR:", e)
			return
		}
		fmt.Println("FILDES", "fn=>", s[0].Filename, "sig=>", s[0].Signature, "status=>", s[0].Status)
		fmt.Println("RAW=>", s[0].Raw)
	}
	// Contscan
	if network != "unix" {
		c, e := clamd.NewClient(network, address)
		if e != nil {
			log.Println(e)
			return
		}
		c.SetConnTimeout(5 * time.Second)
		s, e := c.ContScan("/var/spool/testfiles/")
		if e != nil {
			log.Println("ERROR:", e)
			return
		}
		for _, rt := range s {
			fmt.Println("CONTSCAN", "fn=>", rt.Filename, "sig=>", rt.Signature, "status=>", rt.Status)
			fmt.Println("RAW=>", rt.Raw)
		}
		c, e = clamd.NewClient(network, address)
		if e != nil {
			log.Println(e)
			return
		}
		c.SetConnTimeout(5 * time.Second)
		r, e := c.Reload()
		if e != nil {
			log.Fatal(e)
		}
		fmt.Println("RELOAD", r)
	}
	/*
		e = c.Shutdown()
		if e != nil {
			log.Fatal(e)
		}
		fmt.Println("SHUTDOWN", r)
	*/
}
