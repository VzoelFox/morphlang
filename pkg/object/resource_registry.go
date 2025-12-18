package object

import (
	"sync"
	"sync/atomic"
	"github.com/VzoelFox/morphlang/pkg/memory"
)

var (
	resourceRegistry sync.Map
	resourceIDGen    int64
)

// RegisterResource stores a host object and returns a memory-backed wrapper (Address).
func RegisterResource(obj Object) memory.Ptr {
	id := atomic.AddInt64(&resourceIDGen, 1)
	resourceRegistry.Store(id, obj)

	ptr, err := memory.AllocResource(id)
	if err != nil { panic(err) }
	return ptr
}

// GetResource retrieves the host object from a memory pointer.
func GetResource(ptr memory.Ptr) Object {
	id, err := memory.ReadResource(ptr)
	if err != nil { panic(err) }

	val, ok := resourceRegistry.Load(id)
	if !ok { return nil }
	return val.(Object)
}
