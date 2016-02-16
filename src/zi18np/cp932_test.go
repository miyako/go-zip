package zi18np

import (
	"testing"
)

func TestSjis2Utf8(t *testing.T) {
	// SJISで"GO言語"
	word := string([]byte{0x47, 0x4f, 0x8c, 0xbe, 0x8c, 0xea})
	utf8, err := Sjis2Utf8(word)
	if err != nil {
		t.Errorf("got error[%v]", err)
	}
	if utf8 != "GO言語" {
		t.Errorf("word[%v] utf8[%v]", word, utf8)
	}
}

func TestUtf82Sjis(t *testing.T) {
	// SJISで"GO言語"
	word := string([]byte{0x47, 0x4f, 0x8c, 0xbe, 0x8c, 0xea})
	sjis, err := Sjis2Utf8("GO言語")
	if err != nil {
		t.Errorf("got error[%v]", err)
	}
	if utf8 != "GO言語" {
		t.Errorf("sjis[%v] utf8[%v]", sjis, utf8)
	}
}
