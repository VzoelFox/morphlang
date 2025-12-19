package vm

import (
	"fmt"
	"github.com/VzoelFox/morphlang/pkg/compiler"
	"github.com/VzoelFox/morphlang/pkg/memory"
)

// Helper to check and propagate error objects
func (vm *VM) checkAndPropagateError(ptrs ...memory.Ptr) (bool, error) {
	for _, ptr := range ptrs {
		header, err := memory.ReadHeader(ptr)
		if err != nil { return false, err }
		if header.Type == memory.TagError {
			// Push error back to stack as result
			return true, vm.push(ptr)
		}
	}
	return false, nil
}

// Helper to create and push runtime error
func (vm *VM) pushRuntimeError(msg string) error {
	errPtr, err := vm.newError(msg)
	if err != nil { return err }
	return vm.push(errPtr)
}

func (vm *VM) executeBinaryOperation(op compiler.Opcode) error {
	right, err := vm.pop()
	if err != nil { return err }
	left, err := vm.pop()
	if err != nil { return err }

	if handled, err := vm.checkAndPropagateError(left, right); handled || err != nil {
		return err
	}

	leftHeader, err := memory.ReadHeader(left)
	if err != nil { return err }
	rightHeader, err := memory.ReadHeader(right)
	if err != nil { return err }

	// String Concatenation (Mixed Support)
	if leftHeader.Type == memory.TagString || rightHeader.Type == memory.TagString {
		if op != compiler.OpAdd { return vm.pushRuntimeError("string only supports add") }

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
			if rightVal == 0 { return vm.pushRuntimeError("integer divide by zero") }
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

	if handled, err := vm.checkAndPropagateError(left, right); handled || err != nil {
		return err
	}

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

	if leftHeader.Type == memory.TagNull && rightHeader.Type == memory.TagNull {
		if op == compiler.OpEqual { return vm.push(TruePtr) }
		if op == compiler.OpNotEqual { return vm.push(FalsePtr) }
	}

	if op == compiler.OpEqual {
		return vm.push(FalsePtr) // Default false for different types
	}
	if op == compiler.OpNotEqual {
		return vm.push(TruePtr)
	}

	return vm.pushRuntimeError("unsupported comparison")
}

func (vm *VM) executeBitwiseOperation(op compiler.Opcode) error {
	right, err := vm.pop()
	if err != nil { return err }
	left, err := vm.pop()
	if err != nil { return err }

	if handled, err := vm.checkAndPropagateError(left, right); handled || err != nil {
		return err
	}

	leftHeader, _ := memory.ReadHeader(left)
	rightHeader, _ := memory.ReadHeader(right)

	if leftHeader.Type != memory.TagInteger || rightHeader.Type != memory.TagInteger {
		return vm.pushRuntimeError(fmt.Sprintf("unsupported types for bitwise operation: type tag %d type tag %d", leftHeader.Type, rightHeader.Type))
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
		if rightVal < 0 { return vm.pushRuntimeError("negative shift count") }
		res = leftVal << rightVal
	case compiler.OpRShift:
		if rightVal < 0 { return vm.pushRuntimeError("negative shift count") }
		res = leftVal >> rightVal
	}

	ptr, err := memory.AllocInteger(res)
	if err != nil { return err }
	return vm.push(ptr)
}

func (vm *VM) executeBangOperator() error {
	operand, err := vm.pop()
	if err != nil { return err }

	// Bang can work on error (Error is truthy?), or propagate?
	// Usually !Error -> False (since Error is truthy object)
	// OR we can say logic ops treat Error as Error.
	// Let's propagate for safety.
	if handled, err := vm.checkAndPropagateError(operand); handled || err != nil {
		return err
	}

	ptr, _ := memory.AllocBoolean(!isTruthy(operand))
	return vm.push(ptr)
}

func (vm *VM) executeMinusOperator() error {
	operand, err := vm.pop()
	if err != nil { return err }

	if handled, err := vm.checkAndPropagateError(operand); handled || err != nil {
		return err
	}

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
	return vm.pushRuntimeError(fmt.Sprintf("minus not supported for type tag %d", header.Type))
}

func (vm *VM) executeBitNotOperator() error {
	operand, err := vm.pop()
	if err != nil { return err }

	if handled, err := vm.checkAndPropagateError(operand); handled || err != nil {
		return err
	}

	header, _ := memory.ReadHeader(operand)
	if header.Type == memory.TagInteger {
		val, _ := memory.ReadInteger(operand)
		ptr, _ := memory.AllocInteger(^val)
		return vm.push(ptr)
	}
	return vm.pushRuntimeError(fmt.Sprintf("bitnot not supported for type tag %d", header.Type))
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
	if handled, err := vm.checkAndPropagateError(left, index); handled || err != nil {
		return err
	}

	header, err := memory.ReadHeader(left)
	if err != nil { return err }

	if header.Type == memory.TagArray {
		idx, err := memory.ReadInteger(index)
		if err != nil { return vm.pushRuntimeError("index must be integer") }

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
		if err != nil { return vm.pushRuntimeError("string index must be integer") }

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

	if header.Type == memory.TagStruct {
		schemaPtr, err := memory.ReadStructSchema(left)
		if err != nil { return err }

		_, fieldsArrPtr, err := memory.ReadSchema(schemaPtr)
		if err != nil { return err }

		length, _ := memory.ReadArrayLength(fieldsArrPtr)

		// Expect index to be String
		targetKey, err := memory.ReadString(index)
		if err != nil { return vm.pushRuntimeError("struct key must be string") }

		for i := 0; i < length; i++ {
			fieldPtr, _ := memory.ReadArrayElement(fieldsArrPtr, i)
			fieldName, _ := memory.ReadString(fieldPtr)
			if fieldName == targetKey {
				valPtr, err := memory.ReadStructField(left, i)
				if err != nil { return err }
				return vm.push(valPtr)
			}
		}
		return vm.push(NullPtr)
	}

	return vm.pushRuntimeError(fmt.Sprintf("index not supported for type tag %d", header.Type))
}

func (vm *VM) executeSetIndexExpression(left, index, val memory.Ptr) error {
	if handled, err := vm.checkAndPropagateError(left, index, val); handled || err != nil {
		return err
	}

	header, _ := memory.ReadHeader(left)

	if header.Type == memory.TagArray {
		idx, err := memory.ReadInteger(index)
		if err != nil { return vm.pushRuntimeError("array index must be integer") }

		err = memory.WriteArrayElement(left, int(idx), val)
		if err != nil { return err }
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
		return vm.pushRuntimeError("hash update key not found (dynamic hash unsupported)")
	}

	return vm.pushRuntimeError("set index not supported")
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
	// Error is truthy?
	// TagError is not handled here, so returns true.
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
	if header.Type == memory.TagError {
		msgPtr, _, _, _, _ := memory.ReadError(ptr)
		msg, _ := memory.ReadString(msgPtr)
		return "Error: " + msg
	}
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
	if h1.Type == memory.TagNull { return true }
	return false
}
