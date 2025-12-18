package memory

import (
	"unsafe"
)

// Layout: [Header][NumLocals(4)][NumParams(4)][InstrLen(4)][Instructions...]
// Note: Instructions usually byte array. Padded to 8 bytes.

func AllocCompiledFunction(instructions []byte, numLocals, numParams int) (Ptr, error) {
	instrLen := len(instructions)
	// Payload: 4+4+4 = 12 bytes + instrLen
	payloadSize := 12 + instrLen
	totalSize := HeaderSize + payloadSize
	allocSize := (totalSize + 7) & ^7

	Lemari.mu.Lock()
	defer Lemari.mu.Unlock()

	ptr, err := Lemari.alloc(allocSize)
	if err != nil { return NilPtr, err }

	raw, err := Lemari.resolve(ptr)
	if err != nil { return NilPtr, err }

	// Header
	header := (*Header)(raw)
	header.Type = TagCompiledFunction
	header.Size = uint32(allocSize)

	// Body
	bodyPtr := unsafe.Pointer(uintptr(raw) + uintptr(HeaderSize))

	// Write NumLocals
	*(*int32)(bodyPtr) = int32(numLocals)

	// Write NumParams
	paramPtr := unsafe.Pointer(uintptr(bodyPtr) + 4)
	*(*int32)(paramPtr) = int32(numParams)

	// Write InstrLen
	lenPtr := unsafe.Pointer(uintptr(paramPtr) + 4)
	*(*int32)(lenPtr) = int32(instrLen)

	// Write Instructions
	if instrLen > 0 {
		instrPtr := unsafe.Pointer(uintptr(lenPtr) + 4)
		dest := unsafe.Slice((*byte)(instrPtr), instrLen)
		copy(dest, instructions)
	}

	return ptr, nil
}

// ReadCompiledFunction reads metadata and instructions.
func ReadCompiledFunction(ptr Ptr) ([]byte, int, int, error) {
	Lemari.mu.Lock()
	defer Lemari.mu.Unlock()

	raw, err := Lemari.resolve(ptr)
	if err != nil { return nil, 0, 0, err }

	bodyPtr := unsafe.Pointer(uintptr(raw) + uintptr(HeaderSize))
	numLocals := int(*(*int32)(bodyPtr))

	paramPtr := unsafe.Pointer(uintptr(bodyPtr) + 4)
	numParams := int(*(*int32)(paramPtr))

	lenPtr := unsafe.Pointer(uintptr(paramPtr) + 4)
	instrLen := int(*(*int32)(lenPtr))

	instr := make([]byte, instrLen)
	if instrLen > 0 {
		instrPtr := unsafe.Pointer(uintptr(lenPtr) + 4)
		src := unsafe.Slice((*byte)(instrPtr), instrLen)
		copy(instr, src)
	}

	return instr, numLocals, numParams, nil
}

// ReadCompiledFunctionMeta reads only metadata (Locals, Params).
func ReadCompiledFunctionMeta(ptr Ptr) (int, int, error) {
	Lemari.mu.Lock()
	defer Lemari.mu.Unlock()

	raw, err := Lemari.resolve(ptr)
	if err != nil { return 0, 0, err }

	bodyPtr := unsafe.Pointer(uintptr(raw) + uintptr(HeaderSize))
	numLocals := int(*(*int32)(bodyPtr))

	paramPtr := unsafe.Pointer(uintptr(bodyPtr) + 4)
	numParams := int(*(*int32)(paramPtr))

	return numLocals, numParams, nil
}

// Layout: [Header][FnPtr(8)][FreeCount(4)][FreePtr0(8)]...
func AllocClosure(fnPtr Ptr, freeVars []Ptr) (Ptr, error) {
	freeCount := len(freeVars)
	// Payload: 8 + 4 + (8 * count)
	payloadSize := 12 + (8 * freeCount)
	totalSize := HeaderSize + payloadSize
	allocSize := (totalSize + 7) & ^7

	Lemari.mu.Lock()
	defer Lemari.mu.Unlock()

	ptr, err := Lemari.alloc(allocSize)
	if err != nil { return NilPtr, err }

	raw, err := Lemari.resolve(ptr)
	if err != nil { return NilPtr, err }

	header := (*Header)(raw)
	header.Type = TagClosure
	header.Size = uint32(allocSize)

	bodyPtr := unsafe.Pointer(uintptr(raw) + uintptr(HeaderSize))

	// Write FnPtr
	// Ptr is uint64 (8 bytes)
	*(*uint64)(bodyPtr) = uint64(fnPtr)

	// Write FreeCount
	countPtr := unsafe.Pointer(uintptr(bodyPtr) + 8)
	*(*int32)(countPtr) = int32(freeCount)

	// Write FreeVars
	if freeCount > 0 {
		varsPtr := unsafe.Pointer(uintptr(countPtr) + 4)
		// We treat Ptr as uint64
		// Target slice of uint64
		target := unsafe.Slice((*uint64)(varsPtr), freeCount)
		for i, v := range freeVars {
			target[i] = uint64(v)
		}
	}

	return ptr, nil
}

func ReadClosure(ptr Ptr) (Ptr, []Ptr, error) {
	Lemari.mu.Lock()
	defer Lemari.mu.Unlock()

	raw, err := Lemari.resolve(ptr)
	if err != nil { return NilPtr, nil, err }

	bodyPtr := unsafe.Pointer(uintptr(raw) + uintptr(HeaderSize))
	fnPtr := Ptr(*(*uint64)(bodyPtr))

	countPtr := unsafe.Pointer(uintptr(bodyPtr) + 8)
	freeCount := int(*(*int32)(countPtr))

	freeVars := make([]Ptr, freeCount)
	if freeCount > 0 {
		varsPtr := unsafe.Pointer(uintptr(countPtr) + 4)
		src := unsafe.Slice((*uint64)(varsPtr), freeCount)
		for i, v := range src {
			freeVars[i] = Ptr(v)
		}
	}

	return fnPtr, freeVars, nil
}
