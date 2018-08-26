package packet

import (
	"fmt"
	"encoding/binary"
)

const HDRSIZE = 12

type Packet struct {
	BSeq uint32
	Seq  uint32
	Tlen uint32
	Data []byte
}

func NewPacket(bseq, seq, tlen uint32, data []byte) *Packet {
	return &Packet{
		bseq,
		seq,
		tlen,
		data}
}

func Deserialize(buf []byte, buflen int) (*Packet, error) {
	if buflen < HDRSIZE {
		return nil, fmt.Errorf("malformed packet")
	}

	// From network byte order
	bseq := binary.BigEndian.Uint32(buf[:4])
	seq := binary.BigEndian.Uint32(buf[4:8])
	tlen := binary.BigEndian.Uint32(buf[8:])

	packet := NewPacket(bseq, seq, tlen, append([]byte{}, buf[HDRSIZE:]...))

	return packet, nil
}

func (p *Packet) Serialize() []byte {
	hdr := make([]byte, HDRSIZE)

	// To network byte order
	binary.BigEndian.PutUint32(hdr[:4], p.BSeq)
	binary.BigEndian.PutUint32(hdr[4:8], p.Seq)
	binary.BigEndian.PutUint32(hdr[8:], p.Tlen)

	hdr = append(hdr, p.Data...)

	return hdr
}

func (p *Packet) String() string {
	return fmt.Sprintf("Bseq: %d, Seq: %d, Total: %d, bytes: %x", p.BSeq, p.Seq, p.Tlen, p.Data)
}
