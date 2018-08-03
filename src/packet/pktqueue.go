package packet

import (
	"sort"
)

type Queue struct {
	queue []*Packet
}

func NewQueue() *Queue {
	return &Queue{make([]*Packet, 0)}
}

func (q *Queue) Add(packet *Packet) {
	q.queue = append(q.queue, packet)

	if packet.Seq < q.queue[len(q.queue)-1].Seq {
		// Sort pkt by seq number
		sort.Slice(q.queue, func(i, j int) bool {
			return q.queue[i].Seq < q.queue[i].Seq
		})
	}
}

func (q *Queue) Buffer() ([]byte, bool) {
	var buf []byte = nil
	var bseq uint32 = 0
	var lastlen uint32 = 0
	var lastpkt = -1
	var ok = false

	for i := range q.queue {
		pkt := q.queue[i]

		if i == 0 {
			// Get current BSeq
			bseq = pkt.BSeq
		}

		if bseq != pkt.BSeq {
			// New sequence found before the end of the previous one
			break
		}

		if pkt.Seq-bseq == pkt.Tlen {
			// We have all data! :)
			lastpkt = i
			break
		}
		lastlen = pkt.Seq - bseq
	}

	if lastpkt > -1 {
		for i := 0; i < lastpkt; i++ {
			buf = append(buf, q.queue[i].Data...)
			q.queue[i] = nil
		}
		buf = append(buf, q.queue[lastpkt].Data[:q.queue[lastpkt].Tlen-lastlen]...)
		q.queue[lastpkt] = nil
		q.queue = q.queue[lastpkt+1:] // remove packets from queue
		ok = true
	}

	return buf, ok
}

func (q *Queue) Clear() {
	for i := range q.queue {
		q.queue[i] = nil
	}
	q.queue = nil
}
