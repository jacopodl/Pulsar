package connect

import "os"

const CLICHUNK = 4096

type console struct {
	*ConnectorStats
	rbuf []byte
}

func NewConsoleConnector() Connector {
	return &console{&ConnectorStats{}, nil}
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
	return &console{&ConnectorStats{}, make([]byte, CLICHUNK)}, nil
}

func (c *console) Close() {
}

func (c *console) Read() ([]byte, int, error) {
	length, err := os.Stdin.Read(c.rbuf)
	c.recv += length
	return c.rbuf, length, err
}

func (c *console) Write(buf []byte, length int) (int, error) {
	wl, err := os.Stdout.Write(buf[:length])
	c.send += wl
	return wl, err
}
