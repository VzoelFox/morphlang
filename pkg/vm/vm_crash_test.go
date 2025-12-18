package vm

import (
	"strings"
	"testing"

	"github.com/VzoelFox/morphlang/pkg/compiler"
	"github.com/VzoelFox/morphlang/pkg/lexer"
	"github.com/VzoelFox/morphlang/pkg/object"
	"github.com/VzoelFox/morphlang/pkg/parser"
)

func TestVMMonitorCrash(t *testing.T) {
	// Register a builtin that panics
	object.RegisterBuiltin("__crash_test", func(args ...object.Object) object.Object {
		panic("intentional crash for monitor test")
	})

	input := `__crash_test()`
	l := lexer.New(input)
	p := parser.New(l)
	prog := p.ParseProgram()

	comp := compiler.New()
	err := comp.Compile(prog)
	if err != nil {
		t.Fatalf("compile error: %s", err)
	}

	vm := New(comp.Bytecode())

	// Redirect stdout to capture DumpState?
	// Difficult in test. We check the returned error message.

	err = vm.Run()
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "VM CRASH (Monitor Recovered)") {
		t.Errorf("wrong error message. got=%q", err.Error())
	}
	if !strings.Contains(err.Error(), "intentional crash for monitor test") {
		t.Errorf("error message missing panic cause. got=%q", err.Error())
	}
}
