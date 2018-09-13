package connect

import (
	"os"
)

const CLICHUNK = 4096

type console struct {
	*ConnectorStats
	buf    []byte
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
	return &console{&ConnectorStats{}, make([]byte, CLICHUNK), false}, nil
}

func (c *console) Close() {
	c.closed = true
}

func (c *console) Read() ([]byte, int, error) {
	if c.closed {
		return nil, 0, nil
	}
	length, err := os.Stdin.Read(c.buf)
	c.recv += length
	return c.buf, length, err
}

func (c *console) Write(buf []byte, length int) (int, error) {
	wl, err := os.Stdout.Write(buf[:length])
	c.send += wl
	return wl, err
}
