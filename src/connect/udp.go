package connect

import (
	"packet"
	"net"
	"strconv"
	"os"
)

const UDPCHUNK = 384

type udp struct {
	*ConnectorStats
	*packet.Factory
	*packet.Queue
	plain    bool
	conn     *net.UDPConn
	connHost *net.UDPAddr
	rbuf     []byte
}

func NewUdpConnector() Connector {
	return &udp{&ConnectorStats{}, nil, nil, false, nil, nil, nil}
}

func (u *udp) Name() string {
	return "udp"
}

func (u *udp) Description() string {
	return "Read/write from UDP packets"
}

func (u *udp) Stats() *ConnectorStats {
	return u.ConnectorStats
}

func (u *udp) Connect(listen, plain bool, address string) (Connector, error) {
	var strIp = ""
	var strPort = ""
	var laddr *net.UDPAddr = nil
	var raddr *net.UDPAddr = nil
	var err error = nil

	strIp, strPort, err = net.SplitHostPort(address)
	if err != nil {
		return nil, err
	}

	port, err := strconv.ParseUint(strPort, 10, 16)
	if err != nil {
		return nil, err
	}

	laddr = &net.UDPAddr{IP: net.ParseIP(strIp), Port: int(port), Zone: ""}
	if !listen {
		raddr = laddr
		laddr = &net.UDPAddr{IP: net.IPv4zero, Port: 0, Zone: ""}
	}

	conn, err := net.ListenUDP("udp", laddr)
	if err != nil {
		return nil, err
	}

	pktFactory, _ := packet.NewPacketFactory(UDPCHUNK, uint32(os.Getpid()))

	return &udp{&ConnectorStats{},
		pktFactory,
		packet.NewQueue(),
		plain,
		conn,
		raddr,
		make([]byte, UDPCHUNK)}, nil
}

func (u *udp) Close() {
	u.conn.Close()
	u.Queue.Clear()
}

func (u *udp) Read() ([]byte, int, error) {
	var data []byte = nil
	var ok = false

	for {
		length, addr, err := u.conn.ReadFromUDP(u.rbuf)
		if length == 0 {
			return nil, 0, err
		}
		if u.connHost == nil {
			u.connHost = addr
		}
		if !u.connHost.IP.Equal(addr.IP) || u.connHost.Port != addr.Port {
			continue
		}
		if u.plain {
			u.recv = length
			return u.rbuf, length, nil
		}
		pkt, err := u.Deserialize(u.rbuf, length)
		if err != nil {
			return nil, 0, err
		}
		u.Add(pkt)
		data, ok = u.Buffer()
		if ok {
			u.recv += len(data)
			break
		}
	}
	return data, len(data), nil
}

func (u *udp) Write(buf []byte, length int) (int, error) {
	if u.connHost == nil {
		return 0, nil
	}

	if !u.plain {
		pkts, err := u.Buffer2pkts(buf, length)
		if err != nil {
			return 0, err
		}
		for i := range pkts {
			_, err := u.conn.WriteToUDP(pkts[i].Serialize(), u.connHost)
			if err != nil {
				return 0, err
			}
		}
	} else {
		_, err := u.conn.WriteToUDP(buf[:length], u.connHost)
		if err != nil {
			return 0, err
		}
	}
	u.send += length
	return length, nil
}
