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
	var pkt *packet.Packet = nil
	var peer net.Addr = nil
	var msg *gicmp.Message = nil
	var err error = nil
	var length = 0

	for {
		if length, peer, err = i.conn.ReadFrom(i.rbuf); err != nil {
			return nil, 0, err
		}
		if msg, err = gicmp.ParseMessage(1, i.rbuf[:length]); err != nil {
			return nil, 0, err
		}

		if msg.Type == ipv4.ICMPTypeEcho {
			echo := msg.Body.(*gicmp.Echo)
			if !i.checkPartner(peer, echo.ID) {
				continue
			}
			if pkt, err = i.Deserialize(echo.Data, ICMPCHUNK); err != nil {
				return nil, 0, err
			}
			i.Add(pkt)
			if data = i.Buffer(); data != nil {
				i.recv += len(data)
				break
			}
		}
	}
	return data, len(data), nil
}

func (i *icmp) checkPartner(peer net.Addr, rid int) bool {
	if i.raddr == "" && (i.rid == -1 && rid != os.Getpid()&0xFFFF) {
		i.raddr = peer.String()
		i.rid = rid
		return true
	}
	return i.raddr == peer.String() && i.rid == rid
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
				ID: os.Getpid() & 0xFFFF, Seq: i.gblseq,
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
