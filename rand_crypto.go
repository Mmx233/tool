package tool

import (
	"crypto/rand"
	"math/big"
	"unsafe"
)

func RandCrypto(letters string) (r RandCryptoLetters) {
	r.letters = letters
	return
}

type RandCryptoLetters struct {
	letters string
}

func (r RandCryptoLetters) String(Len int) (string, error) {
	b := make([]byte, Len)
	letterLen := big.NewInt(int64(len(r.letters)))
	for i := range b {
		idx, err := rand.Int(rand.Reader, letterLen)
		if err != nil {
			return "", err
		}
		b[i] = r.letters[idx.Int64()]
	}
	return unsafe.String(unsafe.SliceData(b), len(b)), nil
}
