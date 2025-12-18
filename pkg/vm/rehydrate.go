package vm

import (
	"github.com/VzoelFox/morphlang/pkg/memory"
	"github.com/VzoelFox/morphlang/pkg/object"
)

// Rehydrate converts a memory.Ptr back to a Go object.Object wrapper.
// It delegates to object.FromPtr.
func Rehydrate(ptr memory.Ptr) (object.Object, error) {
	// object.FromPtr panics on error (Fail Fast)
	return object.FromPtr(ptr), nil
}
