package tool

import (
	"math/rand"
	"unsafe"
)

const (
	letterIdBits = 6
	letterIdMask = 1<<letterIdBits - 1
	letterIdMax  = 63 / letterIdBits
)

type Rand struct {
	source rand.Source
	rand   *rand.Rand
}

func NewRand(src rand.Source) Rand {
	return Rand{
		source: src,
		rand:   rand.New(src),
	}
}

// Num [min,max]
func (r Rand) Num(Min, Max int) int {
	return Min + rand.Intn(Max-Min+1)
}

func (r Rand) WithLetters(letters string) RandWithLetters {
	return RandWithLetters{
		Rand:    r,
		letters: letters,
	}
}

type RandWithLetters struct {
	Rand
	letters string
}

func (r RandWithLetters) String(Len int) string {
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
	return *(*string)(unsafe.Pointer(&b))
}
