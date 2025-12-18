package memory

import "unsafe"

// Scan returns a list of pointers to the child pointers contained in the object.
// Assumes Lemari.mu is Locked.
func Scan(ptr Ptr) ([]*Ptr, error) {
	if ptr == NilPtr { return nil, nil }

	raw, err := Lemari.resolve(ptr)
	if err != nil { return nil, err }

	header := (*Header)(raw)
	base := uintptr(raw) + uintptr(HeaderSize)
	children := []*Ptr{}

	switch header.Type {
	case TagArray:
		// Layout: [Capacity(4)][Length(4)][Ptrs...]
		// lenPtr at offset 4
		lenPtr := unsafe.Pointer(base + 4)
		length := int(*(*int32)(lenPtr))
		elementsStart := base + 8
		for i := 0; i < length; i++ {
			p := (*Ptr)(unsafe.Pointer(elementsStart + uintptr(i*8)))
			// We return address of the pointer so GC can update it
			children = append(children, p)
		}

	case TagHash:
		// Layout: [Count(4)][Key0][Val0]...
		countPtr := unsafe.Pointer(base)
		count := int(*(*int32)(countPtr))
		elementsStart := base + 4
		for i := 0; i < count; i++ {
			k := (*Ptr)(unsafe.Pointer(elementsStart + uintptr(i*16)))
			v := (*Ptr)(unsafe.Pointer(elementsStart + uintptr(i*16) + 8))
			children = append(children, k, v)
		}

	case TagClosure:
		// Layout: [FnPtr(8)][FreeCount(4)][FreePtrs...]
		// FnPtr is a CompiledFunction object (pointer)
		fnPtr := (*Ptr)(unsafe.Pointer(base))
		children = append(children, fnPtr)

		countPtr := unsafe.Pointer(base + 8)
		count := int(*(*int32)(countPtr))
		elementsStart := base + 12
		for i := 0; i < count; i++ {
			p := (*Ptr)(unsafe.Pointer(elementsStart + uintptr(i*8)))
			children = append(children, p)
		}

	case TagError:
		// Layout: [MsgPtr(8)][CodePtr(8)]...
		msgPtr := (*Ptr)(unsafe.Pointer(base))
		codePtr := (*Ptr)(unsafe.Pointer(base + 8))
		children = append(children, msgPtr, codePtr)

	case TagPointer:
		// Layout: [Ptr(8)]
		p := (*Ptr)(unsafe.Pointer(base))
		children = append(children, p)
	}

	return children, nil
}
