package vm

import (
	"testing"

	"github.com/VzoelFox/morphlang/pkg/compiler"
	"github.com/VzoelFox/morphlang/pkg/object"
)

func TestStackUnderflow(t *testing.T) {
	tests := []struct {
		name         string
		instructions []byte
	}{
		{"OpPop empty", []byte{byte(compiler.OpPop)}},
		{"OpAdd empty", []byte{byte(compiler.OpAdd)}},
		{"OpAdd partial", []byte{byte(compiler.OpLoadConst), 0, 0, byte(compiler.OpAdd)}}, // Push 1 const, need 2
		{"OpMinus empty", []byte{byte(compiler.OpMinus)}},
		{"OpIndex empty", []byte{byte(compiler.OpIndex)}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock constants if needed
			constants := []object.Object{&object.Integer{Value: 1}}

			bytecode := &compiler.Bytecode{
				Instructions: tt.instructions,
				Constants:    constants,
			}

			vm := New(bytecode)
			err := vm.Run()

			if err == nil {
				t.Fatalf("expected error, got nil")
			}

			if err.Error() != "stack underflow" {
				t.Errorf("expected 'stack underflow' error, got %q", err.Error())
			}
		})
	}
}
