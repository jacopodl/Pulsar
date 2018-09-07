package base

import (
	b32 "encoding/base32"
	"handle"
	"stringsop"
)

type base32 struct {
	buffer []byte
}

func NewBase32() handle.Handler {
	return &base32{}
}

func (b *base32) Name() string {
	return "base32"
}

func (b *base32) Description() string {
	return "Base32 encoder/decoder"
}

func (b *base32) Init(options []string) error {
	return nil
}

func (b *base32) Options() []stringsop.Option {
	return nil
}

func (b *base32) Process(buf []byte, length int, decode bool) ([]byte, int, error) {
	var ready []byte = nil
	var prev = 0

	if !decode {
		encoded := []byte(b32.StdEncoding.EncodeToString(buf[:length]))
		return encoded, len(encoded), nil
	}

	if b.buffer != nil {
		b.buffer = append(b.buffer, buf[:length]...)
		buf = b.buffer
		length = len(buf)
		b.buffer = nil
	}

	for prev != length {
		chunk, ok := SplitInputBuffer(buf[prev:], 8, byte(b32.StdPadding))
		if !ok {
			b.buffer = append(b.buffer, buf[prev:prev+chunk]...)
			break
		}
		decoded, err := b32.StdEncoding.DecodeString(string(buf[prev : prev+chunk]))
		if err != nil {
			b.buffer = nil
			return nil, 0, err
		}
		ready = append(ready, decoded...)
		prev += chunk
	}
	return ready, len(ready), nil
}
