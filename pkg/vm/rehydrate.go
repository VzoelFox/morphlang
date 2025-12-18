package vm

import (
	"fmt"
	"github.com/VzoelFox/morphlang/pkg/memory"
	"github.com/VzoelFox/morphlang/pkg/object"
)

// Rehydrate converts a memory.Ptr back to a Go object.Object wrapper (Hybrid).
func Rehydrate(ptr memory.Ptr) (object.Object, error) {
	if ptr == memory.NilPtr {
		return &object.Null{}, nil
	}

	header, err := memory.ReadHeader(ptr)
	if err != nil { return nil, err }

	switch header.Type {
	case memory.TagInteger:
		val, err := memory.ReadInteger(ptr)
		if err != nil { return nil, err }
		return &object.Integer{Value: val, Address: ptr}, nil
	case memory.TagString:
		val, err := memory.ReadString(ptr)
		if err != nil { return nil, err }
		return &object.String{Value: val, Address: ptr}, nil
	case memory.TagBoolean:
		val, err := memory.ReadBoolean(ptr)
		if err != nil { return nil, err }
		return &object.Boolean{Value: val, Address: ptr}, nil
	case memory.TagBuiltin:
		idx, err := memory.ReadBuiltin(ptr)
		if err != nil { return nil, err }
		if idx < 0 || idx >= len(object.Builtins) {
			return nil, fmt.Errorf("invalid builtin index %d", idx)
		}
		return object.Builtins[idx].Builtin, nil
	case memory.TagClosure:
		fnPtr, freePtrs, err := memory.ReadClosure(ptr)
		if err != nil { return nil, err }

		cf := &object.CompiledFunction{Address: fnPtr}

		freeObjs := make([]object.Object, len(freePtrs))
		for i, p := range freePtrs {
			obj, err := Rehydrate(p)
			if err != nil { return nil, err }
			freeObjs[i] = obj
		}

		return &object.Closure{Fn: cf, FreeVariables: freeObjs}, nil

	case memory.TagCompiledFunction:
		return &object.CompiledFunction{Address: ptr}, nil
	case memory.TagHash:
		count, err := memory.ReadHashCount(ptr)
		if err != nil { return nil, err }

		pairs := make(map[object.HashKey]object.HashPair)
		for i := 0; i < count; i++ {
			kPtr, vPtr, err := memory.ReadHashPair(ptr, i)
			if err != nil { return nil, err }

			kObj, err := Rehydrate(kPtr)
			if err != nil { return nil, err }
			vObj, err := Rehydrate(vPtr)
			if err != nil { return nil, err }

			hashKey, ok := kObj.(object.Hashable)
			if !ok { return nil, fmt.Errorf("unusable as hash key") }

			pairs[hashKey.HashKey()] = object.HashPair{Key: kObj, Value: vObj}
		}
		return &object.Hash{Pairs: pairs}, nil

	default:
		return nil, fmt.Errorf("rehydrate: unknown tag %d", header.Type)
	}
}
