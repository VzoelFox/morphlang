package vm

import (
	"fmt"

	"github.com/VzoelFox/morphlang/pkg/compiler"
	"github.com/VzoelFox/morphlang/pkg/object"
)

const StackSize = 2048

type VM struct {
	constants    []object.Object
	instructions compiler.Instructions

	stack [StackSize]object.Object
	sp    int // Always points to the next value. Top of stack is stack[sp-1]

	LastPoppedStackElem object.Object
}

func New(bytecode *compiler.Bytecode) *VM {
	return &VM{
		instructions: bytecode.Instructions,
		constants:    bytecode.Constants,
		sp:           0,
	}
}

func (vm *VM) StackTop() object.Object {
	if vm.sp == 0 {
		return nil
	}
	return vm.stack[vm.sp-1]
}

func (vm *VM) Run() error {
	for ip := 0; ip < len(vm.instructions); ip++ {
		op := compiler.Opcode(vm.instructions[ip])

		switch op {
		case compiler.OpLoadConst:
			constIndex := compiler.ReadUint16(vm.instructions[ip+1:])
			ip += 2
			err := vm.push(vm.constants[constIndex])
			if err != nil {
				return err
			}

		case compiler.OpAdd, compiler.OpSub, compiler.OpMul, compiler.OpDiv:
			err := vm.executeBinaryOperation(op)
			if err != nil {
				return err
			}

		case compiler.OpPop:
			vm.pop()
		}
	}

	return nil
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

	leftType := left.Type()
	rightType := right.Type()

	if leftType == object.INTEGER_OBJ && rightType == object.INTEGER_OBJ {
		return vm.executeBinaryIntegerOperation(op, left, right)
	}

	return fmt.Errorf("unsupported types for binary operation: %s %s", leftType, rightType)
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
		result = leftVal / rightVal
	default:
		return fmt.Errorf("unknown integer operator: %d", op)
	}

	return vm.push(&object.Integer{Value: result})
}
