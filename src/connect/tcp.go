package connect

import (
	"net"
	"os"
	"packet"
)

const TCPCHUNK = 1320

type tcp struct {
	*ConnectorStats
	*packet.Factory
	*packet.Queue
	plain bool
	conn  net.Conn
	rbuf  []byte
}

func NewTcpConnector() Connector {
	return &tcp{&ConnectorStats{}, nil, nil, false, nil, nil}
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
		if ln, err = net.Listen(t.Name(), address); err != nil {
			return nil, err
		}
		if conn, err = ln.Accept(); err != nil {
			return nil, err
		}
	} else {
		if conn, err = net.Dial(t.Name(), address); err != nil {
			return nil, err
		}
	}

	pktFactory, _ := packet.NewPacketFactory(TCPCHUNK, uint32(os.Getpid()))

	return &tcp{
		&ConnectorStats{},
		pktFactory,
		packet.NewQueue(),
		plain,
		conn,
		make([]byte, TCPCHUNK)}, nil
}

func (t *tcp) Close() {
	t.conn.Close()
	t.Queue.Clear()
}

func (t *tcp) Read() ([]byte, int, error) {
	var data []byte = nil

	for {
		length, err := t.conn.Read(t.rbuf)
		if err != nil {
			return nil, 0, err
		}
		if t.plain {
			t.recv += length
			return t.rbuf, length, nil
		}
		pkt, err := t.Deserialize(t.rbuf, length)
		if err != nil {
			return nil, 0, err
		}
		t.Add(pkt)
		if data = t.Buffer(); data != nil {
			t.recv += len(data)
			break
		}
	}
	return data, len(data), nil
}

func (t *tcp) Write(buf []byte, length int) (int, error) {
	var err error = nil

	if t.plain {
		length, err = t.writePlain(buf, length)
	} else {
		length, err = t.write(buf, length)
	}

	if err != nil {
		return 0, err
	}

	t.send += length
	return length, nil
}

func (t *tcp) write(buf []byte, length int) (int, error) {
	var pkts []*packet.Packet
	var err error = nil

	if pkts, err = t.Buffer2pkts(buf, length); err == nil {
		for i := range pkts {
			if _, err := t.conn.Write(pkts[i].Serialize()); err != nil {
				return 0, err
			}
		}
		return length, nil
	}
	return 0, err
}

func (t *tcp) writePlain(buf []byte, length int) (int, error) {
	if _, err := t.conn.Write(buf[:length]); err != nil {
		return 0, err
	}
	return length, nil
}
