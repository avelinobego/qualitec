package database

import "testing"

func TestMakeIn(t *testing.T) {
	testData := []struct {
		input  uint
		output string
	}{
		{0, "()"},
		{1, "(?)"},
		{2, "(?,?)"},
		{3, "(?,?,?)"},
		{4, "(?,?,?,?)"},
		{5, "(?,?,?,?,?)"},
		{10, "(?,?,?,?,?,?,?,?,?,?)"},
	}

	for _, data := range testData {
		outputReal := MakeIn(data.input)
		if data.output != outputReal {
			t.Errorf("Expected %s, got %s", data.output, outputReal)
		}
	}
}

func TestMakeAndWhere(t *testing.T) {
	testData := []struct {
		object Clauses
		output string
	}{
		{Clauses{""}, ""},
		{Clauses{"A"}, "A"},
		{Clauses{"A", "B"}, "A AND B"},
		{Clauses{"A", "B", "C"}, "A AND B AND C"},
		{Clauses{"A", "(B OR K)", "C"}, "A AND (B OR K) AND C"},
	}

	for _, data := range testData {
		outputReal := data.object.MakeAndWhere(false)
		if data.output != outputReal {
			t.Errorf("Expected %s, got %s", data.output, outputReal)
		}

		outputReal = data.object.MakeAndWhere(true)
		if ("WHERE " + data.output) != outputReal {
			t.Errorf("Expected %s, got %s", "WHERE " + data.output, outputReal)
		}

	}
}
