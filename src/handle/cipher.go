package handle

import (
	"stringsop"
	"io"
)

type cipher struct {
	options []stringsop.Option
	cipher  string
	key     string
}

func NewCipher() Handler {
	return &cipher{[]stringsop.Option{
		{"cipher", []string{"aes", "des", "tdes"}, "aes", "Set cipher type"},
		{"cipher-key", nil, "", "Set cipher key"}},
		"", ""}
}

func (c *cipher) Name() string {
	return "cipher"
}

func (c *cipher) Description() string {
	return "The cipher! (AES,DES,TDES)"
}

func (c *cipher) Init(options []string) error {
	parser := stringsop.NewStringsOp(options, c.options)
	for {
		key, value, err := parser.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		switch key {
		case "cipher":
			c.cipher = value
		case "cipher-key":
			c.key = key
		}
	}
	return nil
}

func (c *cipher) Options() []stringsop.Option {
	return c.options
}

func (c *cipher) Process(buf []byte, length int, decode bool) ([]byte, int, error) {
	return buf, length, nil
}
