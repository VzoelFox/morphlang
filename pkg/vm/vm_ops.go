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

	// String Concatenation (Mixed Support)
	if leftHeader.Type == memory.TagString || rightHeader.Type == memory.TagString {
		if op != compiler.OpAdd { return fmt.Errorf("string only supports add") }

		leftStr := stringify(left)
		rightStr := stringify(right)

		ptr, err := memory.AllocString(leftStr + rightStr)
		if err != nil { return err }
		return vm.push(ptr)
	}

	// Numeric Arithmetic
	isFloat := leftHeader.Type == memory.TagFloat || rightHeader.Type == memory.TagFloat

	if isFloat {
		leftVal := toFloat(left)
		rightVal := toFloat(right)
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
	} else {
		// Integer
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
}

func (vm *VM) executeComparison(op compiler.Opcode) error {
	right, err := vm.pop()
	if err != nil { return err }
	left, err := vm.pop()
	if err != nil { return err }

	leftHeader, _ := memory.ReadHeader(left)
	rightHeader, _ := memory.ReadHeader(right)

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
		ptr, _ := memory.AllocBoolean(val)
		return vm.push(ptr)
	}

	// Mixed Float/Int Comparison
	if (leftHeader.Type == memory.TagInteger || leftHeader.Type == memory.TagFloat) &&
	   (rightHeader.Type == memory.TagInteger || rightHeader.Type == memory.TagFloat) {
		leftVal := toFloat(left)
		rightVal := toFloat(right)
		var val bool
		switch op {
		case compiler.OpEqual: val = leftVal == rightVal
		case compiler.OpNotEqual: val = leftVal != rightVal
		case compiler.OpGreaterThan: val = leftVal > rightVal
		case compiler.OpGreaterEqual: val = leftVal >= rightVal
		}
		ptr, _ := memory.AllocBoolean(val)
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
		ptr, _ := memory.AllocBoolean(val)
		return vm.push(ptr)
	}

	if leftHeader.Type == memory.TagString && rightHeader.Type == memory.TagString {
		leftVal, _ := memory.ReadString(left)
		rightVal, _ := memory.ReadString(right)
		var val bool
		switch op {
		case compiler.OpEqual: val = leftVal == rightVal
		case compiler.OpNotEqual: val = leftVal != rightVal
		}
		ptr, _ := memory.AllocBoolean(val)
		return vm.push(ptr)
	}

	if op == compiler.OpEqual {
		return vm.push(FalsePtr) // Default false for different types
	}
	if op == compiler.OpNotEqual {
		return vm.push(TruePtr)
	}

	return fmt.Errorf("unsupported comparison")
}

func (vm *VM) executeBitwiseOperation(op compiler.Opcode) error {
	right, err := vm.pop()
	if err != nil { return err }
	left, err := vm.pop()
	if err != nil { return err }

	leftHeader, _ := memory.ReadHeader(left)
	rightHeader, _ := memory.ReadHeader(right)

	if leftHeader.Type != memory.TagInteger || rightHeader.Type != memory.TagInteger {
		return fmt.Errorf("unsupported types for bitwise operation: type tag %d type tag %d", leftHeader.Type, rightHeader.Type)
	}

	leftVal, _ := memory.ReadInteger(left)
	rightVal, _ := memory.ReadInteger(right)

	var res int64
	switch op {
	case compiler.OpAnd:
		res = leftVal & rightVal
	case compiler.OpOr:
		res = leftVal | rightVal
	case compiler.OpXor:
		res = leftVal ^ rightVal
	case compiler.OpLShift:
		if rightVal < 0 { return fmt.Errorf("negative shift count") }
		res = leftVal << rightVal
	case compiler.OpRShift:
		if rightVal < 0 { return fmt.Errorf("negative shift count") }
		res = leftVal >> rightVal
	}

	ptr, err := memory.AllocInteger(res)
	if err != nil { return err }
	return vm.push(ptr)
}

func (vm *VM) executeBangOperator() error {
	operand, err := vm.pop()
	if err != nil { return err }
	ptr, _ := memory.AllocBoolean(!isTruthy(operand))
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
	operand, err := vm.pop()
	if err != nil { return err }
	header, _ := memory.ReadHeader(operand)
	if header.Type == memory.TagInteger {
		val, _ := memory.ReadInteger(operand)
		ptr, _ := memory.AllocInteger(^val)
		return vm.push(ptr)
	}
	return fmt.Errorf("bitnot not supported for type tag %d", header.Type)
}

func (vm *VM) buildArray(startIndex, endIndex int) (memory.Ptr, error) {
	length := endIndex - startIndex
	ptr, err := memory.AllocArray(length, length)
	if err != nil { return memory.NilPtr, err }

	for i := 0; i < length; i++ {
		elemPtr := vm.stack[startIndex+i]
		memory.WriteArrayElement(ptr, i, elemPtr)
	}
	return ptr, nil
}

