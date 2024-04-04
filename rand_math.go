package tool

import (
	"math/rand"
	"unsafe"
)

func RandMath(src rand.Source) RandMathNum {
	return RandMathNum{
		source: src,
		rand:   rand.New(src),
	}
}

const (
	letterIdBits = 6
	letterIdMask = 1<<letterIdBits - 1
	letterIdMax  = 63 / letterIdBits
)

type RandMathNum struct {
	source rand.Source
	rand   *rand.Rand
}

// Num [min,max]
func (r RandMathNum) Num(Min, Max int) int {
	return Min + r.rand.Intn(Max-Min+1)
}

func (r RandMathNum) WithLetters(letters string) RandMathWithLetters {
	return RandMathWithLetters{
		RandMathNum: r,
		letters:     letters,
	}
}

type RandMathWithLetters struct {
	RandMathNum
	letters string
}

func (r RandMathWithLetters) String(Len int) string {
	b := make([]byte, Len)
	for i, cache, remain := Len-1, r.source.Int63(), letterIdMax; i >= 0; {
		if remain == 0 {
			cache, remain = r.source.Int63(), letterIdMax
		}
		if idx := int(cache & letterIdMask); idx < len(r.letters) {
			b[i] = r.letters[idx]
			i--
		}
		cache >>= letterIdBits
		remain--
	}
	return unsafe.String(unsafe.SliceData(b), len(b))
}
