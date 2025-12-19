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
	True          *object.Boolean
	False         *object.Boolean
	Null          *object.Null
	TruePtr       memory.Ptr
	FalsePtr      memory.Ptr
	NullPtr       memory.Ptr = memory.NilPtr
	bootstrapOnce sync.Once
	schedulerOnce sync.Once
	taskRegistry  sync.Map
	taskIDGen     int64
	activeVMs     sync.Map
	GlobalVMLock  sync.RWMutex
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

	LastPoppedPtr memory.Ptr

	snapshots []VMSnapshot

	Cabinet *memory.Cabinet
	Drawer  *memory.Drawer

	openUpvalues map[int]memory.Ptr
}

func New(bytecode *compiler.Bytecode) *VM {
	if True == nil {
		True = object.NewBoolean(true)
		False = object.NewBoolean(false)
		Null = object.NewNull()
	}

	ptr, err := memory.AllocCompiledFunction(bytecode.Instructions, 0, 0)
	if err != nil {
		panic(fmt.Sprintf("vm boot error: %v", err))
	}
	mainFn := &object.CompiledFunction{Address: ptr}

	closurePtr, err := memory.AllocClosure(mainFn.Address, []memory.Ptr{})
	if err != nil { panic(err) }
	mainClosure := &object.Closure{Address: closurePtr}

	mainFrame := NewFrame(mainClosure, 0)

	frames := make([]*Frame, MaxFrames)
	frames[0] = mainFrame

	if len(memory.Lemari.Drawers) == 0 {
		memory.InitCabinet()
		// Re-allocate global constants as RAM was wiped
		True = object.NewBoolean(true)
		False = object.NewBoolean(false)
		Null = object.NewNull()
		TruePtr = True.Address
		FalsePtr = False.Address
		NullPtr = Null.Address
	}

	bootstrapOnce.Do(func() {
		if TruePtr == 0 {
			TruePtr = True.Address
		}
		if FalsePtr == 0 {
			FalsePtr = False.Address
		}
		if NullPtr == 0 {
			NullPtr = Null.Address
		}
		memory.Lemari.RootProvider = GlobalRootProvider
		memory.Lemari.GCTrigger = TriggerGC
	})

	schedulerOnce.Do(func() {
		scheduler.Init(4, executeTask)
	})

	drawer := &memory.Lemari.Drawers[0]

	vm := &VM{
		constants:   bytecode.Constants,
		globals:     make([]memory.Ptr, GlobalSize),
		stack:       [StackSize]memory.Ptr{},
		sp:          0,
		frames:      frames,
		framesIndex: 1,
		snapshots:   make([]VMSnapshot, 0),
		Cabinet:     &memory.Lemari,
		Drawer:      drawer,
		openUpvalues: make(map[int]memory.Ptr),
	}
	activeVMs.Store(vm, true)
	return vm
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
		fmt.Printf("Function Locals: %d\n", frame.cl.Fn().NumLocals())
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

	GlobalVMLock.RLock()
	defer GlobalVMLock.RUnlock()

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
			top := vm.stack[vm.sp-1]
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
			vm.closeUpvalues(frame.basePointer)
			if vm.framesIndex == 0 {
				vm.sp = 0
			} else {
				vm.sp = frame.basePointer - 1
			}
			if err := vm.push(returnValue); err != nil { return err }
			if vm.framesIndex == 0 { return nil }

		case compiler.OpReturn:
			frame := vm.popFrame()
			vm.closeUpvalues(frame.basePointer)
			if vm.framesIndex == 0 {
				vm.sp = 0
			} else {
				vm.sp = frame.basePointer - 1
			}
			if err := vm.push(NullPtr); err != nil { return err }
			if vm.framesIndex == 0 { return nil }

		case compiler.OpClosure:
			constIndex := compiler.ReadUint16(ins[ip+1:])
			numFree := int(ins[ip+3])
			vm.currentFrame().ip += 3
			if err := vm.pushClosure(int(constIndex), numFree); err != nil { return err }

		case compiler.OpGetFree:
			freeIndex := int(ins[ip+1])
			vm.currentFrame().ip += 1

			_, freePtrs, err := memory.ReadClosure(vm.currentFrame().cl.Address)
			if err != nil { return err }
			upvaluePtr := freePtrs[freeIndex]

			valPtr, stackIdx, isOpen, err := memory.ReadUpvalue(upvaluePtr)
			if err != nil { return err }

			if isOpen {
				valPtr = vm.stack[stackIdx]
			}
			if err := vm.push(valPtr); err != nil { return err }

		case compiler.OpSetFree:
			freeIndex := int(ins[ip+1])
			vm.currentFrame().ip += 1
			val, err := vm.pop()
			if err != nil { return err }

			_, freePtrs, err := memory.ReadClosure(vm.currentFrame().cl.Address)
			if err != nil { return err }
			upvaluePtr := freePtrs[freeIndex]

			_, stackIdx, isOpen, err := memory.ReadUpvalue(upvaluePtr)
			if err != nil { return err }

			if isOpen {
				vm.stack[stackIdx] = val
			} else {
				if err := memory.CloseUpvalue(upvaluePtr, val); err != nil { return err }
			}

		case compiler.OpCaptureLocal:
			localIndex := int(ins[ip+1])
			vm.currentFrame().ip += 1
			frame := vm.currentFrame()
			absIndex := frame.basePointer + localIndex

			upvaluePtr, ok := vm.openUpvalues[absIndex]
			if !ok {
				var err error
				upvaluePtr, err = memory.AllocUpvalue(absIndex)
				if err != nil { return err }
				vm.openUpvalues[absIndex] = upvaluePtr
			}
			if err := vm.push(upvaluePtr); err != nil { return err }

		case compiler.OpLoadUpvalue:
			freeIndex := int(ins[ip+1])
			vm.currentFrame().ip += 1

			_, freePtrs, err := memory.ReadClosure(vm.currentFrame().cl.Address)
			if err != nil { return err }
			upvaluePtr := freePtrs[freeIndex]
			if err := vm.push(upvaluePtr); err != nil { return err }

		case compiler.OpStruct:
			nameIndex := compiler.ReadUint16(ins[ip+1:])
			fieldCount := int(compiler.ReadUint16(ins[ip+3:]))
			vm.currentFrame().ip += 4

			nameObj := vm.constants[nameIndex]
			namePtr := getObjectAddress(nameObj)

			fieldsPtr, err := vm.buildArray(vm.sp-fieldCount, vm.sp)
			if err != nil { return err }
			vm.sp -= fieldCount

			ptr, err := memory.AllocSchema(namePtr, fieldsPtr)
			if err != nil { return err }
			if err := vm.push(ptr); err != nil { return err }

		case compiler.OpArray:
			numElements := int(compiler.ReadUint16(ins[ip+1:]))
			vm.currentFrame().ip += 2
			ptr, err := vm.buildArray(vm.sp-numElements, vm.sp)
			if err != nil { return err }
			vm.sp -= numElements
			if err := vm.push(ptr); err != nil { return err }

		case compiler.OpHash:
			numElements := int(compiler.ReadUint16(ins[ip+1:]))
			vm.currentFrame().ip += 2
			ptr, err := vm.buildHash(vm.sp-numElements, vm.sp)
			if err != nil { return err }
			vm.sp -= numElements
			if err := vm.push(ptr); err != nil { return err }

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

		case compiler.OpUpdateModule:
			result, err := vm.pop()
			if err != nil { return err }
			modPtr, err := vm.pop()
			if err != nil { return err }

			if err := memory.WriteModuleExports(modPtr, result); err != nil { return err }
			if err := vm.push(result); err != nil { return err }

		case compiler.OpAnd, compiler.OpOr, compiler.OpXor, compiler.OpLShift, compiler.OpRShift:
			if err := vm.executeBitwiseOperation(op); err != nil { return err }

		case compiler.OpJump:
			pos := int(compiler.ReadUint16(ins[ip+1:]))
			vm.currentFrame().ip = pos - 1
			GlobalVMLock.RUnlock()
			GlobalVMLock.RLock()

		case compiler.OpJumpNotTruthy:
			pos := int(compiler.ReadUint16(ins[ip+1:]))
			vm.currentFrame().ip += 2
			condition, err := vm.pop()
			if err != nil { return err }
			if !isTruthy(condition) {
				vm.currentFrame().ip = pos - 1
			}

		default:
			return fmt.Errorf("unknown opcode %d", op)
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
	vm.LastPoppedPtr = ptr
	return ptr, nil
}

