package connect

import (
	"net"
)

type tcp struct {
	*ConnectorStats
	conn net.Conn
}

func NewTcpConnector() Connector {
	return &tcp{&ConnectorStats{}, nil}
}

func (t *tcp) Name() string {
	return "tcp"
}

func (t *tcp) Description() string {
	return "Read/write from TCP"
}

func (t *tcp) Stats() *ConnectorStats {
	return t.ConnectorStats
}

func (t *tcp) Connect(listen bool, address string) (Connector, error) {
	var err error = nil
	var conn net.Conn = nil
	if listen {
		var ln net.Listener = nil
		if ln, err = net.Listen(t.Name(), address); err == nil {
			if conn, err = ln.Accept(); err == nil {
				return &tcp{&ConnectorStats{}, conn}, nil
			}
		}
		return nil, err
	}
	if conn, err = net.Dial(t.Name(), address); err == nil {
		return &tcp{&ConnectorStats{}, conn}, nil
	}
	return nil, err
}

func (t *tcp) Read(buf []byte) ([]byte, int, error) {
	length, err := t.conn.Read(buf)
	t.recv += length
	return buf, length, err
}

func (t *tcp) Write(buf []byte, length int) (int, error) {
	wl, err := t.conn.Write(buf[:length])
	t.send += wl
	return wl, err
}
