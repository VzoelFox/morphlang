package vm

import (
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/VzoelFox/morphlang/pkg/compiler"
	"github.com/VzoelFox/morphlang/pkg/memory"
	"github.com/VzoelFox/morphlang/pkg/object"
	"github.com/VzoelFox/morphlang/pkg/scheduler"
)

const StackSize = 2048
const GlobalSize = 65536
const MaxFrames = 1024

var (
	True          = &object.Boolean{Value: true}
	False         = &object.Boolean{Value: false}
	Null          = &object.Null{}
	TruePtr       memory.Ptr
	FalsePtr      memory.Ptr
	NullPtr       memory.Ptr = memory.NilPtr
	bootstrapOnce sync.Once
	schedulerOnce sync.Once
	taskRegistry  sync.Map
	taskIDGen     int64
)

type TaskContext struct {
	Closure   *object.Closure
	Globals   []memory.Ptr
	Constants []object.Object
	ResultCh  chan object.Object
}

type VMSnapshot struct {
	Stack       []memory.Ptr
	Globals     []memory.Ptr
	Frames      []*Frame
	SP          int
	FramesIndex int
}

type VM struct {
	constants []object.Object
	globals   []memory.Ptr

	stack [StackSize]memory.Ptr
	sp    int

	frames      []*Frame
	framesIndex int

	LastPoppedStackElem object.Object

	snapshots []VMSnapshot

	Cabinet *memory.Cabinet
	Drawer  *memory.Drawer
}

func New(bytecode *compiler.Bytecode) *VM {
	ptr, err := memory.AllocCompiledFunction(bytecode.Instructions, 0, 0)
	if err != nil {
		panic(fmt.Sprintf("vm boot error: %v", err))
	}
	mainFn := &object.CompiledFunction{Address: ptr}
	mainClosure := &object.Closure{Fn: mainFn}
	mainFrame := NewFrame(mainClosure, 0)

	frames := make([]*Frame, MaxFrames)
	frames[0] = mainFrame

	if len(memory.Lemari.Drawers) == 0 {
		memory.InitCabinet()
	}

	bootstrapOnce.Do(func() {
		if TruePtr == 0 {
			TruePtr, _ = memory.AllocBoolean(true)
		}
		if FalsePtr == 0 {
			FalsePtr, _ = memory.AllocBoolean(false)
		}
	})

	schedulerOnce.Do(func() {
		scheduler.Init(4, executeTask)
	})

	drawer := &memory.Lemari.Drawers[0]

	return &VM{
		constants:   bytecode.Constants,
		globals:     make([]memory.Ptr, GlobalSize),
		stack:       [StackSize]memory.Ptr{},
		sp:          0,
		frames:      frames,
		framesIndex: 1,
		snapshots:   make([]VMSnapshot, 0),
		Cabinet:     &memory.Lemari,
		Drawer:      drawer,
	}
}

func (vm *VM) currentFrame() *Frame {
	return vm.frames[vm.framesIndex-1]
}

func (vm *VM) pushFrame(f *Frame) error {
	if vm.framesIndex >= MaxFrames {
		return fmt.Errorf("stack overflow")
	}
	vm.frames[vm.framesIndex] = f
	vm.framesIndex++
	return nil
}

func (vm *VM) popFrame() *Frame {
	vm.framesIndex--
	return vm.frames[vm.framesIndex]
}

func (vm *VM) StackTop() object.Object {
	if vm.sp == 0 {
		return nil
	}
	ptr := vm.stack[vm.sp-1]
	obj, _ := Rehydrate(ptr)
	return obj
}

func (vm *VM) DumpState() {
	fmt.Printf("\n=== VM MONITOR CRASH DUMP ===\n")
	if vm.framesIndex > 0 {
		frame := vm.currentFrame()
		fmt.Printf("IP: %d\n", frame.ip)
		fmt.Printf("Function Locals: %d\n", frame.cl.Fn.NumLocals())
	}
	fmt.Printf("Stack Pointer: %d\n", vm.sp)
	fmt.Printf("=============================\n")
}

