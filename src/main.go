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

	fmt.Printf("Bullt-In Handlers:\n")
	for key, handler := range handle.Handlers {
		fmt.Printf("\t%s\t%s\n", key, handler.Description())
	}

	fmt.Printf("\nBuit-In Connectors:\n")
	for key, connector := range connect.Connectors {
		fmt.Printf("\t%s\t%s\n", key, connector.Description())
	}
}

func main() {
	var options = Options{}
	var wait sync.WaitGroup
	var in connect.Connector = nil
	var out connect.Connector = nil

	flag.StringVar(&options.in, "in", "", "")
	flag.StringVar(&options.out, "out", "", "")
	flag.StringVar(&options.handlers, "handlers", "", "")
	flag.BoolVar(&options.duplex, "decode", false, "")
	flag.BoolVar(&options.duplex, "duplex", false, "")
	flag.IntVar(&options.buflen, "buffer", 512, "")
	flag.IntVar(&options.delay, "delay", 0, "")

	// Handlers
	handle.RegisterHandler(handle.NewStub())
	handle.RegisterHandler(handle.NewBase64())

	// Connectors
	connect.RegisterConnector(connect.NewConsoleConnector())

	flag.Usage = usage
	flag.Parse()

	if options.in != "" {
		connector, address := parseInOutString(options.in)
		cnt, ok := connect.Connectors[connector]
		if !ok {
			onError(fmt.Errorf("unknown IN connector %s, aborted", connector))
		}
		in = cnt.Connect(true, address)
	} else {
		in = connect.Connectors["console"].Connect(true, "")
	}

	if options.out != "" {
		connector, address := parseInOutString(options.out)
		cnt, ok := connect.Connectors[connector]
		if !ok {
			onError(fmt.Errorf("unknown OUT connector %s, aborted", connector))
		}
		out = cnt.Connect(false, address)
	} else {
		out = connect.Connectors["console"].Connect(false, "")
	}

	if options.handlers != "" {
		tmp := strings.Split(options.handlers, ",")
		if err := handle.MakeChain(tmp); err != nil {
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

	fmt.Printf("\nStats:\n")
	fmt.Printf("- IN connector(%s):\n", in.Name())
	fmt.Printf("  %s\n", in.Stats())
	fmt.Printf("- OUT connector(%s):\n", out.Name())
	fmt.Printf("  %s\n", out.Stats())
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
			time.Sleep(time.Duration(options.delay))
		}
	}
}
