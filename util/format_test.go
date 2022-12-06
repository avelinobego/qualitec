package util

import (
	"testing"
)

var (
	dataCEP = [][2]string{
		{"12345678", "12345-678"},
		{"1", "00000-001"},
		{"", "00000-000"},
		{" 1 2 3 4", " 1 2 -3 4"},
		{" ", "00000-00 "},
		{"123456789", "12345-678"},
		{"ABCDEFGHIJKLMN", "ABCDE-FGH"},
	}

	dataPhone = [][2]string{
		{"41020694", "4102-0694"},
		{"997273858", "99727-3858"},
		{"1141020694", "(11) 4102-0694"},
		{"11997273858", "(11) 99727-3858"},
		{"123", "123"},
		{"123456789ABCDEFGH", "123456789ABCDEFGH"},
		{"", ""},
		{"ABCDEFGH", "ABCD-EFGH"},
	}
)

func TestFormatCEP(t *testing.T) {
	for _, val := range dataCEP {
		cep := FormatCEP(val[0])
		if cep != val[1] {
			t.Errorf("Expected %s, got %s", val[1], cep)
		}
	}
}

func BenchmarkFormatCEP(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, val := range dataCEP {
			FormatCEP(val[0])
		}
	}
}

func TestFormatPhone(t *testing.T) {
	for _, val := range dataPhone {
		cep := FormatPhone(val[0])
		if cep != val[1] {
			t.Errorf("Expected %s, got %s", val[1], cep)
		}
	}
}

func BenchmarkFormatPhone(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, val := range dataCEP {
			FormatPhone(val[0])
		}
	}
}
