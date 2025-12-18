package vm

import (
	"fmt"
	"github.com/VzoelFox/morphlang/pkg/compiler"
	"github.com/VzoelFox/morphlang/pkg/memory"
)

func (vm *VM) executeBinaryOperation(op compiler.Opcode) error {
	right, err := vm.pop()
	if err != nil { return err }
	left, err := vm.pop()
	if err != nil { return err }

	leftHeader, err := memory.ReadHeader(left)
	if err != nil { return err }

	if leftHeader.Type == memory.TagInteger {
		leftVal, _ := memory.ReadInteger(left)
		rightVal, _ := memory.ReadInteger(right)

		var res int64
		switch op {
		case compiler.OpAdd: res = leftVal + rightVal
		case compiler.OpSub: res = leftVal - rightVal
		case compiler.OpMul: res = leftVal * rightVal
		case compiler.OpDiv: res = leftVal / rightVal
		}
		ptr, err := memory.AllocInteger(res)
		if err != nil { return err }
		return vm.push(ptr)
	}

	if leftHeader.Type == memory.TagString {
		leftVal, _ := memory.ReadString(left)
		rightVal, _ := memory.ReadString(right)
		if op != compiler.OpAdd { return fmt.Errorf("string only supports add") }
		ptr, err := memory.AllocString(leftVal + rightVal)
		if err != nil { return err }
		return vm.push(ptr)
	}

	return fmt.Errorf("unsupported binary op")
}

func (vm *VM) executeComparison(op compiler.Opcode) error {
	right, err := vm.pop()
	if err != nil { return err }
	left, err := vm.pop()
	if err != nil { return err }

	leftHeader, err := memory.ReadHeader(left)
	if err != nil { return err }

	if leftHeader.Type == memory.TagInteger {
		leftVal, _ := memory.ReadInteger(left)
		rightVal, _ := memory.ReadInteger(right)

		var val bool
		switch op {
		case compiler.OpEqual: val = leftVal == rightVal
		case compiler.OpNotEqual: val = leftVal != rightVal
		case compiler.OpGreaterThan: val = leftVal > rightVal
		case compiler.OpGreaterEqual: val = leftVal >= rightVal
		}

		ptr, err := memory.AllocBoolean(val)
		if err != nil { return err }
		return vm.push(ptr)
	}
	return fmt.Errorf("unsupported comparison")
}

func (vm *VM) executeBitwiseOperation(op compiler.Opcode) error {
	return fmt.Errorf("bitwise not implemented in phase x.4")
}

func (vm *VM) executeBangOperator() error {
	operand, err := vm.pop()
	if err != nil { return err }
	// Read Boolean
	val, _ := memory.ReadBoolean(operand)
	ptr, _ := memory.AllocBoolean(!val)
	return vm.push(ptr)
}

func (vm *VM) executeMinusOperator() error {
	return fmt.Errorf("minus not implemented")
}

func (vm *VM) executeBitNotOperator() error {
	return fmt.Errorf("bitnot not implemented")
}

func (vm *VM) buildArray(startIndex, endIndex int) error {
	length := endIndex - startIndex
	ptr, err := memory.AllocArray(length, length)
	if err != nil { return err }

	for i := 0; i < length; i++ {
		elemPtr := vm.stack[startIndex+i]
		memory.WriteArrayElement(ptr, i, elemPtr)
	}
	return vm.push(ptr)
}

func (vm *VM) buildHash(startIndex, endIndex int) error {
	count := (endIndex - startIndex) / 2
	ptr, err := memory.AllocHash(count)
	if err != nil { return err }

	for i := 0; i < count; i++ {
		key := vm.stack[startIndex + i*2]
		val := vm.stack[startIndex + i*2 + 1]
		memory.WriteHashPair(ptr, i, key, val)
	}
	return vm.push(ptr)
}

func (vm *VM) executeIndexExpression(left, index memory.Ptr) error {
	return fmt.Errorf("index not implemented")
}

func (vm *VM) executeSetIndexExpression(left, index, val memory.Ptr) error {
	return fmt.Errorf("set index not implemented")
}
