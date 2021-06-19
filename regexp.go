package tool

import (
	"regexp"
)

type reg struct{}

var Regexp reg

func (reg) MatchExisting(reg string, a string) (bool, error) {
	m, e := regexp.Match(reg, []byte(a))
	if e != nil {
		return false, e
	}
	return m, nil
}

func (reg) MatchValue(reg string, a string) ([][]string, error) {
	r, e := regexp.Compile(reg)
	if e != nil {
		return nil, e
	}
	m := r.FindAllStringSubmatch(a, -1)
	return m, nil
}

func (reg) Replace(reg string, o string, n string) (string, error) {
	r, e := regexp.Compile(reg)
	if e != nil {
		return o, e
	}
	return r.ReplaceAllString(o, n), e
}
