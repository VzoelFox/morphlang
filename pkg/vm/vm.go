package vm

import (
	"fmt"

	"github.com/VzoelFox/morphlang/pkg/compiler"
	"github.com/VzoelFox/morphlang/pkg/object"
)

const StackSize = 2048
const GlobalSize = 65536
const MaxFrames = 1024

var (
	True  = &object.Boolean{Value: true}
	False = &object.Boolean{Value: false}
	Null  = &object.Null{}
)

type VM struct {
	constants []object.Object
	globals   []object.Object

	stack [StackSize]object.Object
	sp    int // Always points to the next value

	frames      []*Frame
	framesIndex int

	LastPoppedStackElem object.Object
}

func New(bytecode *compiler.Bytecode) *VM {
	mainFn := &object.CompiledFunction{Instructions: bytecode.Instructions}
	mainClosure := &object.Closure{Fn: mainFn}
	mainFrame := NewFrame(mainClosure, 0)

	frames := make([]*Frame, MaxFrames)
	frames[0] = mainFrame

	return &VM{
		constants:   bytecode.Constants,
		globals:     make([]object.Object, GlobalSize),
		stack:       [StackSize]object.Object{},
		sp:          0,
		frames:      frames,
		framesIndex: 1,
	}
}

func (vm *VM) currentFrame() *Frame {
	return vm.frames[vm.framesIndex-1]
}

func (vm *VM) pushFrame(f *Frame) {
	vm.frames[vm.framesIndex] = f
	vm.framesIndex++
}

func (vm *VM) popFrame() *Frame {
	vm.framesIndex--
	return vm.frames[vm.framesIndex]
}

func (vm *VM) StackTop() object.Object {
	if vm.sp == 0 {
		return nil
	}
	return vm.stack[vm.sp-1]
}