func (vm *VM) buildHash(startIndex, endIndex int) (memory.Ptr, error) {
	count := (endIndex - startIndex) / 2
	ptr, err := memory.AllocHash(count)
	if err != nil { return memory.NilPtr, err }

	for i := 0; i < count; i++ {
		key := vm.stack[startIndex + i*2]
		val := vm.stack[startIndex + i*2 + 1]
		memory.WriteHashPair(ptr, i, key, val)
	}
	return ptr, nil
}

func (vm *VM) executeIndexExpression(left, index memory.Ptr) error {
	header, err := memory.ReadHeader(left)
	if err != nil { return err }

	if header.Type == memory.TagArray {
		idx, err := memory.ReadInteger(index)
		if err != nil { return fmt.Errorf("index must be integer") }

		elemPtr, err := memory.ReadArrayElement(left, int(idx))
		if err != nil { return vm.push(NullPtr) }
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

	if header.Type == memory.TagString {
		idx, err := memory.ReadInteger(index)
		if err != nil { return fmt.Errorf("string index must be integer") }

		str, err := memory.ReadString(left)
		if err != nil { return err }

		if idx < 0 || int(idx) >= len(str) {
			return vm.push(NullPtr)
		}

		charStr := string(str[idx])
		ptr, err := memory.AllocString(charStr)
		if err != nil { return err }
		return vm.push(ptr)
	}

	return fmt.Errorf("index not supported for type tag %d", header.Type)
}

func (vm *VM) executeSetIndexExpression(left, index, val memory.Ptr) error {
	header, _ := memory.ReadHeader(left)

	if header.Type == memory.TagArray {
		idx, err := memory.ReadInteger(index)
		if err != nil { return fmt.Errorf("array index must be integer") }

		err = memory.WriteArrayElement(left, int(idx), val)
		if err != nil { return err }
		// Assignment usually returns the value assigned? Or Null?
		// We push Null to signify statement done (popped by Block).
		// Or push Val if it's an expression.
		// Standard: Assignment expression evaluates to assigned value.
		// So we push `val`. (But we popped it).
		// Wait, OpSetIndex pops (left, index, val).
		// If we want to return it, we push `val` back.
		// But in Morph, assignment is statement?
		// Compiler emits `OpSetIndex` for `AssignmentStatement` (IndexExpression).
		// `Compile` for `AssignmentStatement` does NOT emit `OpPop`?
		// Let's check `compiler.go`.
		// `c.emit(OpSetIndex)`.
		// If `OpSetIndex` pushes nothing, then stack is clean.
		// If `OpSetIndex` pushes something, it remains on stack.
		// `ExpressionStatement` emits `OpPop`.
		// `AssignmentStatement` is a statement.
		// Does it return value?
		// In `evaluator`, it returns `Null`.
		// So `OpSetIndex` should push `NullPtr`.
		return vm.push(NullPtr)
	}

	if header.Type == memory.TagHash {
		count, _ := memory.ReadHashCount(left)
		// Update existing
		for i := 0; i < count; i++ {
			k, _, _ := memory.ReadHashPair(left, i)
			if equals(k, index) {
				memory.WriteHashPair(left, i, k, val)
				return vm.push(NullPtr)
			}
		}
		// Append (Dynamic) - Not supported in fixed Hash
		return fmt.Errorf("hash update key not found (dynamic hash unsupported)")
	}

	return fmt.Errorf("set index not supported")
}

// Helpers

func isTruthy(ptr memory.Ptr) bool {
	if ptr == memory.NilPtr { return false }
	header, _ := memory.ReadHeader(ptr)
	if header.Type == memory.TagNull { return false }
	if header.Type == memory.TagBoolean {
		val, _ := memory.ReadBoolean(ptr)
		return val
	}
	if header.Type == memory.TagInteger {
		val, _ := memory.ReadInteger(ptr)
		return val != 0
	}
	return true
}

func toFloat(ptr memory.Ptr) float64 {
	header, _ := memory.ReadHeader(ptr)
	if header.Type == memory.TagFloat {
		val, _ := memory.ReadFloat(ptr)
		return val
	}
	if header.Type == memory.TagInteger {
		val, _ := memory.ReadInteger(ptr)
		return float64(val)
	}
	return 0
}

func stringify(ptr memory.Ptr) string {
	header, _ := memory.ReadHeader(ptr)
	if header.Type == memory.TagString {
		val, _ := memory.ReadString(ptr)
		return val
	}
	if header.Type == memory.TagInteger {
		val, _ := memory.ReadInteger(ptr)
		return fmt.Sprintf("%d", val)
	}
	if header.Type == memory.TagFloat {
		val, _ := memory.ReadFloat(ptr)
		return fmt.Sprintf("%g", val)
	}
	if header.Type == memory.TagBoolean {
		val, _ := memory.ReadBoolean(ptr)
		if val { return "benar" } else { return "salah" }
	}
	if header.Type == memory.TagNull { return "kosong" }
	return fmt.Sprintf("ptr:%d", ptr)
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
