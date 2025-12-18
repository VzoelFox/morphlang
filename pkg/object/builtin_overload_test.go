package object

import (
	"testing"
)

func TestBuiltinOverloading(t *testing.T) {
	// 1. Define Base Builtin
	RegisterBuiltin("test_overload", func(args ...Object) Object {
		return &String{Value: "ORIGINAL"}
	})

	// 2. Overload with function that fails due to Argument Mismatch
	RegisterBuiltin("test_overload", func(args ...Object) Object {
		if len(args) != 1 {
			// This returns an error with Code=E008 (MissingArgs)
			return newArgumentError(len(args), 1)
		}
		return &String{Value: "NEW"}
	})

	// 3. Call with 0 args -> Should trigger "NEW", fail (E008), then fallback to "ORIGINAL"
	// Get the registered function (which is the composed one)
	idx := GetBuiltinByName("test_overload")
	fn := Builtins[idx].Builtin.Fn

	result := fn() // 0 args
	if result.Inspect() != "ORIGINAL" {
		t.Errorf("Overloading fallback failed. Expected 'ORIGINAL', got %s", result.Inspect())
	}

	// 4. Call with 1 arg -> Should succeed in "NEW"
	result = fn(&Null{})
	if result.Inspect() != "NEW" {
		t.Errorf("Overloading new function failed. Expected 'NEW', got %s", result.Inspect())
	}
}
