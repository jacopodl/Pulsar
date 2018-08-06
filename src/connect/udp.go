package connect

import (
	"packet"
	"net"
	"strconv"
)

const UDPCHUNK = 384

type udp struct {
	*ConnectorStats
	*packet.Queue
	conn     *net.UDPConn
	connHost *net.UDPAddr
	rbuf     []byte
}

func NewUdpConnector() Connector {
	return &udp{&ConnectorStats{}, nil, nil, nil, nil}
}

func (u *udp) Name() string {
	return "udp"
}

func (u *udp) Description() string {
	return "Read/write from udp"
}

func (u *udp) Stats() *ConnectorStats {
	return u.ConnectorStats
}

func (u *udp) Connect(listen bool, address string) (Connector, error) {
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

	return &udp{&ConnectorStats{},
		packet.NewQueue(),
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
			println("setuphost")
		}
		if !u.connHost.IP.Equal(addr.IP) || u.connHost.Port != addr.Port {
			continue
			println("jmp")
			println(u.connHost)
			println(addr)
		}
		pkt, err := packet.DeserializePacket(u.rbuf, length)
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
	var total = 0

	if u.connHost == nil {
		println("nil")
		return 0, nil
	}

	pkts, err := packet.MakePackets(buf, length, UDPCHUNK, 0)
	if err != nil {
		return 0, err
	}

	for i := range pkts {
		wl, err := u.conn.WriteToUDP(packet.SerializePacket(pkts[i]), u.connHost)
		if err != nil {
			return 0, err
		}
		u.send += wl
		total += wl
	}

	return total, nil
}
