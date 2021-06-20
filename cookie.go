package tool

import (
	"strings"
)

type cookie struct{}

var Cookie cookie

// Decode 从header中解析cookie进map
func (*cookie) Decode(a string) map[string]string {
	var t = make(map[string]string)
	for _, v := range strings.Split(a, ";") {
		tt := strings.Split(strings.TrimSpace(v), "=")
		switch tt[0] { //滤去一些杂质
		case "path":
		case "HttpOnly":
		case "":
		default:
			t[tt[0]] = strings.TrimSpace(tt[1])
		}
	}
	return t
}
