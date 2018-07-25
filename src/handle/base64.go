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

func (b *base64) Process(buf []byte, decode bool) ([]byte, error) {
	if !decode {
		return []byte(b64.StdEncoding.EncodeToString(buf)), nil
	}
	decoded, err := b64.StdEncoding.DecodeString(string(buf))
	if err != nil {
		return nil, err
	}
	return decoded, nil
}
