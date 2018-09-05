package dns

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

const (
	TYPE_A     uint16 = 1
	TYPE_NS    uint16 = 2
	TYPE_CNAME uint16 = 5
	TYPE_SOA   uint16 = 6
	TYPE_PTR   uint16 = 12
	TYPE_MX    uint16 = 15
	TYPE_TXT   uint16 = 16
	TYPE_AAAA  uint16 = 28

	CLASS_IN  uint16 = 1
	CLASS_CH  uint16 = 3
	CLASS_HS  uint16 = 4
	CLASS_ANY uint16 = 255

	DNSHDRSIZE = 12
	MAXLBLSIZE = 63
)

type Dns struct {
	Id              uint16
	Info            uint16
	TotalQuestions  uint16
	TotalAnswers    uint16
	TotalAuthority  uint16
	TotalAdditional uint16
	Data            []byte
}

func NewDnsPacket(id uint16) *Dns {
	packet := Dns{id, 0, 0, 0, 0, 0, nil}
	return &packet
}

func Deserialize(buf []byte) (*Dns, error) {
	if len(buf) < DNSHDRSIZE {
		return nil, fmt.Errorf("malformed packet")
	}

	pkt := Dns{}
	pkt.Id = binary.BigEndian.Uint16(buf[:2])
	pkt.Info = binary.BigEndian.Uint16(buf[2:4])
	pkt.TotalQuestions = binary.BigEndian.Uint16(buf[4:6])
	pkt.TotalAnswers = binary.BigEndian.Uint16(buf[6:8])
	pkt.TotalAuthority = binary.BigEndian.Uint16(buf[8:10])
	pkt.TotalAdditional = binary.BigEndian.Uint16(buf[10:])

	pkt.Data = append([]byte{}, buf[12:]...)
	return &pkt, nil
}

func getQuestion(buf []byte, start uint16) ([]string, uint16) {
	var question []string
	var last uint16 = 0
	for lr := uint16(buf[start]); lr != 0; lr = uint16(buf[start]) {
		start++
		if lr&0xC0 == 0xC0 {
			if last == 0 {
				last = start + 5
			}
			start = uint16(buf[start]) - 12
			continue
		}
		question = append(question, string(buf[start:start+lr]))
		start += lr
	}
	if last == 0 {
		last = start + 5
	}
	return question, last
}

func (d *Dns) AddQuestion(query *Query) {
	qname := query.Serialize()

	// Try to compress question
	qname = d.compress(qname)

	d.Data = append(d.Data, qname...)
	d.TotalQuestions++
}

func (d *Dns) compress(question []byte) []byte {
	var qidx uint16 = 0
	var ridx uint16 = 0
	var qtot uint16 = 1

	if d.TotalQuestions == 0 {
		return question
	}

	for {
		if ok, l1, _ := compareQuestions(question, qidx, d.Data, ridx); ok {
			binary.BigEndian.PutUint16(question[qidx:qidx+2], ridx+DNSHDRSIZE|0xC000)
			question = append(question[:qidx+2], question[l1+1:]...)
			break
		}
		qidx += uint16(question[qidx]) + 1
		if question[qidx] == 0 {
			ridx += uint16(d.Data[ridx]) + 1
			if ridx == 0 {
				if qtot++; qtot > d.TotalQuestions {
					break
				}
				ridx += getQuery(d.Data[ridx:]) + QUERYSIZE
			}
			qidx = 0
		}
	}
	return question
}

func (d *Dns) GetQuestions() [][]string {
	var questions [][]string
	var question []string
	var qidx uint16 = 0

	for total := uint16(0); total < d.TotalQuestions; total++ {
		question, qidx = getQuestion(d.Data, qidx)
		questions = append(questions, question)
	}

	return questions
}

func (d *Dns) Serialize() []byte {
	data := make([]byte, 12)

	binary.BigEndian.PutUint16(data[:2], d.Id)
	binary.BigEndian.PutUint16(data[2:4], d.Info)
	binary.BigEndian.PutUint16(data[4:6], d.TotalQuestions)
	binary.BigEndian.PutUint16(data[6:8], d.TotalAnswers)
	binary.BigEndian.PutUint16(data[8:10], d.TotalAuthority)
	binary.BigEndian.PutUint16(data[10:], d.TotalAdditional)

	data = append(data, d.Data...)
	return data
}

func getQuery(buf []byte) (query uint16) {
	var i byte = 0
	for ; buf[i] != 0; i += buf[i] + 1 {
		if buf[i] == 0xC0 {
			return uint16(i + 1)
		}
	}
	return uint16(i + 1)
}

func compareQuestions(q1 []byte, idx1 uint16, q2 []byte, idx2 uint16) (bool, uint16, uint16) {
	for q1[idx1] != 0 && q2[idx2] != 0 {
		l1 := uint16(q1[idx1])
		l2 := uint16(q2[idx2])
		idx1++
		idx2++
		if l1 == 0xC0 {
			idx1 += uint16(q1[idx1+1]) - DNSHDRSIZE
			continue
		}
		if l2 == 0xC0 {
			idx2 += uint16(q1[idx2+1]) - DNSHDRSIZE
			continue
		}
		if l1 != l2 || !bytes.Equal(q1[idx1:idx1+l1], q2[idx2:idx2+l2]) {
			return false, 0, 0
		}
		idx1 += l1
		idx2 += l2
	}
	return true, idx1, idx2
}
