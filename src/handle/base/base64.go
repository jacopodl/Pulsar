package base

import (
	b64 "encoding/base64"
	"pulsar/src/handle"
)

type base64 struct {
	encoder *BEncoder
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

func (b *base64) Init(options string) (handle.Handler, error) {
	return &base64{NewBEncoder(b64.StdEncoding, 4, b64.StdPadding)}, nil
}

func (b *base64) Process(buf []byte, length int, decode bool) ([]byte, int, error) {
	var err error = nil

	if !decode {
		encoded := b.encoder.encode(buf[:length])
		return encoded, len(encoded), nil
	}
	buf, length, err = b.encoder.decode(buf[:length])
	return buf, length, err
}
