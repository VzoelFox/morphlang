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
	rightHeader, err := memory.ReadHeader(right)
	if err != nil { return err }

	if leftHeader.Type == memory.TagInteger && rightHeader.Type == memory.TagInteger {
		leftVal, _ := memory.ReadInteger(left)
		rightVal, _ := memory.ReadInteger(right)

		var res int64
		switch op {
		case compiler.OpAdd: res = leftVal + rightVal
		case compiler.OpSub: res = leftVal - rightVal
		case compiler.OpMul: res = leftVal * rightVal
		case compiler.OpDiv:
			if rightVal == 0 { return fmt.Errorf("integer divide by zero") }
			res = leftVal / rightVal
		}
		ptr, err := memory.AllocInteger(res)
		if err != nil { return err }
		return vm.push(ptr)
	}

	if leftHeader.Type == memory.TagFloat && rightHeader.Type == memory.TagFloat {
		leftVal, _ := memory.ReadFloat(left)
		rightVal, _ := memory.ReadFloat(right)

		var res float64
		switch op {
		case compiler.OpAdd: res = leftVal + rightVal
		case compiler.OpSub: res = leftVal - rightVal
		case compiler.OpMul: res = leftVal * rightVal
		case compiler.OpDiv: res = leftVal / rightVal
		}
		ptr, err := memory.AllocFloat(res)
		if err != nil { return err }
		return vm.push(ptr)
	}

	if leftHeader.Type == memory.TagString && rightHeader.Type == memory.TagString {
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
	rightHeader, err := memory.ReadHeader(right)
	if err != nil { return err }

	if leftHeader.Type == memory.TagInteger && rightHeader.Type == memory.TagInteger {
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

	if leftHeader.Type == memory.TagFloat && rightHeader.Type == memory.TagFloat {
		leftVal, _ := memory.ReadFloat(left)
		rightVal, _ := memory.ReadFloat(right)

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

	if leftHeader.Type == memory.TagBoolean && rightHeader.Type == memory.TagBoolean {
		leftVal, _ := memory.ReadBoolean(left)
		rightVal, _ := memory.ReadBoolean(right)

		var val bool
		switch op {
		case compiler.OpEqual: val = leftVal == rightVal
		case compiler.OpNotEqual: val = leftVal != rightVal
		}
		ptr, err := memory.AllocBoolean(val)
		if err != nil { return err }
		return vm.push(ptr)
	}

	// Fallback equality
	if op == compiler.OpEqual {
		return vm.push(FalsePtr)
	}
	if op == compiler.OpNotEqual {
		return vm.push(TruePtr)
	}

	return fmt.Errorf("unsupported comparison")
}

func (vm *VM) executeBitwiseOperation(op compiler.Opcode) error {
	return fmt.Errorf("bitwise not implemented in phase x.4")
}

func (vm *VM) executeBangOperator() error {
	operand, err := vm.pop()
	if err != nil { return err }

	header, _ := memory.ReadHeader(operand)
	isTruthy := true
	if operand == NullPtr || header.Type == memory.TagNull {
		isTruthy = false
	} else if header.Type == memory.TagBoolean {
		val, _ := memory.ReadBoolean(operand)
		isTruthy = val
	} else if header.Type == memory.TagInteger {
		val, _ := memory.ReadInteger(operand)
		isTruthy = val != 0
	}

	ptr, _ := memory.AllocBoolean(!isTruthy)
	return vm.push(ptr)
}

func (vm *VM) executeMinusOperator() error {
	operand, err := vm.pop()
	if err != nil { return err }
	header, _ := memory.ReadHeader(operand)
	if header.Type == memory.TagInteger {
		val, _ := memory.ReadInteger(operand)
		ptr, _ := memory.AllocInteger(-val)
		return vm.push(ptr)
	} else if header.Type == memory.TagFloat {
		val, _ := memory.ReadFloat(operand)
		ptr, _ := memory.AllocFloat(-val)
		return vm.push(ptr)
	}
	return fmt.Errorf("minus not supported for type tag %d", header.Type)
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
	header, err := memory.ReadHeader(left)
	if err != nil { return err }

	if header.Type == memory.TagArray {
		idx, err := memory.ReadInteger(index)
		if err != nil { return fmt.Errorf("index must be integer") }

		elemPtr, err := memory.ReadArrayElement(left, int(idx))
		if err != nil { return vm.push(NullPtr) } // Bounds check failed -> Null
		return vm.push(elemPtr)
	}

	if header.Type == memory.TagHash {
		count, _ := memory.ReadHashCount(left)
		for i := 0; i < count; i++ {
			k, v, _ := memory.ReadHashPair(left, i)
			if equals(k, index) {
				return vm.push(v)
			}
		}
		return vm.push(NullPtr)
	}

	return fmt.Errorf("index not supported for type tag %d", header.Type)
}

func (vm *VM) executeSetIndexExpression(left, index, val memory.Ptr) error {
	return fmt.Errorf("set index not implemented")
}

func equals(p1, p2 memory.Ptr) bool {
	if p1 == p2 { return true }

	h1, _ := memory.ReadHeader(p1)
	h2, _ := memory.ReadHeader(p2)

	if h1.Type != h2.Type { return false }

	if h1.Type == memory.TagInteger {
		v1, _ := memory.ReadInteger(p1)
		v2, _ := memory.ReadInteger(p2)
		return v1 == v2
	}
	if h1.Type == memory.TagString {
		v1, _ := memory.ReadString(p1)
		v2, _ := memory.ReadString(p2)
		return v1 == v2
	}
	if h1.Type == memory.TagBoolean {
		v1, _ := memory.ReadBoolean(p1)
		v2, _ := memory.ReadBoolean(p2)
		return v1 == v2
	}
	return false
}