func (vm *VM) Run() error {
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
			err := vm.push(vm.constants[constIndex])
			if err != nil {
				return err
			}

		case compiler.OpAdd, compiler.OpSub, compiler.OpMul, compiler.OpDiv:
			err := vm.executeBinaryOperation(op)
			if err != nil {
				return err
			}

		case compiler.OpEqual, compiler.OpNotEqual, compiler.OpGreaterThan, compiler.OpGreaterEqual:
			err := vm.executeComparison(op)
			if err != nil {
				return err
			}

		case compiler.OpBang:
			err := vm.executeBangOperator()
			if err != nil {
				return err
			}

		case compiler.OpMinus:
			err := vm.executeMinusOperator()
			if err != nil {
				return err
			}

		case compiler.OpPop:
			vm.pop()

		case compiler.OpDup:
			top := vm.StackTop()
			if top == nil {
				return fmt.Errorf("stack underflow on dup")
			}
			err := vm.push(top)
			if err != nil {
				return err
			}

		case compiler.OpJump:
			pos := int(compiler.ReadUint16(ins[ip+1:]))
			vm.currentFrame().ip = pos - 1

		case compiler.OpJumpNotTruthy:
			pos := int(compiler.ReadUint16(ins[ip+1:]))
			vm.currentFrame().ip += 2

			condition := vm.pop()
			if !isTruthy(condition) {
				vm.currentFrame().ip = pos - 1
			}

		case compiler.OpStoreGlobal:
			globalIndex := compiler.ReadUint16(ins[ip+1:])
			vm.currentFrame().ip += 2
			vm.globals[globalIndex] = vm.pop()

		case compiler.OpLoadGlobal:
			globalIndex := compiler.ReadUint16(ins[ip+1:])
			vm.currentFrame().ip += 2
			err := vm.push(vm.globals[globalIndex])
			if err != nil {
				return err
			}

		case compiler.OpStoreLocal:
			localIndex := int(ins[ip+1])
			vm.currentFrame().ip += 1
			frame := vm.currentFrame()
			vm.stack[frame.basePointer+localIndex] = vm.pop()

		case compiler.OpLoadLocal:
			localIndex := int(ins[ip+1])
			vm.currentFrame().ip += 1
			frame := vm.currentFrame()
			err := vm.push(vm.stack[frame.basePointer+localIndex])
			if err != nil {
				return err
			}

		case compiler.OpGetBuiltin:
			builtinIndex := int(ins[ip+1])
			vm.currentFrame().ip += 1

			definition := object.Builtins[builtinIndex]
			err := vm.push(definition.Builtin)
			if err != nil {
				return err
			}

		case compiler.OpArray:
			numElements := int(compiler.ReadUint16(ins[ip+1:]))
			vm.currentFrame().ip += 2

			array := vm.buildArray(vm.sp-numElements, vm.sp)
			vm.sp = vm.sp - numElements

			err := vm.push(array)
			if err != nil {
				return err
			}

		case compiler.OpHash:
			numElements := int(compiler.ReadUint16(ins[ip+1:]))
			vm.currentFrame().ip += 2

			hash, err := vm.buildHash(vm.sp-numElements, vm.sp)
			if err != nil {
				return err
			}
			vm.sp = vm.sp - numElements

			err = vm.push(hash)
			if err != nil {
				return err
			}

		case compiler.OpIndex:
			index := vm.pop()
			left := vm.pop()

			err := vm.executeIndexExpression(left, index)
			if err != nil {
				return err
			}

		case compiler.OpSetIndex:
			val := vm.pop()
			index := vm.pop()
			left := vm.pop()

			err := vm.executeSetIndexExpression(left, index, val)
			if err != nil {
				return err
			}

		case compiler.OpCall:
			numArgs := int(ins[ip+1])
			vm.currentFrame().ip += 1

			err := vm.executeCall(numArgs)
			if err != nil {
				return err
			}

		case compiler.OpReturnValue:
			returnValue := vm.pop()

			if vm.framesIndex == 1 {
				vm.popFrame()
				return nil
			}

			frame := vm.popFrame()
			vm.sp = frame.basePointer - 1 // Pop locals AND function
			if vm.sp < 0 {
				return fmt.Errorf("stack underflow on return")
			}

			err := vm.push(returnValue)
			if err != nil {
				return err
			}

		case compiler.OpReturn:
			if vm.framesIndex == 1 {
				vm.popFrame()
				return nil
			}

			frame := vm.popFrame()
			vm.sp = frame.basePointer - 1
			if vm.sp < 0 {
				return fmt.Errorf("stack underflow on return")
			}

			err := vm.push(Null)
			if err != nil {
				return err
			}

		case compiler.OpClosure:
			constIndex := compiler.ReadUint16(ins[ip+1:])
			numFree := int(ins[ip+3])
			vm.currentFrame().ip += 3

			err := vm.pushClosure(int(constIndex), numFree)
			if err != nil {
				return err
			}

		case compiler.OpGetFree:
			freeIndex := int(ins[ip+1])
			vm.currentFrame().ip += 1

			currentClosure := vm.currentFrame().cl
			err := vm.push(currentClosure.FreeVariables[freeIndex])
			if err != nil {
				return err
			}

		}
	}

	return nil
}

func (vm *VM) pushClosure(constIndex int, numFree int) error {
	constant := vm.constants[constIndex]
	function, ok := constant.(*object.CompiledFunction)
	if !ok {
		return fmt.Errorf("not a function: %+v", constant)
	}

	free := make([]object.Object, numFree)
	for i := 0; i < numFree; i++ {
		free[i] = vm.stack[vm.sp-numFree+i]
	}
	vm.sp = vm.sp - numFree

	closure := &object.Closure{Fn: function, FreeVariables: free}
	return vm.push(closure)
}

func (vm *VM) executeCall(numArgs int) error {
	callee := vm.stack[vm.sp-1-numArgs]
	switch callee := callee.(type) {
	case *object.Closure:
		if numArgs != callee.Fn.NumParameters {
			return fmt.Errorf("wrong number of arguments: want=%d, got=%d",
				callee.Fn.NumParameters, numArgs)
		}

		frame := NewFrame(callee, vm.sp-numArgs)
		vm.pushFrame(frame)
		vm.sp = frame.basePointer + callee.Fn.NumLocals
		return nil

	case *object.Builtin:
		return vm.executeBuiltinCall(callee, numArgs)

	default:
		return fmt.Errorf("calling non-function")
	}
}

