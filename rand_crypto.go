package tool

import (
	"crypto/rand"
	"math/big"
)

func RandCrypto(letters string) (r RandCryptoLetters) {
	r.letters = letters
	return
}

type RandCryptoLetters struct {
	letters string
}

func (r RandCryptoLetters) Text(Len int) ([]byte, error) {
	b := make([]byte, Len)
	letterLen := big.NewInt(int64(len(r.letters)))
	for i := range b {
		idx, err := rand.Int(rand.Reader, letterLen)
		if err != nil {
			return nil, err
		}
		b[i] = r.letters[idx.Int64()]
	}
	return b, nil
}
