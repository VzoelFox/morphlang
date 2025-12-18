package object

import "testing"

func TestObjectInspect(t *testing.T) {
	tests := []struct {
		obj      Object
		expected string
	}{
		{NewInteger(10), "10"},
		{NewInteger(-5), "-5"},
		{NewBoolean(true), "benar"},
		{NewBoolean(false), "salah"},
		{NewNull(), "kosong"},
		{NewString("Halo"), "Halo"},
		{NewString(""), ""},
		{NewString("Morph Language"), "Morph Language"},
		{&Error{Message: "Salah"}, "Error di [:0:0]:\n  Salah\n"},
		{&Error{Message: "Division by zero"}, "Error di [:0:0]:\n  Division by zero\n"},
		{&ReturnValue{Value: NewInteger(5)}, "5"},
		{&ReturnValue{Value: NewBoolean(true)}, "benar"},
	}

	for _, tt := range tests {
		if tt.obj.Inspect() != tt.expected {
			t.Errorf("Inspect() wrong. expected=%q, got=%q", tt.expected, tt.obj.Inspect())
		}
	}
}

func TestObjectType(t *testing.T) {
	tests := []struct {
		obj      Object
		expected ObjectType
	}{
		{NewInteger(10), INTEGER_OBJ},
		{NewBoolean(true), BOOLEAN_OBJ},
		{NewNull(), NULL_OBJ},
		{NewString("Halo"), STRING_OBJ},
		{&Error{Message: "Err"}, ERROR_OBJ},
		{&ReturnValue{Value: NewInteger(5)}, RETURN_VALUE_OBJ},
	}

	for _, tt := range tests {
		if tt.obj.Type() != tt.expected {
			t.Errorf("Type() wrong. expected=%q, got=%q", tt.expected, tt.obj.Type())
		}
	}
}
