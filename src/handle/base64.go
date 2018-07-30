package handle

import (
	b64 "encoding/base64"
	"stringsop"
)

type base64 struct{}

func NewBase64() Handler {
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
	if !decode {
		encoded := []byte(b64.StdEncoding.EncodeToString(buf[:length]))
		return encoded, len(encoded), nil
	}
	decoded, err := b64.StdEncoding.DecodeString(string(buf[:length]))
	if err != nil {
		return nil, 0, err
	}
	return decoded, len(decoded), nil
}
