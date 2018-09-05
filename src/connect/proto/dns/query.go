package dns

import (
	"encoding/binary"
)

const QUERYSIZE = 4

type Query struct {
	data  []byte
	ttype uint16
	class uint16
}

func NewQuery(name string, ttype uint16, class uint16) *Query {
	return &Query{dname2qname(name), ttype, class}
}

func (q *Query) Serialize() []byte {
	dlen := len(q.data)

	buf := make([]byte, dlen+QUERYSIZE)
	copy(buf, q.data)
	binary.BigEndian.PutUint16(buf[dlen:dlen+2], q.ttype)
	binary.BigEndian.PutUint16(buf[dlen+2:], q.class)

	return buf
}

func (q *Query) Deserialize(buf []byte) *Query {
	tmp := Query{}

	tmp.data = append(tmp.data, buf[:len(buf)-QUERYSIZE]...)
	tmp.ttype = binary.BigEndian.Uint16(buf[:2])
	tmp.class = binary.BigEndian.Uint16(buf[2:])

	return &tmp
}

func dname2qname(dname string) []byte {
	var ins = 0
	var count byte = 0
	var i = 0
	var qname = make([]byte, len(dname)+2)

	for i = range dname {
		if dname[i] == '.' {
			qname[ins] = count
			ins = i + 1
			count = 0
			continue
		}
		qname[i+1] = dname[i]
		count++
	}

	qname[ins] = count
	qname[i+2] = 0x00
	return qname
}
