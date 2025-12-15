package evaluator

import (
	"testing"

	"github.com/VzoelFox/morphlang/pkg/object"
)

func TestBuiltinFunctions(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{`panjang("")`, 0},
		{`panjang("empat")`, 5},
		{`panjang("hello world")`, 11},
		{`tipe(1)`, "INTEGER"},
		{`tipe("a")`, "STRING"},
		{`tipe(benar)`, "BOOLEAN"},
		{`adalah_galat(galat("oops"))`, true},
		{`adalah_galat(1)`, false},
		{`pesan_galat(galat("bad input"))`, "bad input"},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)

		switch expected := tt.expected.(type) {
		case int:
			testIntegerObject(t, evaluated, int64(expected))
		case string:
			// For string objects or error messages?
			// The current test helpers don't have testStringObject, let's look at it.
			// `tipe` returns a String object.
			// `pesan_galat` returns a String object.
			testStringObject(t, evaluated, expected)
		case bool:
			testBooleanObject(t, evaluated, expected)
		}
	}
}

func testStringObject(t *testing.T, obj object.Object, expected string) bool {
	result, ok := obj.(*object.String)
	if !ok {
		t.Errorf("object is not String. got=%T (%+v)", obj, obj)
		return false
	}
	if result.Value != expected {
		t.Errorf("object has wrong value. got=%q, want=%q", result.Value, expected)
		return false
	}
	return true
}
