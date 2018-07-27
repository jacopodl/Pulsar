package handle

import b64 "encoding/base64"

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

func (b *base64) Process(buf []byte, length int, decode bool) ([]byte, int, error) {
	if !decode {
		encoded := []byte(b64.StdEncoding.EncodeToString(buf))
		return encoded, len(encoded), nil
	}
	decoded, err := b64.StdEncoding.DecodeString(string(buf))
	if err != nil {
		return nil, 0, err
	}
	return decoded, len(decoded), nil
}
