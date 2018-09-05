package connect

import (
	dproto "connect/proto/dns"
	"encoding/base32"
	"fmt"
	"net"
	"os"
	"packet"
	"strings"
)

const BASECHUNK = 340

type dns struct {
	*ConnectorStats
	*packet.Factory
	*packet.Queue
	domain   string
	id       uint16
	conn     *net.UDPConn
	connHost *net.UDPAddr
	rbuf     []byte
}

func NewDnsConnector() Connector {
	return &dns{&ConnectorStats{}, nil, nil, "", 0, nil, nil, nil}
}

func splitDomainAddr(address string) (string, string, error) {
	domaddr := strings.Split(address, "@")
	if len(domaddr) < 2 {
		return "", "", fmt.Errorf("missing domain name")
	}
	if !strings.HasPrefix(domaddr[0], ".") {
		domaddr[0] = "." + domaddr[0]
	}
	return domaddr[0], domaddr[1], nil
}

func (d *dns) Name() string {
	return "dns"
}

func (d *dns) Description() string {
	return "Read/write from DNS packets"
}

func (d *dns) Stats() *ConnectorStats {
	return d.ConnectorStats
}

func (d *dns) Connect(listen, plain bool, address string) (Connector, error) {
	var laddr *net.UDPAddr = nil
	var raddr *net.UDPAddr = nil
	var err error = nil

	domain, addr, err := splitDomainAddr(address)
	if err != nil {
		return nil, err
	}

	ip, port, err := ParseAddress(addr)
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

	pktFactory, _ := packet.NewPacketFactory(computeChunkSize(domain), uint32(os.Getpid()))

	return &dns{
		&ConnectorStats{},
		pktFactory,
		packet.NewQueue(),
		domain,
		uint16(os.Getpid()),
		conn,
		raddr,
		make([]byte, BASECHUNK)}, nil
}

func (d *dns) Close() {
	d.Write(nil, 0)
	d.conn.Close()
	d.Queue.Clear()
}

func (d *dns) Read() ([]byte, int, error) {
	var data []byte = nil
	var pkt *packet.Packet = nil
	var addr *net.UDPAddr = nil
	var err error = nil
	var length = 0
	var ok = false

	for {
		if length, addr, err = d.conn.ReadFromUDP(d.rbuf); err != nil {
			return nil, 0, err
		}

		if data, length, err = extractData(d.rbuf[:length], d.domain); err != nil {
			return nil, 0, err
		}

		if length == 0 || !d.checkPartner(addr) {
			continue
		}

		if pkt, err = d.Deserialize(data, length); err != nil {
			return nil, 0, err
		}
		d.Add(pkt)
		if data, ok = d.Buffer(); ok {
			d.recv += len(data)
			break
		}
	}
	return data, len(data), nil
}

func (d *dns) checkPartner(addr *net.UDPAddr) bool {
	if d.connHost == nil {
		d.connHost = addr
		return true
	}
	return d.connHost.IP.Equal(addr.IP) && d.connHost.Port == addr.Port
}

func (d *dns) Write(buf []byte, length int) (int, error) {
	if d.connHost == nil {
		return 0, nil
	}

	pkts, err := d.Buffer2pkts(buf, length)
	if err != nil {
		return 0, err
	}

	for i := range pkts {
		dpkt := dproto.NewDnsPacket(d.id)
		data := base32.StdEncoding.EncodeToString(pkts[i].Serialize())
		ldata := len(data)
		wlen := dproto.MAXLBLSIZE
		wb := 0
		for ; ldata != 0; ldata -= wlen {
			if wlen >= ldata {
				wlen = ldata
			}
			dpkt.AddQuestion(data[wb:wb+wlen]+d.domain, dproto.TYPE_A, dproto.CLASS_IN)
			wb += wlen
		}
		_, err := d.conn.WriteToUDP(dpkt.Serialize(), d.connHost)
		if err != nil {
			return 0, err
		}
		d.id += 1
	}

	d.send += length
	return length, nil
}

func computeChunkSize(domain string) (chunk int) {
	chunk = BASECHUNK - len(domain) - dproto.DNSHDRSIZE
	chunk = (chunk / 8) * 5
	chunk -= (dproto.QUERYSIZE + 1) * (((chunk / 5) * 8) / dproto.MAXLBLSIZE)
	return
}

func extractData(buf []byte, domain string) ([]byte, int, error) {
	var data []byte
	var pkt *dproto.Dns
	var err error
	var b32data = ""

	if pkt, err = dproto.Deserialize(buf); err == nil {
		questions := pkt.GetQuestions()
		for i := range questions {
			if len(questions[i]) < 2 {
				continue
			}
			if dom := strings.Join(questions[i][1:], "."); dom == domain[1:] {
				b32data += questions[i][0]
			}
		}
		if data, err = base32.StdEncoding.DecodeString(b32data); err == nil {
			return data, len(data), nil
		}
	}
	return nil, 0, err
}
