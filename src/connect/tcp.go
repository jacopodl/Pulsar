package connect

import (
	"net"
	"packet"
)

const TCPCHUNK = 1320

type tcp struct {
	*ConnectorStats
	*packet.Queue
	plain bool
	conn  net.Conn
	rbuf  []byte
}

func NewTcpConnector() Connector {
	return &tcp{&ConnectorStats{}, nil, false, nil, nil}
}

func (t *tcp) Name() string {
	return "tcp"
}

func (t *tcp) Description() string {
	return "Read/write from TCP stream"
}

func (t *tcp) Stats() *ConnectorStats {
	return t.ConnectorStats
}

func (t *tcp) Connect(listen, plain bool, address string) (Connector, error) {
	var err error = nil
	var conn net.Conn = nil
	if listen {
		var ln net.Listener = nil
		if ln, err = net.Listen(t.Name(), address); err == nil {
			if conn, err = ln.Accept(); err == nil {
				return &tcp{&ConnectorStats{},
					packet.NewQueue(),
					plain,
					conn,
					make([]byte, TCPCHUNK)}, nil
			}
		}
		return nil, err
	}
	if conn, err = net.Dial(t.Name(), address); err == nil {
		return &tcp{&ConnectorStats{},
			packet.NewQueue(),
			plain,
			conn,
			make([]byte, TCPCHUNK)}, nil
	}
	return nil, err
}

func (t *tcp) Close() {
	t.conn.Close()
	t.Queue.Clear()
}

func (t *tcp) Read() ([]byte, int, error) {
	var data []byte = nil
	var ok = false

	for {
		length, err := t.conn.Read(t.rbuf)
		if length == 0 {
			return nil, 0, err
		}
		if t.plain {
			t.recv = length
			return t.rbuf, length, nil
		}
		pkt, err := packet.DeserializePacket(t.rbuf, length)
		if err != nil {
			return nil, 0, err
		}
		t.Add(pkt)
		data, ok = t.Buffer()
		if ok {
			t.recv += len(data)
			break
		}
	}

	return data, len(data), nil
}

func (t *tcp) Write(buf []byte, length int) (int, error) {
	if !t.plain {
		pkts, err := packet.MakePackets(buf, length, TCPCHUNK, 0)
		if err != nil {
			return 0, err
		}
		for i := range pkts {
			_, err := t.conn.Write(packet.SerializePacket(pkts[i]))
			if err != nil {
				return 0, err
			}
		}
	} else {
		_, err := t.conn.Write(buf[:length])
		if err != nil {
			return 0, err
		}
	}
	t.send += length
	return length, nil
}
