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
	APP     = "Pulsar"
	VERSION = "1.0.0"
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
	fmt.Fprint(os.Stderr, err.Error())
	os.Exit(-1)
}

func parseInOutString(value string) (connector string, address string) {
	tmp := strings.SplitN(value, ":", 2)
	connector = tmp[0]
	if len(tmp) > 1 {
		address = tmp[1]
	}
	return
}

func usage() {
	fmt.Fprintf(os.Stderr, "%s, Version: %s\n", APP, VERSION)
	flag.PrintDefaults()

	fmt.Fprintf(os.Stderr, "\nBullt-In Handlers:\n")
	for key, handler := range handle.Handlers {
		fmt.Fprintf(os.Stderr, "%s\n\t%s\n", key, handler.Description())
		if opts := handler.Options(); opts != nil {
			fmt.Fprintf(os.Stderr, "* OPTIONS:\n")
			for _, opt := range opts {
				fmt.Fprintf(os.Stderr, "  %s\n\t%s\n", opt.String(), opt.Description)
			}
			fmt.Println()
		}
	}

	fmt.Fprintf(os.Stderr, "\nBuit-In Connectors:\n")
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
	handle.RegisterHandler(base.NewBase32())
	handle.RegisterHandler(base.NewBase64())
	handle.RegisterHandler(handle.NewCipher())

	// Connectors
	connect.RegisterConnector(connect.NewConsoleConnector())
	connect.RegisterConnector(connect.NewTcpConnector())
	connect.RegisterConnector(connect.NewUdpConnector())
	connect.RegisterConnector(connect.NewIcmpConnector())
	connect.RegisterConnector(connect.NewDnsConnector())

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
		tmp := strings.Split(options.handlers, ",")
		if err := handle.MakeChain(tmp, flag.Args()); err != nil {
			onError(err)
		}
	}

	if options.in != "" {
		connector, address := parseInOutString(options.in)
		if in, err = connect.MakeConnect(connector, true, options.inplain, address); err != nil {
			onError(err)
		}
	}

	if options.out != "" {
		connector, address := parseInOutString(options.out)
		if out, err = connect.MakeConnect(connector, false, options.outplain, address); err != nil {
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