func (vm *VM) Run() (err error) {
	defer func() {
		if r := recover(); r != nil {
			vm.DumpState()
			err = fmt.Errorf("VM CRASH: %v", r)
		}
	}()

	var ip int
	var ins compiler.Instructions
	var op compiler.Opcode

	for vm.currentFrame().ip < len(vm.currentFrame().Instructions())-1 {
		vm.currentFrame().ip++

		ip = vm.currentFrame().ip
		ins = vm.currentFrame().Instructions()
		op = compiler.Opcode(ins[ip])

		switch op {
		case compiler.OpLoadConst:
			constIndex := compiler.ReadUint16(ins[ip+1:])
			vm.currentFrame().ip += 2
			obj := vm.constants[constIndex]
			if err := ensureOnHeap(obj); err != nil { return err }
			if err := vm.push(getObjectAddress(obj)); err != nil { return err }

		case compiler.OpAdd, compiler.OpSub, compiler.OpMul, compiler.OpDiv:
			if err := vm.executeBinaryOperation(op); err != nil { return err }

		case compiler.OpPop:
			if _, err := vm.pop(); err != nil { return err }

		case compiler.OpDup:
			top := vm.stack[vm.sp-1] // Unsafe peek
			if err := vm.push(top); err != nil { return err }

		case compiler.OpStoreGlobal:
			globalIndex := compiler.ReadUint16(ins[ip+1:])
			vm.currentFrame().ip += 2
			val, err := vm.pop()
			if err != nil { return err }
			vm.globals[globalIndex] = val

		case compiler.OpLoadGlobal:
			globalIndex := compiler.ReadUint16(ins[ip+1:])
			vm.currentFrame().ip += 2
			if err := vm.push(vm.globals[globalIndex]); err != nil { return err }

		case compiler.OpStoreLocal:
			localIndex := int(ins[ip+1])
			vm.currentFrame().ip += 1
			frame := vm.currentFrame()
			val, err := vm.pop()
			if err != nil { return err }
			vm.stack[frame.basePointer+localIndex] = val

		case compiler.OpLoadLocal:
			localIndex := int(ins[ip+1])
			vm.currentFrame().ip += 1
			frame := vm.currentFrame()
			if err := vm.push(vm.stack[frame.basePointer+localIndex]); err != nil { return err }

		case compiler.OpGetBuiltin:
			builtinIndex := int(ins[ip+1])
			vm.currentFrame().ip += 1
			ptr, err := memory.AllocBuiltin(builtinIndex)
			if err != nil { return err }
			if err := vm.push(ptr); err != nil { return err }

		case compiler.OpCall:
			numArgs := int(ins[ip+1])
			vm.currentFrame().ip += 1
			if err := vm.executeCall(numArgs); err != nil { return err }

		case compiler.OpReturnValue:
			returnValue, err := vm.pop()
			if err != nil { return err }
			frame := vm.popFrame()
			vm.sp = frame.basePointer - 1
			if err := vm.push(returnValue); err != nil { return err }

		case compiler.OpReturn:
			frame := vm.popFrame()
			vm.sp = frame.basePointer - 1
			if err := vm.push(NullPtr); err != nil { return err }

		case compiler.OpClosure:
			constIndex := compiler.ReadUint16(ins[ip+1:])
			numFree := int(ins[ip+3])
			vm.currentFrame().ip += 3
			if err := vm.pushClosure(int(constIndex), numFree); err != nil { return err }

		case compiler.OpGetFree:
			freeIndex := int(ins[ip+1])
			vm.currentFrame().ip += 1
			currentClosure := vm.currentFrame().cl
			obj := currentClosure.FreeVariables[freeIndex]
			if err := ensureOnHeap(obj); err != nil { return err }
			if err := vm.push(getObjectAddress(obj)); err != nil { return err }

		case compiler.OpArray:
			numElements := int(compiler.ReadUint16(ins[ip+1:]))
			vm.currentFrame().ip += 2
			if err := vm.buildArray(vm.sp-numElements, vm.sp); err != nil { return err }
			vm.sp -= numElements

		case compiler.OpHash:
			numElements := int(compiler.ReadUint16(ins[ip+1:]))
			vm.currentFrame().ip += 2
			if err := vm.buildHash(vm.sp-numElements, vm.sp); err != nil { return err }
			vm.sp -= numElements

		case compiler.OpIndex:
			index, err := vm.pop()
			if err != nil { return err }
			left, err := vm.pop()
			if err != nil { return err }
			if err := vm.executeIndexExpression(left, index); err != nil { return err }

		case compiler.OpSetIndex:
			val, err := vm.pop()
			if err != nil { return err }
			index, err := vm.pop()
			if err != nil { return err }
			left, err := vm.pop()
			if err != nil { return err }
			if err := vm.executeSetIndexExpression(left, index, val); err != nil { return err }

		case compiler.OpEqual, compiler.OpNotEqual, compiler.OpGreaterThan, compiler.OpGreaterEqual:
			if err := vm.executeComparison(op); err != nil { return err }

		case compiler.OpBang:
			if err := vm.executeBangOperator(); err != nil { return err }

		case compiler.OpMinus:
			if err := vm.executeMinusOperator(); err != nil { return err }

		case compiler.OpBitNot:
			if err := vm.executeBitNotOperator(); err != nil { return err }

		case compiler.OpAnd, compiler.OpOr, compiler.OpXor, compiler.OpLShift, compiler.OpRShift:
			if err := vm.executeBitwiseOperation(op); err != nil { return err }

		default:
			// return fmt.Errorf("unknown opcode %d", op)
		}
	}
	return nil
}