func (vm *VM) GetLastPopped() object.Object {
	obj, _ := Rehydrate(vm.LastPoppedPtr)
	return obj
}

func (vm *VM) pushClosure(constIndex int, numFree int) error {
	constant := vm.constants[constIndex]
	fn := constant.(*object.CompiledFunction)

	freeVars := make([]memory.Ptr, numFree)
	for i := 0; i < numFree; i++ {
		ptr := vm.stack[vm.sp-numFree+i]
		freeVars[i] = ptr
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
		clWrapper := &object.Closure{Address: calleePtr}
		fnWrapper := clWrapper.Fn()

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

	if header.Type == memory.TagModule {
		if numArgs != 0 {
			return fmt.Errorf("module import takes 0 args")
		}
		return vm.executeModuleCall(calleePtr)
	}

	if header.Type == memory.TagSchema {
		return vm.executeSchemaCall(calleePtr, numArgs)
	}

	return fmt.Errorf("calling non-function")
}

func (vm *VM) executeSchemaCall(schemaPtr memory.Ptr, numArgs int) error {
	_, fieldsPtr, err := memory.ReadSchema(schemaPtr)
	if err != nil { return err }

	length, err := memory.ReadArrayLength(fieldsPtr)
	if err != nil { return err }

	if numArgs != length {
		return fmt.Errorf("struct init arg mismatch: want %d, got %d", length, numArgs)
	}

	structPtr, err := memory.AllocStruct(schemaPtr, numArgs)
	if err != nil { return err }

	for i := numArgs - 1; i >= 0; i-- {
		valPtr := vm.stack[vm.sp-1]
		vm.sp--
		memory.WriteStructField(structPtr, i, valPtr)
	}
	vm.sp-- // Pop Schema
	return vm.push(structPtr)
}

func (vm *VM) executeModuleCall(modPtr memory.Ptr) error {
	initPtr, expPtr, err := memory.ReadModule(modPtr)
	if err != nil { return err }

	// 1. Check if Loaded
	if expPtr != memory.NilPtr {
		// Loaded. Replace Module on stack with Exports.
		vm.sp-- // Pop Module
		return vm.push(expPtr)
	}

	// 2. Mark Loading (using Null as placeholder)
	memory.WriteModuleExports(modPtr, NullPtr)

	// 3. Prepare Trampoline
	// Bytecode: OpLoadLocal(0), 0, OpLoadLocal(0), 1, OpCall, 0, OpUpdateModule, OpReturnValue
	// Hex: 13 00 13 01 40 00 50 42
	instr := []byte{
		byte(compiler.OpLoadLocal), 0,
		byte(compiler.OpLoadLocal), 1,
		byte(compiler.OpCall), 0,
		byte(compiler.OpUpdateModule),
		byte(compiler.OpReturnValue),
	}

	trampPtr, err := memory.AllocCompiledFunction(instr, 0, 2) // 2 params
	if err != nil { return err }

	clPtr, err := memory.AllocClosure(trampPtr, nil)
	if err != nil { return err }

	// Replace Module with Trampoline Closure
	vm.stack[vm.sp-1] = clPtr

	// Wrap InitFn in Closure
	initClPtr, err := memory.AllocClosure(initPtr, nil)
	if err != nil { return err }

	// Push Args (Module, InitClosure)
	if err := vm.push(modPtr); err != nil { return err }
	if err := vm.push(initClPtr); err != nil { return err }

	// Execute Trampoline (2 args)
	return vm.executeCall(2)
}

func (vm *VM) executeBuiltinCall(builtinPtr memory.Ptr, numArgs int) error {
	builtinObj, _ := Rehydrate(builtinPtr)
	builtin := builtinObj.(*object.Builtin)

	args := make([]object.Object, numArgs)
	for i := 0; i < numArgs; i++ {
		ptr := vm.stack[vm.sp-numArgs+i]
		args[i], _ = Rehydrate(ptr)
	}

	res := builtin.Fn(args...)

	if errObj, ok := res.(*object.Error); ok {
		if errObj.GetMessage() == "luncurkan() requires VM context" {
			tObj, err := vm.spawn(args)
			if err != nil {
				res = object.NewError(err.Error(), "", 0, 0)
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

	return object.NewThread(resultCh), nil
}

func executeTask(ptr memory.Ptr) {
	GlobalVMLock.RLock()
	id, _ := memory.ReadInteger(ptr)
	GlobalVMLock.RUnlock()

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
		sp: ctx.Closure.Fn().NumLocals(),
		frames: frames,
		framesIndex: 1,
		Cabinet: &memory.Lemari,
	}
	activeVMs.Store(newVM, true)
	defer activeVMs.Delete(newVM)

	err := newVM.Run()
	if err != nil {
		ctx.ResultCh <- object.NewError(err.Error(), "", 0, 0)
	} else {
		val := newVM.StackTop()
		if val != nil {
			ctx.ResultCh <- val
		} else {
			ctx.ResultCh <- Null
		}
	}
	close(ctx.ResultCh)
}

func (vm *VM) snapshot() error {
	return nil
}

func (vm *VM) rollback() error { return nil }
func (vm *VM) commit() error { return nil }

func ensureOnHeap(obj object.Object) error {
	if obj.GetAddress() == memory.NilPtr {
		return fmt.Errorf("ensureOnHeap: object %s has nil address", obj.Type())
	}
	return nil
}

func getObjectAddress(obj object.Object) memory.Ptr {
	return obj.GetAddress()
}

func TriggerGC() {
	// Startup Safety Guard:
	// We attempt to release the RLock. If this panics, it means the caller
	// did NOT hold the RLock (e.g., inside New() or unmanaged thread).
	// In that case, we are in an unsafe state to run GC (roots might be missing).
	// So we abort the GC attempt and let Alloc fail with ErrOOM.
	safeToRun := false
	func() {
		defer func() {
			if r := recover(); r != nil {
				safeToRun = false
			}
		}()
		GlobalVMLock.RUnlock()
		safeToRun = true
	}()

	if !safeToRun {
		return
	}

	GlobalVMLock.Lock()
	// Ensure we restore the lock state for the caller
	defer func() {
		GlobalVMLock.Unlock()
		GlobalVMLock.RLock()
	}()

	roots := GlobalRootProvider()
	err := memory.Lemari.MarkAndCompact(roots)
	if err != nil {
		fmt.Printf("GC Error: %v\n", err)
	}
}

func GlobalRootProvider() []*memory.Ptr {
	roots := []*memory.Ptr{}

	// Global constants
	if TruePtr != memory.NilPtr { roots = append(roots, &TruePtr) }
	if FalsePtr != memory.NilPtr { roots = append(roots, &FalsePtr) }
	if NullPtr != memory.NilPtr { roots = append(roots, &NullPtr) }

	// Scheduler Roots
	if scheduler.Global != nil && scheduler.Global.Queue != nil {
		roots = append(roots, &scheduler.Global.Queue.Address)
	}

	// Task Registry Roots
	taskRegistry.Range(func(key, value interface{}) bool {
		ctx := value.(TaskContext)
		if ctx.Closure != nil {
			roots = append(roots, &ctx.Closure.Address)
		}
		for i := range ctx.Globals {
			if ctx.Globals[i] != memory.NilPtr {
				roots = append(roots, &ctx.Globals[i])
			}
		}
		for _, obj := range ctx.Constants {
			switch val := obj.(type) {
			case *object.Integer: roots = append(roots, &val.Address)
			case *object.Float: roots = append(roots, &val.Address)
			case *object.Boolean: roots = append(roots, &val.Address)
			case *object.String: roots = append(roots, &val.Address)
			case *object.CompiledFunction: roots = append(roots, &val.Address)
			case *object.Null: roots = append(roots, &val.Address)
			case *object.Module:
				roots = append(roots, &val.Address)
			}
		}
		return true
	})

	activeVMs.Range(func(key, value interface{}) bool {
		vm := key.(*VM)
		roots = append(roots, vm.GetRoots()...)
		return true
	})

	return roots
}

func (vm *VM) GetRoots() []*memory.Ptr {
	roots := []*memory.Ptr{}

	// 1. Stack (Live)
	for i := 0; i < vm.sp; i++ {
		roots = append(roots, &vm.stack[i])
	}

	// 2. Globals
	for i := range vm.globals {
		if vm.globals[i] != memory.NilPtr {
			roots = append(roots, &vm.globals[i])
		}
	}

	// 3. Constants
	for _, obj := range vm.constants {
		switch val := obj.(type) {
		case *object.Integer: roots = append(roots, &val.Address)
		case *object.Float: roots = append(roots, &val.Address)
		case *object.Boolean: roots = append(roots, &val.Address)
		case *object.String: roots = append(roots, &val.Address)
		case *object.CompiledFunction: roots = append(roots, &val.Address)
		case *object.Null: roots = append(roots, &val.Address)
		case *object.Module: roots = append(roots, &val.Address)
		// Add others if needed
		}
	}

	// 4. Frames (Closures)
	for i := 0; i < vm.framesIndex; i++ {
		f := vm.frames[i]
		if f != nil && f.cl != nil {
			roots = append(roots, &f.cl.Address)
		}
	}

	return roots
}

func (vm *VM) closeUpvalues(limit int) {
	for idx, uvPtr := range vm.openUpvalues {
		if idx >= limit {
			val := vm.stack[idx]
			if err := memory.CloseUpvalue(uvPtr, val); err != nil {
				panic(err)
			}
			delete(vm.openUpvalues, idx)
		}
	}
}
