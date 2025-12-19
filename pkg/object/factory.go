package object

import (
	"fmt"
	"github.com/VzoelFox/morphlang/pkg/memory"
)

// FromPtr rehydrates an Object Wrapper from a raw Memory Pointer.
// It reads the Header to determine the Type.
func FromPtr(ptr memory.Ptr) Object {
	if ptr == memory.NilPtr {
		// NilPtr usually maps to Null object in logic,
		// but if we have explicit TagNull, we might get a ptr to it.
		// For safety, return nil or a default Null wrapper with NilPtr?
		// Let's return nil and let caller handle, or &Null{Address: NilPtr}.
		return &Null{Address: memory.NilPtr}
	}

	header, err := memory.ReadHeader(ptr)
	if err != nil {
		// Panic on memory corruption/invalid pointer as we can't recover easily
		panic(fmt.Sprintf("FromPtr: failed to read header for ptr %d: %v", ptr, err))
	}

	switch header.Type {
	case 0: // Uninitialized or Tag 0
		return &Null{Address: ptr}
	case memory.TagInteger:
		return &Integer{Address: ptr}
	case memory.TagBoolean:
		return &Boolean{Address: ptr}
	case memory.TagFloat:
		return &Float{Address: ptr}
	case memory.TagString:
		return &String{Address: ptr}
	case memory.TagNull:
		return &Null{Address: ptr}
	case memory.TagArray:
		return &Array{Address: ptr}
	case memory.TagHash:
		return &Hash{Address: ptr}
	case memory.TagCompiledFunction:
		return &CompiledFunction{Address: ptr}
	case memory.TagClosure:
		return &Closure{Address: ptr}
	case memory.TagBuiltin:
		idx, err := memory.ReadBuiltin(ptr)
		if err != nil {
			panic(err)
		}
		if idx < 0 || idx >= len(Builtins) {
			panic(fmt.Sprintf("invalid builtin index %d", idx))
		}
		// Return a new wrapper with Address and the Function pointer from registry
		return &Builtin{Address: ptr, Fn: Builtins[idx].Builtin.Fn}
	case memory.TagError:
		return &Error{Address: ptr}
	case memory.TagResource:
		return GetResource(ptr)
	case memory.TagPointer:
		return &Pointer{Address: ptr}
	case memory.TagModule:
		return &Module{Address: ptr}
	default:
		// Fallback or Panic
		panic(fmt.Sprintf("FromPtr: unknown type tag %d", header.Type))
	}
}
