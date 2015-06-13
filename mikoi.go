package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/jessevdk/go-flags"
)

const unknownReturnCode = 3

const portPlaceHolder = "{}"

var opts struct {
	Host string `short:"H" long:"hostname" description:"host name" required:"true"`
	Port string `short:"p" long:"port" description:"port number" required:"true"`

	Timeout time.Duration `short:"t" long:"timeout" description:"connection times out" default:"10s"`

	Verbose bool `short:"V" long:"verbose" description:"verbose" default:"false"`

	ProxyProto bool `short:"P" long:"proxyproto" description:"use ProxyProto" default:"true"`
}

// When there is no argument for command line, mikoi is in proxy server mode.
var serverMode bool

func main() {
	args, err := flags.Parse(&opts)
	if err != nil {
		os.Exit(unknownReturnCode)
	}

	if opts.Verbose {
		fmt.Fprintln(os.Stderr, "cmd args:", args)
	}

	if len(args) > 1 && !hasPortPlaceHolder(args) {
		fmt.Fprintln(os.Stderr, "command that mikoi will fork must have at least one port place holder")
		os.Exit(unknownReturnCode)
	}

	if len(args) == 0 {
		if opts.Verbose {
			fmt.Fprintln(os.Stderr, "mikoi is now running as proxy server mode")
		}
		serverMode = true
	}

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		fmt.Fprintln(os.Stderr, "unable to listen to localhost")
		os.Exit(unknownReturnCode)
	}

	laddr, ok := ln.Addr().(*net.TCPAddr)
	if !ok {
		fmt.Fprintln(os.Stderr, "unable to get a ephemeral port")
		os.Exit(unknownReturnCode)
	}

	if opts.Verbose {
		fmt.Fprintln(os.Stderr, "mikoi is now listening to", laddr.Port)
	}

	serverErrCh := make(chan error)
	// launching proxy server
	go server(ln, serverErrCh)

	cmdErrCh := make(chan error)
	if !serverMode {
		cmdArgs := replacePortPlaceHolder(laddr.Port, args)
		if opts.Verbose {
			fmt.Fprintln(os.Stderr, "cmd:", cmdArgs)
		}

		go cmdExecuter(cmdArgs, cmdErrCh)

		if opts.Verbose {
			fmt.Fprintln(os.Stderr, "waiting for command to be finished")
		}
	}

	for {
		select {
		case err := <-serverErrCh:
			fmt.Fprintln(os.Stderr, err)
			if serverMode {
				return
			}
		case err := <-cmdErrCh:
			if opts.Verbose {
				fmt.Fprintf(os.Stderr, "command was finished: %#v\n", err)
			}
			if err == nil {
				return
			}
			if eerr, ok := err.(*exec.ExitError); ok {
				if ws, wsok := eerr.Sys().(syscall.WaitStatus); wsok {
					es := ws.ExitStatus()
					os.Exit(es)
				}
			}
			os.Exit(unknownReturnCode)
		}
	}
}

func serve(conn net.Conn, errCh chan<- error) {
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(opts.Timeout))

	// opening a connection to server
	pconn, err := net.DialTimeout("tcp", opts.Host+":"+opts.Port, opts.Timeout)
	if err != nil {
		errCh <- err
		return
	}
	defer pconn.Close()
	pconn.SetDeadline(time.Now().Add(opts.Timeout))

	if opts.Verbose {
		fmt.Fprintln(os.Stderr, "mikoi connects a connection to server")
	}

	// Upgrade pconn to use ProxyProtocol
	if opts.ProxyProto {
		pconn = &ProxyConn{Conn: pconn}
	}

	// proxying a connection from plugin to server
	doneCh := make(chan struct{})
	go func() {
		io.Copy(pconn, conn)
		doneCh <- struct{}{}
	}()
	go func() {
		io.Copy(conn, pconn)
		doneCh <- struct{}{}
	}()

	if opts.Verbose {
		fmt.Fprintln(os.Stderr, "mikoi is now proxying traffic")
	}

	<-doneCh

	if opts.Verbose {
		fmt.Fprintln(os.Stderr, "mikoi is closing connections")
	}
}

func server(ln net.Listener, errCh chan<- error) {
	if opts.Verbose {
		fmt.Fprintln(os.Stderr, "mikoi is launching server")
	}

	defer ln.Close()
	for {
		conn, err := ln.Accept()
		if err != nil {
			errCh <- err
			continue
		}

		if opts.Verbose {
			fmt.Fprintln(os.Stderr, "mikoi accepts a connection from plugin")
		}

		go serve(conn, errCh)
	}
}

func cmdExecuter(cmdArgs []string, errCh chan<- error) {
	if opts.Verbose {
		fmt.Fprintln(os.Stderr, "mikoi is executing command")
	}
	cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	errCh <- cmd.Run()
}

func hasPortPlaceHolder(args []string) bool {
	for _, a := range args {
		if strings.Contains(a, portPlaceHolder) {
			return true
		}
	}
	return false
}

func replacePortPlaceHolder(port int, args []string) []string {
	p := strconv.Itoa(port)
	ret := make([]string, 0, len(args))
	for _, a := range args {
		ret = append(ret, strings.Replace(a, portPlaceHolder, p, -1))
	}
	return ret
}