func (vm *VM) push(ptr memory.Ptr) error {
	if vm.sp >= StackSize {
		return fmt.Errorf("stack overflow")
	}
	vm.stack[vm.sp] = ptr
	vm.sp++
	return nil
}

func (vm *VM) pop() (memory.Ptr, error) {
	if vm.sp == 0 {
		return memory.NilPtr, fmt.Errorf("stack underflow")
	}
	ptr := vm.stack[vm.sp-1]
	vm.sp--
	obj, _ := Rehydrate(ptr)
	vm.LastPoppedStackElem = obj
	return ptr, nil
}

func (vm *VM) pushClosure(constIndex int, numFree int) error {
	constant := vm.constants[constIndex]
	fn := constant.(*object.CompiledFunction)

	freeVars := make([]memory.Ptr, numFree)
	freeObjs := make([]object.Object, numFree)
	for i := 0; i < numFree; i++ {
		ptr := vm.stack[vm.sp-numFree+i]
		freeVars[i] = ptr
		obj, _ := Rehydrate(ptr)
		freeObjs[i] = obj
	}
	vm.sp -= numFree

	ptr, err := memory.AllocClosure(fn.Address, freeVars)
	if err != nil { return err }

	return vm.push(ptr)
}

func (vm *VM) executeCall(numArgs int) error {
	calleePtr := vm.stack[vm.sp-1-numArgs]

	header, err := memory.ReadHeader(calleePtr)
	if err != nil { return err }

	if header.Type == memory.TagClosure {
		fnPtr, _, err := memory.ReadClosure(calleePtr)
		if err != nil { return err }

		fnWrapper := &object.CompiledFunction{Address: fnPtr}

		_, freePtrs, _ := memory.ReadClosure(calleePtr)
		freeObjs := make([]object.Object, len(freePtrs))
		for i, p := range freePtrs {
			freeObjs[i], _ = Rehydrate(p)
		}

		clWrapper := &object.Closure{Fn: fnWrapper, FreeVariables: freeObjs}

		if numArgs != fnWrapper.NumParameters() {
			return fmt.Errorf("arg mismatch")
		}

		frame := NewFrame(clWrapper, vm.sp-numArgs)
		vm.pushFrame(frame)
		vm.sp = frame.basePointer + fnWrapper.NumLocals()
		return nil
	}

	if header.Type == memory.TagBuiltin {
		return vm.executeBuiltinCall(calleePtr, numArgs)
	}

	return fmt.Errorf("calling non-function")
}

