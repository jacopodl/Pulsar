package handle

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/des"
	"fmt"
	"math/rand"
	"strings"
	"time"
)

const DEFAULTCIPHER = "aes"

type cipherSt struct {
	block cipher.Block
}

func NewCipher() Handler {
	return &cipherSt{}
}

func (c *cipherSt) Name() string {
	return "cipher"
}

func (c *cipherSt) Description() string {
	return "CTR cipher - cipher:key|[aes|des|tdes#key]"
}

func (c *cipherSt) Init(options string) (Handler, error) {
	var hcipher = cipherSt{}
	var err error = nil

	algo, key := splitCipherOptions(options)
	switch algo {
	case "aes":
		hcipher.block, err = aes.NewCipher([]byte(key))
	case "des":
		hcipher.block, err = des.NewCipher([]byte(key))
	case "tdes":
		hcipher.block, err = des.NewTripleDESCipher([]byte(key))
	default:
		return nil, fmt.Errorf("invalid cipher: %s", algo)
	}

	if err == nil {
		return &hcipher, nil
	}

	return nil, err
}

func (c *cipherSt) Process(buf []byte, length int, decode bool) ([]byte, int, error) {
	buf, err := c.processCTR(buf[:length], decode)
	return buf, len(buf), err
}

func (c *cipherSt) processCTR(data []byte, decrypt bool) ([]byte, error) {
	if !decrypt {
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

func splitCipherOptions(options string) (algo, key string) {
	split := strings.SplitN(options, "#", 2)
	if len(split) > 1 {
		algo = split[0]
		key = split[1]
		return
	}
	algo = DEFAULTCIPHER
	key = split[0]
	return
}
