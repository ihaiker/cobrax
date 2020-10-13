package cobrax

import (
	"bytes"
	"strings"
	"unicode"
)

// 驼峰式写法转为下划线写法
func camel2Case(name string, split rune) string {
	buffer := bytes.NewBuffer([]byte{})
	for i, r := range name {
		if unicode.IsUpper(r) {
			if i != 0 {
				buffer.WriteRune(split)
			}
			buffer.WriteRune(unicode.ToLower(r))
		} else {
			buffer.WriteRune(r)
		}
	}
	return buffer.String()
}

func envName(name string) string {
	envName := camel2Case(name, '_')
	buffer := bytes.NewBuffer([]byte{})
	for _, r := range envName {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			buffer.WriteRune(unicode.ToUpper(r))
		} else {
			buffer.WriteRune('_')
		}
	}
	return strings.ToUpper(buffer.String())
}