func (vm *VM) executeBuiltinCall(builtinPtr memory.Ptr, numArgs int) error {
	idx, _ := memory.ReadBuiltin(builtinPtr)
	builtin := object.Builtins[idx].Builtin

	args := make([]object.Object, numArgs)
	for i := 0; i < numArgs; i++ {
		ptr := vm.stack[vm.sp-numArgs+i]
		args[i], _ = Rehydrate(ptr)
	}

	res := builtin.Fn(args...)

	if errObj, ok := res.(*object.Error); ok {
		if errObj.Message == "luncurkan() requires VM context" {
			tObj, err := vm.spawn(args)
			if err != nil {
				res = &object.Error{Message: err.Error()}
			} else {
				res = tObj
			}
		}
	}

	vm.sp -= (numArgs + 1)

	ensureOnHeap(res)
	return vm.push(getObjectAddress(res))
}

func (vm *VM) spawn(args []object.Object) (*object.Thread, error) {
	cl := args[0].(*object.Closure)

	taskID := atomic.AddInt64(&taskIDGen, 1)
	newGlobals := make([]memory.Ptr, len(vm.globals))
	copy(newGlobals, vm.globals)
	resultCh := make(chan object.Object, 1)

	ctx := TaskContext{
		Closure: cl,
		Globals: newGlobals,
		Constants: vm.constants,
		ResultCh: resultCh,
	}
	taskRegistry.Store(taskID, ctx)

	ptr, _ := memory.AllocInteger(taskID)
	scheduler.Submit(ptr)

	return &object.Thread{Result: resultCh}, nil
}

func executeTask(ptr memory.Ptr) {
	id, _ := memory.ReadInteger(ptr)
	val, ok := taskRegistry.Load(id)
	if !ok { return }
	taskRegistry.Delete(id)
	ctx := val.(TaskContext)

	frames := make([]*Frame, MaxFrames)
	frames[0] = NewFrame(ctx.Closure, 0)

	newVM := &VM{
		constants: ctx.Constants,
		globals: ctx.Globals,
		stack: [StackSize]memory.Ptr{},
		sp: ctx.Closure.Fn.NumLocals(),
		frames: frames,
		framesIndex: 1,
		Cabinet: &memory.Lemari,
	}

	err := newVM.Run()
	if err != nil {
		ctx.ResultCh <- &object.Error{Message: err.Error()}
	} else {
		if newVM.LastPoppedStackElem != nil {
			ctx.ResultCh <- newVM.LastPoppedStackElem
		} else {
			ctx.ResultCh <- Null
		}
	}
	close(ctx.ResultCh)
}

func (vm *VM) snapshot() error {
	// Not implemented for Phase X.4 MVP (Requires rewriting snapshot struct)
	return nil
}

func (vm *VM) rollback() error { return nil }
func (vm *VM) commit() error { return nil }

func ensureOnHeap(obj object.Object) error {
	switch o := obj.(type) {
	case *object.Integer:
		if o.Address != 0 { return nil }
		ptr, err := memory.AllocInteger(o.Value)
		if err != nil { return err }
		o.Address = ptr
	case *object.String:
		if o.Address != 0 { return nil }
		ptr, err := memory.AllocString(o.Value)
		if err != nil { return err }
		o.Address = ptr
	case *object.Boolean:
		if o.Address != 0 { return nil }
		ptr, err := memory.AllocBoolean(o.Value)
		if err != nil { return err }
		o.Address = ptr
	}
	return nil
}

func getObjectAddress(obj object.Object) memory.Ptr {
	switch o := obj.(type) {
	case *object.Integer: return o.Address
	case *object.String: return o.Address
	case *object.Boolean: return o.Address
	case *object.Null: return memory.NilPtr
	}
	return 0
}
