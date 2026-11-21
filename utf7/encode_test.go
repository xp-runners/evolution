package utf7

import (
	"reflect"
	"testing"
)

func assertEqual(expect, actual interface{}, t *testing.T) {
	if !reflect.DeepEqual(expect, actual) {
		t.Errorf("Items not equal:\nexpected %q\nhave     %q\n", expect, actual)
	}
}

func Test_encoding_empty(t *testing.T) {
	assertEqual("", Encode(""), t)
}

func Test_encoding_us_ascii(t *testing.T) {
	fixtures := map[string]string{
		"hello":       "hello",
		"hello world": "hello world",
	}

	for input, result := range fixtures {
		t.Run(input, func(t *testing.T) {
			assertEqual(result, Encode(input), t)
		})
	}
}

func Test_encoding_special_chars(t *testing.T) {
	fixtures := map[string]string{
		"£1":        "+AKM-1",
		"£†":        "+AKMgIA-",
		"1 + 1 = 2": "1 +- 1 +AD0 2",
		"1&1":       "1+ACY-1",
		"£-":        "+AKM--",
	}

	for input, result := range fixtures {
		t.Run(input, func(t *testing.T) {
			assertEqual(result, Encode(input), t)
		})
	}
}

func Test_encoding_utf8(t *testing.T) {
	fixtures := map[string]string{
		"Über":    "+ANw-ber",
		"中华人民共和国": "+Ti1TTk66bBFRcVSMVv0-",
	}

	for input, result := range fixtures {
		t.Run(input, func(t *testing.T) {
			assertEqual(result, Encode(input), t)
		})
	}
}

func Test_encoding_emoji(t *testing.T) {
	fixtures := map[string]string{
		"😀":  "+2D3eAA-",
		"☀":  "+JgA-",
		"🇩🇪": "+2Dzd6dg83eo-",
	}

	for input, result := range fixtures {
		t.Run(input, func(t *testing.T) {
			assertEqual(result, Encode(input), t)
		})
	}
}
