package base

import (
	b32 "encoding/base32"
	"pulsar/src/handle"
)

type base32 struct {
	encoder *BEncoder
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

func (b *base32) Init(options string) (handle.Handler, error) {
	return &base32{NewBEncoder(b32.StdEncoding, 8, b32.StdPadding)}, nil
}

func (b *base32) Process(buf []byte, length int, decode bool) ([]byte, int, error) {
	var err error = nil

	if !decode {
		encoded := b.encoder.encode(buf[:length])
		return encoded, len(encoded), nil
	}
	buf, length, err = b.encoder.decode(buf[:length])
	return buf, length, err
}
