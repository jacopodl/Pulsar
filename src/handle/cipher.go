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

const (
	DEFAULTCIPHER = "aes"
	VERIFYBLOCK   = 6
)

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

	if !decode {
		ciphertext, err := encryptCTR(buf[:length], c.block)
		return ciphertext, len(ciphertext), err
	}
	plaintext, err := decryptCTR(buf[:length], c.block)
	return plaintext, len(plaintext), err
}

func encryptCTR(data []byte, block cipher.Block) ([]byte, error) {
	var vblock []byte = nil
	var err error = nil

	ciphertext := make([]byte, block.BlockSize()+VERIFYBLOCK+len(data))

	iv := ciphertext[:block.BlockSize()]

	rand.Seed(time.Now().UnixNano())
	if _, err = rand.Read(iv); err != nil {
		return nil, err
	}

	if vblock, err = mkKeyVerBlock(data, byte(block.BlockSize())); err != nil {
		return nil, err
	}

	stream := cipher.NewCTR(block, iv)
	stream.XORKeyStream(ciphertext[block.BlockSize():], vblock)
	return ciphertext, nil
}

func decryptCTR(data []byte, block cipher.Block) ([]byte, error) {
	plaintext := make([]byte, len(data)-block.BlockSize())
	stream := cipher.NewCTR(block, data[:block.BlockSize()])
	stream.XORKeyStream(plaintext, data[block.BlockSize():])

	if !verifyKey(plaintext[:VERIFYBLOCK], byte(block.BlockSize())) {
		return nil, fmt.Errorf("invalid algorithm or key")
	}

	return plaintext[VERIFYBLOCK:], nil
}

func mkKeyVerBlock(data []byte, blocksize byte) ([]byte, error) {
	block := make([]byte, VERIFYBLOCK)
	rand.Seed(time.Now().UnixNano())
	if _, err := rand.Read(block); err != nil {
		return nil, err
	}

	//   |0|1|2|3|4|5|
	// 1) ^-------^ <-----cipher block size
	// 2)     ^-----^
	// 3)   ^---^

	// 1
	block[0] = blocksize
	block[4] = block[0]
	// 2
	block[2] = block[5]
	// 3
	block[1] = block[3]

	return append(block, data...), nil
}

func verifyKey(block []byte, blocksize byte) bool {
	return blocksize == block[0] && block[0] == block[4] && block[2] == block[5] && block[1] == block[3]
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
