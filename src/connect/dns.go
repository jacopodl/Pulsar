package connect

import (
	"packet"
	"net"
	"strconv"
	"os"
	"encoding/base32"
	"strings"
	"fmt"
	dproto "connect/proto/dns"
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
	var strIp = ""
	var strPort = ""
	var laddr *net.UDPAddr = nil
	var raddr *net.UDPAddr = nil
	var err error = nil

	domaddr := strings.Split(address, "@")
	if len(domaddr) < 2 {
		return nil, fmt.Errorf("missing domain name")
	}
	if !strings.HasPrefix(domaddr[0], ".") {
		domaddr[0] = "." + domaddr[0]
	}

	strIp, strPort, err = net.SplitHostPort(domaddr[1])
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

	// Calculating chunk size
	chunk := BASECHUNK - len(domaddr[0]) - dproto.DNSHDRSIZE
	chunk = (chunk / 8) * 5
	chunk -= (dproto.QUERYSIZE + 1) * (((chunk / 5) * 8) / dproto.MAXLBLSIZE)
	// EOL

	pktFactory, _ := packet.NewPacketFactory(chunk, uint32(os.Getpid()))

	return &dns{&ConnectorStats{},
		pktFactory,
		packet.NewQueue(),
		domaddr[0],
		uint16(os.Getpid()),
		conn,
		raddr,
		make([]byte, BASECHUNK)}, nil
}

func (d *dns) Close() {
	d.conn.Close()
	d.Queue.Clear()
}

func (d *dns) Read() ([]byte, int, error) {
	var data []byte = nil
	var ok = false

	for {
		length, addr, err := d.conn.ReadFromUDP(d.rbuf)
		if length == 0 {
			return nil, 0, err
		}

		buf, length, err := d.extractData()
		if err != nil {
			return nil, 0, err
		}

		if d.connHost == nil && length > 0 {
			d.connHost = addr
		}

		if !d.connHost.IP.Equal(addr.IP) || d.connHost.Port != addr.Port {
			continue
		}

		pkt, err := d.Deserialize(buf, length)
		if err != nil {
			return nil, 0, err
		}
		d.Add(pkt)
		data, ok = d.Buffer()
		if ok {
			d.recv += len(data)
			break
		}
	}
	return data, len(data), nil
}

func (d *dns) extractData() ([]byte, int, error) {
	b32data := ""

	pkt := dproto.Deserialize(d.rbuf)
	questions := pkt.GetQuestions()
	for i := range questions {
		if len(questions[i]) < 2 {
			continue
		}
		if domain := strings.Join(questions[i][1:], "."); domain == d.domain[1:] {
			b32data += questions[i][0]
		}
	}
	data, err := base32.StdEncoding.DecodeString(b32data)
	if err != nil {
		return nil, 0, err
	}
	return data, len(data), nil
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
		wlen := 63
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