func (vm *VM) executeBuiltinCall(builtin *object.Builtin, numArgs int) error {
	args := vm.stack[vm.sp-numArgs : vm.sp]

	result := builtin.Fn(args...)

	// INTERCEPT: luncurkan
	if errObj, ok := result.(*object.Error); ok && errObj.Message == "luncurkan() requires VM context" {
		threadObj, err := vm.spawn(args)
		if err != nil {
			result = &object.Error{Message: err.Error()}
		} else {
			result = threadObj
		}
	}

	vm.sp = vm.sp - numArgs - 1 // Pop args + function

	if result != nil {
		return vm.push(result)
	} else {
		return vm.push(Null)
	}
}

func (vm *VM) spawn(args []object.Object) (*object.Thread, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("luncurkan takes exactly 1 argument")
	}
	cl, ok := args[0].(*object.Closure)
	if !ok {
		return nil, fmt.Errorf("luncurkan argument must be a function")
	}
	if cl.Fn.NumParameters != 0 {
		return nil, fmt.Errorf("luncurkan function must accept 0 arguments")
	}

	// Create new VM with ISOLATED globals (Copy-on-Spawn)
	// This ensures that variable modifications in the thread do not affect the parent.
	frames := make([]*Frame, MaxFrames)
	frames[0] = NewFrame(cl, 0)

	newGlobals := make([]object.Object, len(vm.globals))
	copy(newGlobals, vm.globals)

	newVM := &VM{
		constants:   vm.constants,
		globals:     newGlobals,
		stack:       [StackSize]object.Object{},
		sp:          cl.Fn.NumLocals, // Reserve space for locals on the stack
		frames:      frames,
		framesIndex: 1,
	}

	resultCh := make(chan object.Object, 1)

	go func() {
		err := newVM.Run()
		if err != nil {
			// If run fails, send error
			resultCh <- &object.Error{Message: fmt.Sprintf("Background task error: %v", err)}
		} else {
			// If run succeeds, send last popped element (return value) or Null
			if newVM.LastPoppedStackElem != nil {
				resultCh <- newVM.LastPoppedStackElem
			} else {
				resultCh <- Null
			}
		}
		close(resultCh)
	}()

	return &object.Thread{Result: resultCh}, nil
}

func (vm *VM) push(o object.Object) error {
	if vm.sp >= StackSize {
		return fmt.Errorf("stack overflow")
	}

	vm.stack[vm.sp] = o
	vm.sp++

	return nil
}

func (vm *VM) pop() object.Object {
	o := vm.stack[vm.sp-1]
	vm.sp--
	vm.LastPoppedStackElem = o
	return o
}

func (vm *VM) executeBinaryOperation(op compiler.Opcode) error {
	right := vm.pop()
	left := vm.pop()

	if left == nil {
		return fmt.Errorf("executeBinaryOperation: left operand is nil")
	}
	if right == nil {
		return fmt.Errorf("executeBinaryOperation: right operand is nil")
	}

	leftType := left.Type()
	rightType := right.Type()

	if leftType == object.INTEGER_OBJ && rightType == object.INTEGER_OBJ {
		return vm.executeBinaryIntegerOperation(op, left, right)
	}

	if leftType == object.STRING_OBJ {
		return vm.executeBinaryStringOperation(op, left, right)
	}

	if leftType == object.ARRAY_OBJ && rightType == object.ARRAY_OBJ {
		return vm.executeBinaryArrayOperation(op, left, right)
	}

	return vm.push(&object.Error{Message: fmt.Sprintf("unsupported types for binary operation: %s %s", leftType, rightType)})
}

func (vm *VM) executeBinaryArrayOperation(op compiler.Opcode, left, right object.Object) error {
	if op != compiler.OpAdd {
		return vm.push(&object.Error{Message: fmt.Sprintf("unknown array operator: %d", op)})
	}

	leftVal := left.(*object.Array).Elements
	rightVal := right.(*object.Array).Elements

	newElements := make([]object.Object, len(leftVal)+len(rightVal))
	copy(newElements, leftVal)
	copy(newElements[len(leftVal):], rightVal)

	return vm.push(&object.Array{Elements: newElements})
}

