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

func (d *Dns) AddQuestion(dname string, qtype uint16, qclass uint16) {
	query := Query{qtype, qclass}
	qname := dname2qname(dname)

	// Try to compress question
	qname = d.compress(qname)

	d.Data = append(d.Data, qname...)
	d.Data = append(d.Data, query.Serialize()...)
	d.TotalQuestions++
}

func (d *Dns) compress(question []byte) []byte {
	var qidx uint16 = 0
	var ridx uint16 = 0
	var qtot uint16 = 1
	var sridx uint16 = 0
	var sqidx uint16 = 0

	if d.TotalQuestions == 0 {
		return question
	}

	for {
		lq := uint16(question[qidx])
		lr := uint16(d.Data[ridx])

		if lr == 0 && lq == 0 {
			break
		}

		if lr&0xC0 == 0xC0 {
			ridx = uint16(d.Data[ridx+1]) - 12
			continue
		}

		if lq != lr {
			ridx += lr + 1
			if lr == 0 {
				qtot++
				if qtot > d.TotalQuestions {
					qidx += lq + 1
					if question[qidx] == 0 {
						return question
					}
					ridx = 0
					qtot = 1
					continue
				}
				ridx += 4 // Len of Query struct
			}
			continue
		}
		if !bytes.Equal(d.Data[ridx:ridx+lr], question[qidx:qidx+lq]) {
			ridx += lr + 1
			sridx = 0
			sqidx = 0
			continue
		}
		if sridx == 0 {
			sridx = ridx
			sqidx = qidx
		}
		ridx += lr + 1
		qidx += lq + 1

	}
	binary.BigEndian.PutUint16(question[sqidx:sqidx+2], sridx+12|0xC000)
	question = question[:sqidx+2]
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
