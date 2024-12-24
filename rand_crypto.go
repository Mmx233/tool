package tool

import (
	"crypto/rand"
	"math/big"
)

// RandCryptoNum Num [min,max]
func RandCryptoNum(min, max *big.Int) (*big.Int, error) {
	val, err := rand.Int(rand.Reader, max.Sub(max, min).Add(max, big.NewInt(1)))
	if err != nil {
		return nil, err
	}
	return val.Add(min, val), nil
}

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
