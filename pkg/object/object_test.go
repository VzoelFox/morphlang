package object

import (
	"testing"
)

func TestStringHash(t *testing.T) {
	// Placeholder for hash tests if we implement Hashable interface later
}

func TestObjectInspect(t *testing.T) {
	tests := []struct {
		obj      Object
		expected string
	}{
		{&Integer{Value: 10}, "10"},
		{&Boolean{Value: true}, "benar"},
		{&Boolean{Value: false}, "salah"},
		{&String{Value: "Halo"}, "Halo"},
		{&Null{}, "kosong"},
		{&Error{Message: "terjadi kesalahan"}, "ERROR: terjadi kesalahan"},
		{&ReturnValue{Value: &Integer{Value: 5}}, "5"},
	}

	for _, tt := range tests {
		if tt.obj.Inspect() != tt.expected {
			t.Errorf("wrong inspect value. expected=%q, got=%q",
				tt.expected, tt.obj.Inspect())
		}
	}
}

func TestObjectType(t *testing.T) {
	tests := []struct {
		obj      Object
		expected ObjectType
	}{
		{&Integer{Value: 10}, INTEGER_OBJ},
		{&Boolean{Value: true}, BOOLEAN_OBJ},
		{&String{Value: "Halo"}, STRING_OBJ},
		{&Null{}, NULL_OBJ},
		{&Error{Message: "msg"}, ERROR_OBJ},
		{&ReturnValue{Value: &Integer{Value: 5}}, RETURN_VALUE_OBJ},
	}

	for _, tt := range tests {
		if tt.obj.Type() != tt.expected {
			t.Errorf("wrong type. expected=%q, got=%q",
				tt.expected, tt.obj.Type())
		}
	}
}