func (vm *VM) executeBinaryIntegerOperation(op compiler.Opcode, left, right object.Object) error {
	leftVal := left.(*object.Integer).Value
	rightVal := right.(*object.Integer).Value

	var result int64

	switch op {
	case compiler.OpAdd:
		result = leftVal + rightVal
	case compiler.OpSub:
		result = leftVal - rightVal
	case compiler.OpMul:
		result = leftVal * rightVal
	case compiler.OpDiv:
		if rightVal == 0 {
			return vm.push(&object.Error{Message: "division by zero"})
		}
		result = leftVal / rightVal
	default:
		return vm.push(&object.Error{Message: fmt.Sprintf("unknown integer operator: %d", op)})
	}

	return vm.push(&object.Integer{Value: result})
}

func (vm *VM) executeBinaryStringOperation(op compiler.Opcode, left, right object.Object) error {
	if op != compiler.OpAdd {
		return vm.push(&object.Error{Message: fmt.Sprintf("unknown string operator: %d", op)})
	}

	leftVal := left.(*object.String).Value
	var rightVal string

	if rightStr, ok := right.(*object.String); ok {
		rightVal = rightStr.Value
	} else {
		rightVal = right.Inspect()
	}

	return vm.push(&object.String{Value: leftVal + rightVal})
}

func (vm *VM) executeComparison(op compiler.Opcode) error {
	right := vm.pop()
	left := vm.pop()

	if left.Type() == object.INTEGER_OBJ && right.Type() == object.INTEGER_OBJ {
		return vm.executeIntegerComparison(op, left, right)
	}

	if left.Type() == object.STRING_OBJ && right.Type() == object.STRING_OBJ {
		return vm.executeStringComparison(op, left, right)
	}

	switch op {
	case compiler.OpEqual:
		return vm.push(nativeBoolToBooleanObject(left == right || (left.Type() == object.BOOLEAN_OBJ && right.Type() == object.BOOLEAN_OBJ && left.(*object.Boolean).Value == right.(*object.Boolean).Value)))
	case compiler.OpNotEqual:
		return vm.push(nativeBoolToBooleanObject(left != right && !(left.Type() == object.BOOLEAN_OBJ && right.Type() == object.BOOLEAN_OBJ && left.(*object.Boolean).Value == right.(*object.Boolean).Value)))
	default:
		return vm.push(&object.Error{Message: fmt.Sprintf("unsupported comparison: %s %d %s", left.Type(), op, right.Type())})
	}
}

func (vm *VM) executeIntegerComparison(op compiler.Opcode, left, right object.Object) error {
	leftVal := left.(*object.Integer).Value
	rightVal := right.(*object.Integer).Value

	switch op {
	case compiler.OpEqual:
		return vm.push(nativeBoolToBooleanObject(leftVal == rightVal))
	case compiler.OpNotEqual:
		return vm.push(nativeBoolToBooleanObject(leftVal != rightVal))
	case compiler.OpGreaterThan:
		return vm.push(nativeBoolToBooleanObject(leftVal > rightVal))
	case compiler.OpGreaterEqual:
		return vm.push(nativeBoolToBooleanObject(leftVal >= rightVal))
	default:
		return fmt.Errorf("unknown integer operator: %d", op)
	}
}

func (vm *VM) executeStringComparison(op compiler.Opcode, left, right object.Object) error {
	leftVal := left.(*object.String).Value
	rightVal := right.(*object.String).Value

	switch op {
	case compiler.OpEqual:
		return vm.push(nativeBoolToBooleanObject(leftVal == rightVal))
	case compiler.OpNotEqual:
		return vm.push(nativeBoolToBooleanObject(leftVal != rightVal))
	default:
		return fmt.Errorf("unknown string operator: %d", op)
	}
}

