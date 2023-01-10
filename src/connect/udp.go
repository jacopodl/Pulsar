package connect

import (
	"net"
	"os"
	"pulsar/src/packet"
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
	var laddr *net.UDPAddr = nil
	var raddr *net.UDPAddr = nil

	ip, port, err := ParseAddress(address)
	if err != nil {
		return nil, err
	}

	laddr = &net.UDPAddr{IP: ip, Port: port, Zone: ""}
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
	u.Write(nil, 0)
	u.conn.Close()
	u.Queue.Clear()
}

func (u *udp) Read() ([]byte, int, error) {
	var data []byte = nil
	var pkt *packet.Packet = nil
	var addr *net.UDPAddr = nil
	var err error = nil
	var length = 0
	var ok = false

	for {
		if length, addr, err = u.conn.ReadFromUDP(u.rbuf); err != nil {
			return nil, 0, err
		}
		if !u.checkPartner(addr) {
			continue
		}
		if u.plain {
			u.recv += length
			return u.rbuf, length, nil
		}
		if pkt, err = u.Deserialize(u.rbuf, length); err != nil {
			return nil, 0, err
		}
		u.Add(pkt)
		if data, ok = u.Buffer(); ok {
			u.recv += len(data)
			break
		}
	}
	return data, len(data), nil
}

func (u *udp) checkPartner(addr *net.UDPAddr) bool {
	if u.connHost == nil {
		u.connHost = addr
		return true
	}
	return u.connHost.IP.Equal(addr.IP) && u.connHost.Port == addr.Port
}

func (u *udp) Write(buf []byte, length int) (int, error) {
	var err error = nil

	if u.connHost == nil {
		return 0, nil
	}

	if u.plain {
		length, err = u.writePlain(buf, length)
	} else {
		length, err = u.write(buf, length)
	}

	if err != nil {
		return 0, err
	}

	u.send += length
	return length, nil
}

func (u *udp) write(buf []byte, length int) (int, error) {
	var pkts []*packet.Packet
	var err error = nil

	if pkts, err = u.Buffer2pkts(buf, length); err == nil {
		for i := range pkts {
			if _, err = u.conn.WriteToUDP(pkts[i].Serialize(), u.connHost); err != nil {
				return 0, err
			}
		}
		return length, nil
	}
	return 0, err
}

func (u *udp) writePlain(buf []byte, length int) (int, error) {
	if _, err := u.conn.WriteToUDP(buf[:length], u.connHost); err != nil {
		return 0, err
	}
	return length, nil
}
