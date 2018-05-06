package utf7

import (
	"encoding/base64"
	"strings"
	"unicode/utf16"
)

var literal = " \r\n\tABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789'(),-./:?"
var baseSet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"
var encoder = base64.NewEncoding(baseSet)

func encodeSection(b []byte) string {
	return "+" + strings.TrimRight(encoder.EncodeToString(b), "=")
}

func appendRune(b []byte, r rune) []byte {
	r1, r2 := utf16.EncodeRune(r)
	if r1 == 0xfffd {
		return append(b, byte(r>>8), byte(r))
	} else {
		return append(b, byte(r1>>8), byte(r1), byte(r2>>8), byte(r2))
	}
}

// Encode takes an input string and encodes it to UTF-7 as per RFC 2152
func Encode(in string) string {
	e := ""
	section := false

	var b []byte
	for _, r := range in {
		if section {
			if strings.IndexRune(baseSet, r) > -1 {
				e += encodeSection(b) + "-" + string(r)
				section = false
			} else if strings.IndexRune(literal, r) > -1 {
				e += encodeSection(b) + string(r)
				section = false
			} else {
				b = appendRune(b, r)
			}
		} else {
			if r == '&' || r == '+' {
				e += string(r) + "-"
			} else if strings.IndexRune(literal, r) > -1 {
				e += string(r)
			} else {
				section = true
				b = appendRune([]byte{}, r)
			}
		}
	}

	if section {
		return e + encodeSection(b) + "-"
	} else {
		return e
	}
}
