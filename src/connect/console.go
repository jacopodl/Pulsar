package connect

import (
	"os"
	"time"
)

const CLICHUNK = 4096

type clidata struct {
	length int
	buf    []byte
	err    error
}

type console struct {
	*ConnectorStats
	data   chan clidata
	closed bool
}

func NewConsoleConnector() Connector {
	return &console{&ConnectorStats{}, nil, true}
}

func (c *console) Name() string {
	return "console"
}

func (c *console) Description() string {
	return "Read/write from stdin/stdout"
}

func (c *console) Stats() *ConnectorStats {
	return c.ConnectorStats
}

func (c *console) Connect(listen, plain bool, address string) (Connector, error) {
	cli := console{&ConnectorStats{}, make(chan clidata), false}

	go func() {
		for {
			buf := make([]byte, CLICHUNK)
			length, err := os.Stdin.Read(buf)
			cli.data <- clidata{length, buf, err}
			if err != nil || cli.closed {
				return
			}
		}
	}()

	return &cli, nil
}

func (c *console) Close() {
	c.closed = true
}

func (c *console) Read() ([]byte, int, error) {
	for {
		select {
		case data := <-c.data:
			c.recv += data.length
			return data.buf, data.length, data.err
		case <-time.After(10 * time.Millisecond):
			if c.closed {
				return nil, 0, nil
			}
		}
	}
}

func (c *console) Write(buf []byte, length int) (int, error) {
	wl, err := os.Stdout.Write(buf[:length])
	c.send += wl
	return wl, err
}
