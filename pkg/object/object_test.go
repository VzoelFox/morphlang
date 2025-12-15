package object

import "testing"

func TestObjectInspect(t *testing.T) {
	tests := []struct {
		obj      Object
		expected string
	}{
		{&Integer{Value: 10}, "10"},
		{&Integer{Value: -5}, "-5"},
		{&Boolean{Value: true}, "benar"},
		{&Boolean{Value: false}, "salah"},
		{&Null{}, "kosong"},
		{&String{Value: "Halo"}, "Halo"},
		{&String{Value: ""}, ""},
		{&String{Value: "Morph Language"}, "Morph Language"},
		{&Error{Message: "Salah"}, "Error di [:0:0]:\n  Salah\n"},
		{&Error{Message: "Division by zero"}, "Error di [:0:0]:\n  Division by zero\n"},
		{&ReturnValue{Value: &Integer{Value: 5}}, "5"},
		{&ReturnValue{Value: &Boolean{Value: true}}, "benar"},
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
		{&Integer{Value: 10}, INTEGER_OBJ},
		{&Boolean{Value: true}, BOOLEAN_OBJ},
		{&Null{}, NULL_OBJ},
		{&String{Value: "Halo"}, STRING_OBJ},
		{&Error{Message: "Err"}, ERROR_OBJ},
		{&ReturnValue{Value: &Integer{Value: 5}}, RETURN_VALUE_OBJ},
	}

	for _, tt := range tests {
		if tt.obj.Type() != tt.expected {
			t.Errorf("Type() wrong. expected=%q, got=%q", tt.expected, tt.obj.Type())
		}
	}
}
