package tool

import (
	"math/rand"
	"time"
)

type ranD struct{}

var Rand ranD

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ!@#$%^&*()_+1234567890")

// RandNum [min,max]
func (*ranD) RandNum(Min int, Max int) int {
	rand.Seed(time.Now().UnixNano())
	return Min + rand.Intn(Max-Min+1)
}

func (*ranD) RandString(Len int) string {
	b := make([]rune, Len)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
