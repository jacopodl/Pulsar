package connect

import (
	gicmp "golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"net"
	"os"
	"packet"
)

const ICMPCHUNK = 128
const ICMPPKT = 1500 - 20 - 8

type icmp struct {
	*ConnectorStats
	*packet.Factory
	*packet.Queue
	conn   *gicmp.PacketConn
	gblseq int
	raddr  string
	rid    int
	rbuf   []byte
}

func NewIcmpConnector() Connector {
	return &icmp{
		&ConnectorStats{},
		nil,
		nil,
		nil,
		1,
		"",
		-1,
		nil}
}

func (i *icmp) Name() string {
	return "icmp"
}

func (i *icmp) Description() string {
	return "Read/write from ICMP packets"
}

func (i *icmp) Stats() *ConnectorStats {
	return i.ConnectorStats
}

func (i *icmp) Connect(listen, plain bool, address string) (Connector, error) {
	var laddr = address
	var raddr = ""

	if !listen {
		raddr = laddr
		laddr = "0.0.0.0"
	}

	conn, err := gicmp.ListenPacket("ip4:icmp", laddr)
	if err != nil {
		return nil, err
	}

	pktFactory, _ := packet.NewPacketFactory(ICMPCHUNK, uint32(os.Getpid()&0xFFFF))

	return &icmp{&ConnectorStats{},
		pktFactory,
		packet.NewQueue(),
		conn,
		1,
		raddr,
		-1,
		make([]byte, ICMPPKT)}, nil
}

func (i *icmp) Close() {
	i.conn.Close()
	i.Queue.Clear()
}

func (i *icmp) Read() ([]byte, int, error) {
	var data []byte = nil
	var ok = false

	for {
		length, peer, err := i.conn.ReadFrom(i.rbuf)
		if length == 0 {
			return nil, 0, err
		}
		if i.raddr == "" {
			i.raddr = peer.String()
		}
		if i.raddr != peer.String() {
			continue
		}
		rm, err := gicmp.ParseMessage(1, i.rbuf[:length])
		if err != nil {
			return nil, 0, err
		}

		if rm.Type == ipv4.ICMPTypeEcho {
			echo := rm.Body.(*gicmp.Echo)
			if i.rid == -1 && echo.ID != os.Getpid()&0xFFFF {
				i.rid = echo.ID
			}
			if i.rid != echo.ID {
				continue
			}
			pkt, err := i.Deserialize(echo.Data, ICMPCHUNK)
			if err != nil {
				return nil, 0, err
			}
			i.Add(pkt)
			data, ok = i.Buffer()
			if ok {
				i.recv += len(data)
				break
			}
		}
	}
	return data, len(data), nil
}

func (i *icmp) Write(buf []byte, length int) (int, error) {
	if i.raddr == "" {
		return 0, nil
	}

	pkts, err := i.Buffer2pkts(buf, length)
	if err != nil {
		return 0, err
	}
	for j := range pkts {
		wm := gicmp.Message{
			Type: ipv4.ICMPTypeEcho, Code: 0,
			Body: &gicmp.Echo{
				ID:   os.Getpid() & 0xFFFF, Seq: i.gblseq,
				Data: pkts[j].Serialize()}}
		wb, err := wm.Marshal(nil)
		if err != nil {
			return 0, err
		}
		_, err = i.conn.WriteTo(wb, &net.IPAddr{IP: net.ParseIP(i.raddr)})
		if err != nil {
			return 0, err
		}
		i.gblseq++
	}
	return length, nil
}
