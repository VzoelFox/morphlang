package compiler

import "fmt"

type Opcode byte

const (
	// Stack Manipulation
	OpPop Opcode = 0x01
	OpDup Opcode = 0x02

	// Constants & Variables
	OpLoadConst  Opcode = 0x10
	OpLoadGlobal Opcode = 0x11
	OpStoreGlobal Opcode = 0x12
	OpLoadLocal  Opcode = 0x13
	OpStoreLocal Opcode = 0x14
	OpGetBuiltin Opcode = 0x15

	// Collections
	OpArray Opcode = 0x1A
	OpHash  Opcode = 0x1B
	OpIndex Opcode = 0x1C
	OpSetIndex Opcode = 0x1D

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

	// Bitwise
	OpAnd     Opcode = 0x28
	OpOr      Opcode = 0x29
	OpXor     Opcode = 0x2A
	OpLShift  Opcode = 0x2B
	OpRShift  Opcode = 0x2C
	OpBitNot  Opcode = 0x2D // Unary bitwise not

	// Control Flow
	OpJump        Opcode = 0x30
	OpJumpNotTruthy Opcode = 0x31 // JUMP_IF_FALSE in spec

	// Functions
	OpCall      Opcode = 0x40
	OpReturn    Opcode = 0x41 // RETURN in spec (returns null/void)
	OpReturnValue Opcode = 0x42 // RETURN_VAL in spec (returns value)
	OpClosure   Opcode = 0x43
	OpGetFree   Opcode = 0x44

	// Modules
	OpUpdateModule Opcode = 0x50
)

type Definition struct {
	Name          string
	OperandWidths []int
}

var definitions = map[Opcode]*Definition{
	OpPop:         {"OpPop", []int{}},
	OpDup:         {"OpDup", []int{}},
	OpLoadConst:   {"OpLoadConst", []int{2}}, // u16 index
	OpLoadGlobal:  {"OpLoadGlobal", []int{2}}, // u16 index
	OpStoreGlobal: {"OpStoreGlobal", []int{2}}, // u16 index
	OpLoadLocal:   {"OpLoadLocal", []int{1}}, // u8 index
	OpStoreLocal:  {"OpStoreLocal", []int{1}}, // u8 index
	OpGetBuiltin:  {"OpGetBuiltin", []int{1}}, // u8 index
	OpArray:       {"OpArray", []int{2}}, // u16 element count
	OpHash:        {"OpHash", []int{2}}, // u16 pair count (keys + values)
	OpIndex:       {"OpIndex", []int{}},
	OpSetIndex:    {"OpSetIndex", []int{}},
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
	OpAnd:         {"OpAnd", []int{}},
	OpOr:          {"OpOr", []int{}},
	OpXor:         {"OpXor", []int{}},
	OpLShift:      {"OpLShift", []int{}},
	OpRShift:      {"OpRShift", []int{}},
	OpBitNot:      {"OpBitNot", []int{}},
	OpJump:        {"OpJump", []int{2}}, // u16 offset
	OpJumpNotTruthy: {"OpJumpNotTruthy", []int{2}}, // u16 offset
	OpCall:        {"OpCall", []int{1}}, // u8 numArgs
	OpReturn:      {"OpReturn", []int{}},
	OpReturnValue: {"OpReturnValue", []int{}},
	OpClosure:     {"OpClosure", []int{2, 1}}, // u16 constIndex, u8 numFreeVars
	OpGetFree:     {"OpGetFree", []int{1}},    // u8 index
	OpUpdateModule: {"OpUpdateModule", []int{}},
}

func Lookup(op byte) (*Definition, error) {
	def, ok := definitions[Opcode(op)]
	if !ok {
		return nil, fmt.Errorf("opcode %d undefined", op)
	}
	return def, nil
}
