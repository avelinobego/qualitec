package util

import (
	"reflect"
	"testing"
)

func TestSliceStrToInt16(t *testing.T) {
	// Dados colocando zeros nos erros
	testData := []struct {
		input  []string
		output []int16
	}{
		{[]string{}, []int16{}},             // Empty slice
		{[]string{""}, []int16{0}},          // Bad value
		{[]string{"32768"}, []int16{32767}}, // Overflow
		{[]string{"0", "1", "2", "3"}, []int16{0, 1, 2, 3}},
		{[]string{"1", "", "3"}, []int16{1, 0, 3}},  // Good and bad values in same slice
		{[]string{"1", "A", "B"}, []int16{1, 0, 0}}, // Good and bad values in same slice
	}

	// Dados ignorando erros
	testData2 := []struct {
		input  []string
		output []int16
	}{
		{[]string{}, []int16{}},        // Empty slice
		{[]string{""}, []int16{}},      // Bad value
		{[]string{"32768"}, []int16{}}, // Overflow
		{[]string{"0", "1", "2", "3"}, []int16{0, 1, 2, 3}},
		{[]string{"1", "", "3"}, []int16{1, 3}}, // Good and bad values in same slice
		{[]string{"1", "A", "B"}, []int16{1}},   // Good and bad values in same slice
	}

	for _, data := range testData {
		outputReal := SliceStrToInt16(data.input, false)
		if (len(data.output) > 0 || len(outputReal) > 0) && !reflect.DeepEqual(data.output, outputReal) {
			t.Errorf("Expected %v, got %v", data.output, outputReal)
		}
	}

	for _, data := range testData2 {
		outputReal := SliceStrToInt16(data.input, true)
		if (len(data.output) > 0 || len(outputReal) > 0) && !reflect.DeepEqual(data.output, outputReal) {
			t.Errorf("Expected %v, got %v", data.output, outputReal)
		}
	}

}

func TestSliceIntersectInt16(t *testing.T) {
	testData := []struct {
		input1 []int16
		input2 []int16
		output []int16
	}{
		{[]int16{1}, []int16{2}, []int16{}},
		{[]int16{1}, []int16{1}, []int16{1}},
		{[]int16{1, 1, 1}, []int16{1}, []int16{1}},
		{[]int16{1}, []int16{1, 1, 1}, []int16{1}},
		{[]int16{1, 1, 1}, []int16{1, 1, 1}, []int16{1}},
		{[]int16{1, 2, 3}, []int16{1, 2, 3}, []int16{1, 2, 3}},
		{[]int16{1, 2, 3}, []int16{3, 4, 5}, []int16{3}},
		{[]int16{9, 1, 2, 3, 3, 4, 5, 7}, []int16{3, 4, 5, 5, 1, 9}, []int16{9, 1, 3, 4, 5}},
	}

	for _, data := range testData {
		outputReal := SliceIntersectInt16(data.input1, data.input2)
		// Por alguma razão, DeepEqual retorna false quando len == 0 para os dois slices
		if (len(outputReal) > 0 || len(data.output) > 0) && !reflect.DeepEqual(data.output, outputReal) {
			t.Errorf("Expected %v, got %v", data.output, outputReal)
		}
	}
}
