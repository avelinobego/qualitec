package database

import (
	"testing"
)

type s struct {
	str     string
	integer int
}

func TestMarker_Mark(t *testing.T) {
	var s s
	var mark Marker

	s.integer = 20
	s.str = "100"

	// Precisa ser inicializado vazio
	if mark.Marks() != nil {
		t.Errorf("Not NIL")
	}

	// Adição
	mark.Mark("integer", &s.integer)
	t1 := mark.Marks()["integer"]
	if s.integer != *t1.(*int) {
		t.Errorf("Experado valor %d, obtido %d", s.integer, *t1.(*int))
	}

	// Substituição por outro tipo
	mark.Mark("integer", &s.str)
	t1 = mark.Marks()["integer"]
	if s.str != *t1.(*string) {
		t.Errorf("Experado valor %s, obtido %s", s.str, *t1.(*string))
	}

	// Adição de segundo tipo
	mark.Mark("string", &s.str)
	t1 = mark.Marks()["string"]
	if s.str != *t1.(*string) {
		t.Errorf("Experado valor %s, obtido %s", s.str, *t1.(*string))
	}

	// Alteração de variável de origem
	s.integer = 40
	s.str = "a"
	mark.Mark("integer", &s.str)
	t1 = mark.Marks()["integer"]
	if s.str != *t1.(*string) {
		t.Errorf("Experado valor %s, obtido %s", s.str, *t1.(*string))
	}
	mark.Mark("string", &s.str)
	t1 = mark.Marks()["string"]
	if s.str != *t1.(*string) {
		t.Errorf("Experado valor %s, obtido %s", s.str, *t1.(*string))
	}

	// Limpeza
	mark.UnMark()
	if mark.Marks() != nil {
		t.Errorf("UnMark falhou!")
	}

}

func TestMarker_MakeNamedSQL(t *testing.T) {
	var s s
	var mark Marker

	s.integer = 20
	s.str = "100"
	mark.Mark("integer", &s.integer)
	mark.Mark("string", &s.str)

	sql := mark.MakeNamedSQL()
	sqlExpected := "integer=:integer, string=:string"
	if sql != sqlExpected {
		t.Errorf("Esperado SQL %s, obtido %s", sqlExpected, sql)
	}
}
