package vm

import (
	"github.com/VzoelFox/morphlang/pkg/compiler"
	"github.com/VzoelFox/morphlang/pkg/memory"
	"github.com/VzoelFox/morphlang/pkg/object"
)

type Frame struct {
	cl          *object.Closure
	ip          int
	basePointer int
}

func NewFrame(cl *object.Closure, basePointer int) *Frame {
	return &Frame{cl: cl, ip: -1, basePointer: basePointer}
}

func (f *Frame) Instructions() compiler.Instructions {
	instr, _, _, err := memory.ReadCompiledFunction(f.cl.Fn.Address)
	if err != nil {
		// Panic is acceptable here as it indicates memory corruption/system failure during execution
		panic(err)
	}
	return instr
}
