package main

import (
	"connect"
	"flag"
	"fmt"
	"handle"
	"handle/base"
	"io"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	// APP     = "Pulsar"
	VERSION = "1.0.0"
	LOGO    = `
.______    __    __   __          _______.     ___      .______
|   _  \  |  |  |  | |  |        /       |    /   \     |   _  \
|  |_)  | |  |  |  | |  |       |   (----'   /  ^  \    |  |_)  |
|   ___/  |  |  |  | |  |        \   \      /  /_\  \   |      /
|  |      |  '--'  | |  '----.----)   |    /  _____  \  |  |\  \----.
| _|       \______/  |_______|_______/    /__/     \__\ | _| \._____|`
)

type Options struct {
	in       string
	out      string
	handlers string
	plain    string
	inplain  bool
	outplain bool
	decode   bool
	duplex   bool
	verbose  bool
	closed   bool
	delay    int
}

func onError(err error) {
	fmt.Fprintf(os.Stderr, "%s\n", err.Error())
	os.Exit(-1)
}

func usage() {
	fmt.Fprintf(os.Stderr, LOGO)
	fmt.Fprintf(os.Stderr, " V: %s\n\n", VERSION)
	flag.PrintDefaults()

	fmt.Fprintf(os.Stderr, "\nBuilt-In Handlers:\n\n")
	for key, handler := range handle.Handlers {
		fmt.Fprintf(os.Stderr, "%s\n\t%s\n", key, handler.Description())
	}

	fmt.Fprintf(os.Stderr, "\nBuilt-In Connectors:\n\n")
	for key, connector := range connect.Connectors {
		fmt.Fprintf(os.Stderr, "%s\n\t%s\n", key, connector.Description())
	}
}

func main() {
	var options = Options{}
	var in connect.Connector = nil
	var out connect.Connector = nil
	var err error = nil
	var wait sync.WaitGroup

	flag.BoolVar(&options.decode, "decode", false, "If enabled the data from IN connector will be decoded instead of encoded")
	flag.BoolVar(&options.duplex, "duplex", false, "Enable two-way data flow")
	flag.StringVar(&options.handlers, "handlers", "", "Specify data handlers separated by a comma")
	flag.StringVar(&options.in, "in", "console", "Specify the input connector")
	flag.StringVar(&options.out, "out", "console", "Specify the output connector")
	flag.StringVar(&options.plain, "plain", "", "Specifies the directions which must use a plain connector")
	flag.BoolVar(&options.verbose, "v", false, "Enable verbosity")
	flag.IntVar(&options.delay, "delay", 0, "Delay in millisecond to wait between I/O loop")

	// Handlers
	handle.Register(handle.NewStub())
	handle.Register(base.NewBase32())
	handle.Register(base.NewBase64())
	handle.Register(handle.NewCipher())

	// Connectors
	connect.Register(connect.NewConsoleConnector())
	connect.Register(connect.NewTcpConnector())
	connect.Register(connect.NewUdpConnector())
	connect.Register(connect.NewIcmpConnector())
	connect.Register(connect.NewDnsConnector())

	flag.Usage = usage
	flag.Parse()

	if options.plain != "" {
		tmp := strings.Split(options.plain, ",")
		for i := range tmp {
			switch tmp[i] {
			case "in":
				options.inplain = true
			case "out":
				options.outplain = true
			default:
				onError(fmt.Errorf("%s is an invalid direction, you must use only 'in'/'out'", tmp[i]))
			}
		}
	}

	if options.handlers != "" {
		handler := strings.Split(options.handlers, ",")
		if err := handle.MakeChain(handler); err != nil {
			onError(err)
		}
	}

	if options.in != "" {
		if in, err = connect.MakeConnect(options.in, true, options.inplain); err != nil {
			onError(err)
		}
	}

	if options.out != "" {
		if out, err = connect.MakeConnect(options.out, false, options.outplain); err != nil {
			onError(err)
		}
	}

	wait.Add(1)
	go dataLoop(&in, &out, options.decode, &options, &wait)
	if options.duplex {
		wait.Add(1)
		go dataLoop(&out, &in, !options.decode, &options, &wait)
	}

	wait.Wait()

	if options.verbose {
		fmt.Fprintf(os.Stderr, "\nStats:\n")
		fmt.Fprintf(os.Stderr, "- IN connector(%s):\n", in.Name())
		fmt.Fprintf(os.Stderr, "  %s\n\n", in.Stats())
		fmt.Fprintf(os.Stderr, "- OUT connector(%s):\n", out.Name())
		fmt.Fprintf(os.Stderr, "  %s\n", out.Stats())
	}
}

func dataLoop(in, out *connect.Connector, decode bool, options *Options, wait *sync.WaitGroup) {
	defer wait.Done()
	for {
		buffer, length, err := (*in).Read()
		if length == 0 {
			if !options.closed && err != nil && err != io.EOF {
				onError(err)
			}
			options.closed = true
			(*out).Close()
			return
		}
		tbuf, length, err := handle.Process(buffer, length, decode)
		if err != nil {
			onError(err)
		}
		if _, err := (*out).Write(tbuf, length); err != nil {
			onError(err)
		}
		if options.delay > 0 {
			time.Sleep(time.Duration(options.delay) * time.Millisecond)
		}
	}
}
