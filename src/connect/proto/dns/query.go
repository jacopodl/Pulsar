package dns

import (
	"encoding/binary"
	"strings"
)

const (
	QUERYSIZE  = 4
	MAXLBLSIZE = 63
	LABELSEP   = '.'
)

type Query struct {
	Domain string
	Type   uint16
	Class  uint16
}

func NewQuery(domain string, qtype uint16, class uint16) *Query {
	return &Query{domain, qtype, class}
}

func DeserializeQuery(buf []byte) *Query {
	tmp := Query{}
	sep := len(buf) - QUERYSIZE

	tmp.Domain = qname2dname(buf[:sep])
	tmp.Type = binary.BigEndian.Uint16(buf[sep : sep+2])
	tmp.Class = binary.BigEndian.Uint16(buf[sep+2:])

	return &tmp
}

func (q *Query) Serialize() []byte {
	qname := dname2qname(q.Domain)
	buf := make([]byte, len(qname)+QUERYSIZE)

	lq := copy(buf, qname)
	binary.BigEndian.PutUint16(buf[lq:lq+2], q.Type)
	binary.BigEndian.PutUint16(buf[lq+2:], q.Class)

	return buf
}

func (q *Query) CountLabel() int {
	count := 0
	for i := range q.Domain {
		if q.Domain[i] == LABELSEP {
			count++
		}
	}
	return count
}

func (q *Query) Labels() []string {
	return strings.Split(q.Domain, string(LABELSEP))
}

func (q *Query) SplitLast() (string, string) {
	split := strings.SplitN(q.Domain, string(LABELSEP), 2)
	if len(split) < 2 {
		return "", split[0]
	}
	return split[0], split[1]
}

func dname2qname(dname string) []byte {
	var ins = 0
	var count byte = 0
	var i = 0
	var qname = make([]byte, len(dname)+2)

	for i = range dname {
		if dname[i] == LABELSEP {
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

func qname2dname(qname []byte) string {
	domain := ""
	idx := 0

	for lq := int(qname[idx]); lq != 0x00; lq = int(qname[idx]) {
		domain += string(qname[idx+1 : idx+lq+1])
		if idx += lq + 1; qname[idx] != 0x00 {
			domain += string(LABELSEP)
		}
	}

	return domain
}
