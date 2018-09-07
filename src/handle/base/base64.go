package base

import (
	b64 "encoding/base64"
	"handle"
	"stringsop"
)

type base64 struct {
	buffer []byte
}

func NewBase64() handle.Handler {
	return &base64{}
}

func (b *base64) Name() string {
	return "base64"
}

func (b *base64) Description() string {
	return "Base64 encoder/decoder"
}

func (b *base64) Init(options []string) error {
	return nil
}

func (b *base64) Options() []stringsop.Option {
	return nil
}

func (b *base64) Process(buf []byte, length int, decode bool) ([]byte, int, error) {
	var ready []byte = nil
	var prev = 0

	if !decode {
		encoded := []byte(b64.StdEncoding.EncodeToString(buf[:length]))
		return encoded, len(encoded), nil
	}

	if b.buffer != nil {
		b.buffer = append(b.buffer, buf[:length]...)
		buf = b.buffer
		length = len(buf)
		b.buffer = nil
	}

	for prev != length {
		chunk, ok := SplitInputBuffer(buf[prev:], 4, byte(b64.StdPadding))
		if !ok {
			b.buffer = append(b.buffer, buf[prev:prev+chunk]...)
			break
		}
		decoded, err := b64.StdEncoding.DecodeString(string(buf[prev : prev+chunk]))
		if err != nil {
			b.buffer = nil
			return nil, 0, err
		}
		ready = append(ready, decoded...)
		prev += chunk
	}
	return ready, len(ready), nil
}
