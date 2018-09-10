package handle

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/des"
	"encoding/binary"
	"fmt"
	"math/rand"
	"strings"
	"time"
)

const (
	DEFAULTCIPHER = "aes"
	VERIFYBLOCK   = 4
	HEADERSIZE    = 2
)

type cipherSt struct {
	block  cipher.Block
	buffer []byte
}

func NewCipher() Handler {
	return &cipherSt{}
}

func (c *cipherSt) Name() string {
	return "cipher"
}

func (c *cipherSt) Description() string {
	return "CTR cipher - key|[aes|des|tdes#key]"
}

func (c *cipherSt) Init(options string) (Handler, error) {
	var hcipher = cipherSt{}
	var err error = nil

	algo, key := splitCipherOptions(options, DEFAULTCIPHER)
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
	var pdata []byte = nil
	buf = buf[:length]

	if !decode {
		ciphertext, err := encryptCTR(buf, c.block)
		return ciphertext, len(ciphertext), err
	}

	for {
		ciphertext, haveNext, next := c.cipherChunk(buf)
		if ciphertext == nil {
			return pdata, len(pdata), nil
		}

		plaintext, err := decryptCTR(ciphertext, c.block)
		if !haveNext && pdata == nil {
			return plaintext, len(plaintext), err
		}
		pdata = append(pdata, plaintext...)
		buf = buf[next:]
	}
}

func (c *cipherSt) cipherChunk(data []byte) ([]byte, bool, int) {
	if c.buffer != nil {
		if len(c.buffer) == 1 {
			c.buffer = append(c.buffer, data[0])
			data = data[1:]
		}
		length := int(binary.BigEndian.Uint16(c.buffer[:HEADERSIZE]))
		remain := int(length) - len(c.buffer) + HEADERSIZE
		if len(data) < remain {
			c.buffer = append(c.buffer, data...)
			return nil, false, -1
		}
		c.buffer = append(c.buffer, data[:remain]...)
		tmp := c.buffer
		c.buffer = nil
		return tmp, len(data) > remain, remain
	}

	if len(data) > 1 {
		length := int(binary.BigEndian.Uint16(data[:HEADERSIZE]))
		if len(data[HEADERSIZE:]) >= length {
			return data[:HEADERSIZE+length], len(data[HEADERSIZE+length:]) > 0, HEADERSIZE + length
		}
	}
	c.buffer = append(c.buffer, data...)
	return nil, false, -1
}

func encryptCTR(data []byte, block cipher.Block) ([]byte, error) {
	var vblk []byte = nil
	var err error = nil

	ciphertext := make([]byte, HEADERSIZE+block.BlockSize()+VERIFYBLOCK+len(data))

	binary.BigEndian.PutUint16(ciphertext[:HEADERSIZE], uint16(len(data)+block.BlockSize())+VERIFYBLOCK)

	iv := ciphertext[HEADERSIZE : HEADERSIZE+block.BlockSize()]
	if err = mkRandom(iv); err != nil {
		return nil, err
	}

	if vblk, err = mkKeyVerBlock(data, byte(block.BlockSize())); err != nil {
		return nil, err
	}

	stream := cipher.NewCTR(block, iv)
	stream.XORKeyStream(ciphertext[HEADERSIZE+block.BlockSize():], vblk)
	return ciphertext, nil
}

func decryptCTR(data []byte, block cipher.Block) ([]byte, error) {
	plaintext := make([]byte, len(data)-block.BlockSize()-HEADERSIZE)
	stream := cipher.NewCTR(block, data[HEADERSIZE:HEADERSIZE+block.BlockSize()])
	stream.XORKeyStream(plaintext, data[HEADERSIZE+block.BlockSize():])

	if !verifyKey(plaintext[:VERIFYBLOCK], byte(block.BlockSize())) {
		return nil, fmt.Errorf("invalid algorithm or key")
	}

	return plaintext[VERIFYBLOCK:], nil
}

func mkRandom(data []byte) error {
	rand.Seed(time.Now().UnixNano())
	_, err := rand.Read(data)
	return err
}

func mkKeyVerBlock(data []byte, blksize byte) ([]byte, error) {
	block := make([]byte, VERIFYBLOCK)
	if err := mkRandom(block); err != nil {
		return nil, err
	}

	//   |0|1|2|3|
	// 1) ^---^ <-----cipher block size
	// 2)   ^---^

	// 1
	block[0] = blksize
	block[2] = blksize
	// 2
	block[1] = block[3]

	return append(block, data...), nil
}

func verifyKey(block []byte, blksize byte) bool {
	return blksize == block[0] && block[0] == block[2] && block[1] == block[3]
}

func splitCipherOptions(options, defcipher string) (algo, key string) {
	split := strings.SplitN(options, "#", 2)
	if len(split) > 1 {
		algo = split[0]
		key = split[1]
		return
	}
	algo = defcipher
	key = split[0]
	return
}