func (vm *VM) executeBangOperator() error {
	operand := vm.pop()

	if isTruthy(operand) {
		return vm.push(False)
	}
	return vm.push(True)
}

func (vm *VM) executeMinusOperator() error {
	operand := vm.pop()

	if operand.Type() != object.INTEGER_OBJ {
		return vm.push(&object.Error{Message: fmt.Sprintf("unsupported type for negation: %s", operand.Type())})
	}

	value := operand.(*object.Integer).Value
	return vm.push(&object.Integer{Value: -value})
}

func (vm *VM) buildArray(startIndex, endIndex int) object.Object {
	elements := make([]object.Object, endIndex-startIndex)

	for i := 0; i < len(elements); i++ {
		elements[i] = vm.stack[startIndex+i]
	}

	return &object.Array{Elements: elements}
}

func (vm *VM) buildHash(startIndex, endIndex int) (object.Object, error) {
	hashedPairs := make(map[object.HashKey]object.HashPair)

	for i := startIndex; i < endIndex; i += 2 {
		key := vm.stack[i]
		value := vm.stack[i+1]

		pair := object.HashPair{Key: key, Value: value}

		hashKey, ok := key.(object.Hashable)
		if !ok {
			return nil, fmt.Errorf("unusable as hash key: %s", key.Type())
		}

		hashedPairs[hashKey.HashKey()] = pair
	}

	return &object.Hash{Pairs: hashedPairs}, nil
}

func (vm *VM) executeIndexExpression(left, index object.Object) error {
	switch {
	case left.Type() == object.ARRAY_OBJ && index.Type() == object.INTEGER_OBJ:
		return vm.executeArrayIndex(left, index)
	case left.Type() == object.HASH_OBJ:
		return vm.executeHashIndex(left, index)
	default:
		return fmt.Errorf("index operator not supported: %s", left.Type())
	}
}

func (vm *VM) executeSetIndexExpression(left, index, val object.Object) error {
	switch {
	case left.Type() == object.ARRAY_OBJ && index.Type() == object.INTEGER_OBJ:
		return vm.executeArraySetIndex(left, index, val)
	case left.Type() == object.HASH_OBJ:
		return vm.executeHashSetIndex(left, index, val)
	default:
		return fmt.Errorf("index assignment not supported: %s", left.Type())
	}
}

func (vm *VM) executeArraySetIndex(array, index, val object.Object) error {
	arrayObject := array.(*object.Array)
	i := index.(*object.Integer).Value
	max := int64(len(arrayObject.Elements) - 1)

	if i < 0 || i > max {
		return fmt.Errorf("index out of bounds: %d", i)
	}

	arrayObject.Elements[i] = val
	return nil
}

func (vm *VM) executeHashSetIndex(hash, index, val object.Object) error {
	hashObject := hash.(*object.Hash)
	key, ok := index.(object.Hashable)
	if !ok {
		return fmt.Errorf("unusable as hash key: %s", index.Type())
	}

	pair := object.HashPair{Key: index, Value: val}
	hashObject.Pairs[key.HashKey()] = pair
	return nil
}

func (vm *VM) executeArrayIndex(array, index object.Object) error {
	arrayObject := array.(*object.Array)
	i := index.(*object.Integer).Value
	max := int64(len(arrayObject.Elements) - 1)

	if i < 0 || i > max {
		return vm.push(Null)
	}

	return vm.push(arrayObject.Elements[i])
}

func (vm *VM) executeHashIndex(hash, index object.Object) error {
	hashObject := hash.(*object.Hash)

	key, ok := index.(object.Hashable)
	if !ok {
		return fmt.Errorf("unusable as hash key: %s", index.Type())
	}

	pair, ok := hashObject.Pairs[key.HashKey()]
	if !ok {
		return vm.push(Null)
	}

	return vm.push(pair.Value)
}

func isTruthy(obj object.Object) bool {
	switch obj := obj.(type) {
	case *object.Boolean:
		return obj.Value
	case *object.Null:
		return false
	default:
		return true
	}
}

func nativeBoolToBooleanObject(input bool) *object.Boolean {
	if input {
		return True
	}
	return False
}
