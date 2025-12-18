package vm

import (
	"os"
	"testing"

	"github.com/VzoelFox/morphlang/pkg/object"
)

func TestIOBuiltins(t *testing.T) {
	tmpFile := "test_io.txt"
	content := "Hello Morph I/O"

	// Ensure cleanup
	defer os.Remove(tmpFile)

	tests := []vmTestCase{
		{
			input: `tulis_file("` + tmpFile + `", "` + content + `")`,
			expected: nil, // tulis_file returns null
		},
		{
			input: `baca_file("` + tmpFile + `")`,
			expected: content,
		},
	}

	runVmTests(t, tests)
}

func TestIOErrors(t *testing.T) {
	tests := []vmTestCase{
		{
			input: `baca_file("non_existent_file.txt")`,
			expected: object.NewError("baca_file error: open non_existent_file.txt: no such file or directory", "", 0, 0),
		},
	}

	runVmTests(t, tests)
}
