package main

import (
	"flag"
	"connect"
	"handle"
	"fmt"
	"sync"
	"os"
	"time"
	"strings"
)

const (
	APP     = "Pulsar"
	VERSION = "1.0.0"
)

type Options struct {
	in       string
	out      string
	handlers string
	decode   bool
	duplex   bool
	verbose  bool
	buflen   int
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
	fmt.Printf("%s, Version: %s\n", APP, VERSION)
	flag.PrintDefaults()

	fmt.Printf("\nBullt-In Handlers:\n")
	for key, handler := range handle.Handlers {
		fmt.Printf("%s\t%s\n", key, handler.Description())
		if opts := handler.Options(); opts != nil {
			fmt.Println("- Handler options:")
			for _, opt := range opts {
				fmt.Printf("  %s=<%s>\n\t%s\n", opt.Name, strings.Join(opt.Values, "|"), opt.Description)
			}
			fmt.Println()
		}
	}

	fmt.Printf("\nBuit-In Connectors:\n")
	for key, connector := range connect.Connectors {
		fmt.Printf("%s\t%s\n", key, connector.Description())
	}
}

func main() {
	var options = Options{}
	var in connect.Connector = nil
	var out connect.Connector = nil
	var err error = nil
	var wait sync.WaitGroup

	flag.StringVar(&options.in, "in", "console", "Specify the input connector")
	flag.StringVar(&options.out, "out", "console", "Specify the output connector")
	flag.StringVar(&options.handlers, "handlers", "", "Specify data handlers separated by a comma")
	flag.BoolVar(&options.decode, "decode", false, "If enabled the data from IN connector will be decoded instead of encoded")
	flag.BoolVar(&options.duplex, "duplex", false, "Enable two-way data flow")
	flag.BoolVar(&options.verbose, "v", false, "Enable verbosity")
	flag.IntVar(&options.buflen, "buffer", 512, "Set the size of the reading buffer")
	flag.IntVar(&options.delay, "delay", 0, "Delay in millisecond to wait between I/O loop")

	// Handlers
	handle.RegisterHandler(handle.NewStub())
	handle.RegisterHandler(handle.NewBase64())

	// Connectors
	connect.RegisterConnector(connect.NewConsoleConnector())
	connect.RegisterConnector(connect.NewTcpConnector())

	flag.Usage = usage
	flag.Parse()

	if options.in != "" {
		connector, address := parseInOutString(options.in)
		if in, err = connect.MakeConnect(connector, true, address); err != nil {
			onError(err)
		}
	}

	if options.out != "" {
		connector, address := parseInOutString(options.out)
		if out, err = connect.MakeConnect(connector, false, address); err != nil {
			onError(err)
		}
	}

	if options.handlers != "" {
		tmp := strings.Split(options.handlers, ",")
		if err := handle.MakeChain(tmp, flag.Args()); err != nil {
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
		fmt.Printf("\nStats:\n")
		fmt.Printf("- IN connector(%s):\n", in.Name())
		fmt.Printf("  %s\n\n", in.Stats())
		fmt.Printf("- OUT connector(%s):\n", out.Name())
		fmt.Printf("  %s\n", out.Stats())
	}
}

func dataLoop(in *connect.Connector, out *connect.Connector, decode bool, options *Options, wait *sync.WaitGroup) {
	defer wait.Done()
	buffer := make([]byte, options.buflen)
	for {
		buffer, length, err := (*in).Read(buffer)
		if err != nil {
			if err.Error() == "EOF" {
				return
			}
			onError(err)
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
