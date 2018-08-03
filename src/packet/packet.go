package packet

import (
	"math/rand"
	"time"
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

func SerializePacket(packet *Packet) []byte {
	bseq := make([]byte, 4)
	seq := make([]byte, 4)
	tlen := make([]byte, 4)

	// To network byte order
	binary.BigEndian.PutUint32(bseq, packet.BSeq)
	binary.BigEndian.PutUint32(seq, packet.Seq)
	binary.BigEndian.PutUint32(tlen, packet.Tlen)

	buffer := append([]byte{}, bseq...)
	buffer = append(buffer, seq...)
	buffer = append(buffer, tlen...)
	buffer = append(buffer, packet.Data...)

	return buffer
}

func DeserializePacket(buf []byte, buflen int) (*Packet, error) {
	if buflen < 12 {
		return nil, fmt.Errorf("malformed packet")
	}

	// From network byte order
	bseq := binary.BigEndian.Uint32(buf[0:4])
	seq := binary.BigEndian.Uint32(buf[4:8])
	tlen := binary.BigEndian.Uint32(buf[8:12])

	packet := NewPacket(bseq, seq, tlen, append([]byte{}, buf[12:]...))

	return packet, nil
}

func MakePackets(buf []byte, buflen int, chunk, bseq uint32) ([]*Packet, error) {
	var done uint32 = 0
	var remaining = uint32(buflen)
	pkts := make([]*Packet, 0)

	chunk -= HDRSIZE // Header of packet!

	if chunk <= 0 {
		return nil, fmt.Errorf("invalid chunk size")
	}

	for ; remaining > 0; {
		write := uint32(chunk)
		if remaining < chunk {
			write = remaining
		}
		data := make([]byte, chunk)
		copy(data, buf[done:done+write])
		if write < chunk {
			rand.Seed(time.Now().UnixNano())
			if _, err := rand.Read(data[write:]); err != nil {
				return nil, err
			}
		}
		pkts = append(pkts, NewPacket(bseq, bseq+done+write, uint32(buflen), data))
		done += write
		remaining -= write
	}
	return pkts, nil
}

func (p *Packet) String() string {
	return fmt.Sprintf("Bseq: %d, Seq: %d, Total: %d, bytes: %x", p.BSeq, p.Seq, p.Tlen, p.Data)
}
