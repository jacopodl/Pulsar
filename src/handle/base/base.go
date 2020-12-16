package base

type BEIface interface {
	EncodeToString(src []byte) string
	DecodeString(s string) ([]byte, error)
}

type BEncoder struct {
	buffer     []byte
	baseChunk  int
	stdPadding byte
	base       BEIface
}

func NewBEncoder(base BEIface, baseChunk int, stdPadding rune) *BEncoder {
	return &BEncoder{nil, baseChunk, byte(stdPadding), base}
}

func (be *BEncoder) encode(buf []byte) []byte {
	ebuf := []byte(be.base.EncodeToString(buf))
	return ebuf
}

func (be *BEncoder) decode(buf []byte) ([]byte, int, error) {
	var ready []byte = nil
	var length = len(buf)
	var prev = 0

	if be.buffer != nil {
		be.buffer = append(be.buffer, buf...)
		buf = be.buffer
		length = len(buf)
		be.buffer = nil
	}

	for prev != length {
		chunk, ok := SplitInputBuffer(buf[prev:], be.baseChunk, be.stdPadding)
		if !ok {
			be.buffer = append(be.buffer, buf[prev:prev+chunk]...)
			break
		}
		decoded, err := be.base.DecodeString(string(buf[prev : prev+chunk]))
		if err != nil {
			be.buffer = nil
			return nil, 0, err
		}
		ready = append(ready, decoded...)
		prev += chunk
	}
	return ready, len(ready), nil
}

func SplitInputBuffer(buf []byte, baseChunk int, paddingRune byte) (int, bool) {
	for i := baseChunk; i < len(buf); i += baseChunk {
		if buf[i-1] == paddingRune {
			return i, true
		}
	}
	return len(buf), len(buf)%baseChunk == 0
}
