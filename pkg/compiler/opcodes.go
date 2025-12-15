package compiler

import "fmt"

type Opcode byte

const (
	// Stack Manipulation
	OpPop Opcode = 0x01

	// Constants & Variables
	OpLoadConst  Opcode = 0x10
	OpLoadGlobal Opcode = 0x11
	OpStoreGlobal Opcode = 0x12
	OpLoadLocal  Opcode = 0x13
	OpStoreLocal Opcode = 0x14

	// Arithmetic & Logic
	OpAdd Opcode = 0x20
	OpSub Opcode = 0x21
	OpMul Opcode = 0x22
	OpDiv Opcode = 0x23
	OpEqual    Opcode = 0x24
	OpNotEqual Opcode = 0x25
	OpGreaterThan Opcode = 0x26
	OpGreaterEqual Opcode = 0x27
	OpMinus Opcode = 0x2F // Unary minus
	OpBang  Opcode = 0x2E // Unary not

	// Control Flow
	OpJump        Opcode = 0x30
	OpJumpNotTruthy Opcode = 0x31 // JUMP_IF_FALSE in spec

	// Functions
	OpCall      Opcode = 0x40
	OpReturn    Opcode = 0x41 // RETURN in spec (returns null/void)
	OpReturnValue Opcode = 0x42 // RETURN_VAL in spec (returns value)
)

type Definition struct {
	Name          string
	OperandWidths []int
}

var definitions = map[Opcode]*Definition{
	OpPop:         {"OpPop", []int{}},
	OpLoadConst:   {"OpLoadConst", []int{2}}, // u16 index
	OpLoadGlobal:  {"OpLoadGlobal", []int{2}}, // u16 index
	OpStoreGlobal: {"OpStoreGlobal", []int{2}}, // u16 index
	OpLoadLocal:   {"OpLoadLocal", []int{1}}, // u8 index
	OpStoreLocal:  {"OpStoreLocal", []int{1}}, // u8 index
	OpAdd:         {"OpAdd", []int{}},
	OpSub:         {"OpSub", []int{}},
	OpMul:         {"OpMul", []int{}},
	OpDiv:         {"OpDiv", []int{}},
	OpEqual:       {"OpEqual", []int{}},
	OpNotEqual:    {"OpNotEqual", []int{}},
	OpGreaterThan: {"OpGreaterThan", []int{}},
	OpGreaterEqual: {"OpGreaterEqual", []int{}},
	OpMinus:       {"OpMinus", []int{}},
	OpBang:        {"OpBang", []int{}},
	OpJump:        {"OpJump", []int{2}}, // u16 offset
	OpJumpNotTruthy: {"OpJumpNotTruthy", []int{2}}, // u16 offset
	OpCall:        {"OpCall", []int{1}}, // u8 numArgs
	OpReturn:      {"OpReturn", []int{}},
	OpReturnValue: {"OpReturnValue", []int{}},
}

func Lookup(op byte) (*Definition, error) {
	def, ok := definitions[Opcode(op)]
	if !ok {
		return nil, fmt.Errorf("opcode %d undefined", op)
	}
	return def, nil
}
