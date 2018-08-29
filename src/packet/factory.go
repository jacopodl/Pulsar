package packet

import (
	"fmt"
	"math/rand"
	"time"
)

type Factory struct {
	chunk int
	bseq  uint32
}

func NewPacketFactory(chunk int, bseq uint32) (*Factory, error) {
	if chunk <= HDRSIZE {
		return nil, fmt.Errorf("invalid chunk size, minimum size: %d", HDRSIZE)
	}
	return &Factory{chunk - HDRSIZE, bseq}, nil
}

func (f *Factory) Serialize(packet *Packet) ([]byte, error) {
	if len(packet.Data) > f.chunk {
		return nil, fmt.Errorf("invalid packet length, length > chunk")
	}

	return packet.Serialize(), nil
}

func (f *Factory) Deserialize(buf []byte, buflen int) (*Packet, error) {
	if buflen < HDRSIZE {
		return nil, fmt.Errorf("malformed packet")
	}
	if len(buf[HDRSIZE:]) > f.chunk {
		return nil, fmt.Errorf("length of packet data > chunk")
	}
	return Deserialize(buf, buflen)
}

func (f *Factory) Buffer2pkts(buf []byte, buflen int) ([]*Packet, error) {
	var done uint32 = 0
	var remaining = uint32(buflen)
	pkts := make([]*Packet, 0)

	for {
		write := uint32(f.chunk)
		if remaining < uint32(f.chunk) {
			write = remaining
		}
		data := make([]byte, f.chunk)
		copy(data, buf[done:done+write])
		if write < uint32(f.chunk) {
			rand.Seed(time.Now().UnixNano())
			if _, err := rand.Read(data[write:]); err != nil {
				return nil, err
			}
		}
		done += write
		remaining -= write
		pkts = append(pkts, NewPacket(f.bseq, f.bseq+done, uint32(buflen), data))

		if remaining == 0 {
			break
		}
	}
	f.bseq += 1 // Updating bseq
	return pkts, nil
}
