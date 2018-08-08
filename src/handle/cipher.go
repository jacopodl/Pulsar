package handle

import (
	"stringsop"
	"io"
	"crypto/cipher"
	"crypto/aes"
	"crypto/des"
	"math/rand"
	"time"
)

type cipherSt struct {
	options []stringsop.Option
	cipher  string
	block   cipher.Block
}

func NewCipher() Handler {
	return &cipherSt{[]stringsop.Option{
		{"cipher", []string{"aes", "des", "tdes"}, "aes", "Set cipher type"},
		{"cipher-key", nil, "", "Set cipher key"}},
		"", nil}
}

func (c *cipherSt) Name() string {
	return "cipher"
}

func (c *cipherSt) Description() string {
	return "The cipher! (AES,DES,TDES)"
}

func (c *cipherSt) Init(options []string) error {
	var cipher_key = ""
	var err error = nil
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
			cipher_key = value
		}
	}
	switch c.cipher {
	case "aes":
		c.block, err = aes.NewCipher([]byte(cipher_key))
	case "des":
		c.block, err = des.NewCipher([]byte(cipher_key))
	case "tdes":
		c.block, err = des.NewTripleDESCipher([]byte(cipher_key))
	}
	return err
}

func (c *cipherSt) Options() []stringsop.Option {
	return c.options
}

func (c *cipherSt) Process(buf []byte, length int, decode bool) ([]byte, int, error) {
	buf, err := c.processCTR(buf[:length], decode)
	return buf, len(buf), err
}

func (c *cipherSt) processCTR(data []byte, decrypt bool) ([]byte, error) {
	if ! decrypt {
		ciphertext := make([]byte, c.block.BlockSize()+len(data))
		iv := ciphertext[:c.block.BlockSize()]
		rand.Seed(time.Now().UnixNano())
		if _, err := rand.Read(iv); err != nil {
			return nil, err
		}
		stream := cipher.NewCTR(c.block, iv)
		stream.XORKeyStream(ciphertext[c.block.BlockSize():], data)
		return ciphertext, nil
	}
	plaintext := make([]byte, len(data)-c.block.BlockSize())
	iv := data[:c.block.BlockSize()]
	stream := cipher.NewCTR(c.block, iv)
	stream.XORKeyStream(plaintext, data[c.block.BlockSize():])
	return plaintext, nil
}
