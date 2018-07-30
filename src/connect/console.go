package connect

import "os"

type console struct {
	*ConnectorStats
}

func NewConsoleConnector() Connector {
	return &console{&ConnectorStats{}}
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

func (c *console) Connect(listen bool, address string) (Connector, error) {
	return NewConsoleConnector(), nil
}

func (c *console) Read(buf []byte) ([]byte, int, error) {
	length, err := os.Stdin.Read(buf)
	c.recv += length
	return buf, length, err
}

func (c *console) Write(buf []byte, length int) (int, error) {
	wl, err := os.Stdout.Write(buf[:length])
	c.send += wl
	return wl, err
}
