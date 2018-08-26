package dns

import "encoding/binary"

const QUERYSIZE = 4

type Query struct {
	Type  uint16
	Class uint16
}

func (q *Query) Serialize() []byte {
	buf := make([]byte, 4)

	binary.BigEndian.PutUint16(buf[:2], q.Type)
	binary.BigEndian.PutUint16(buf[2:], q.Class)

	return buf
}

func (q *Query) Deserialize(buf []byte) *Query {
	tmp := Query{}

	tmp.Type = binary.BigEndian.Uint16(buf[:2])
	tmp.Class = binary.BigEndian.Uint16(buf[2:])

	return &tmp
}
