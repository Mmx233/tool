package tool

import (
	"math/rand"
)

type ranD struct{}

var Rand ranD

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")

// Num [min,max]
func (ranD) Num(Min int, Max int) int {
	return Min + rand.Intn(Max-Min+1)
}

func (ranD) String(Len int) string {
	b := make([]rune, Len)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
