package compiler

import (
	"fmt"

	"github.com/VzoelFox/morphlang/pkg/object"
	"github.com/VzoelFox/morphlang/pkg/parser"
)

type Compiler struct {
	instructions Instructions
	constants    []object.Object
}

func New() *Compiler {
	return &Compiler{
		instructions: Instructions{},
		constants:    []object.Object{},
	}
}

func (c *Compiler) Compile(node parser.Node) error {
	switch node := node.(type) {
	case *parser.Program:
		for _, s := range node.Statements {
			err := c.Compile(s)
			if err != nil {
				return err
			}
		}

	case *parser.ExpressionStatement:
		err := c.Compile(node.Expression)
		if err != nil {
			return err
		}
		c.emit(OpPop)

	case *parser.InfixExpression:
		// Re-order for comparison operators < and <=
		if node.Operator == "<" {
			err := c.Compile(node.Right)
			if err != nil {
				return err
			}
			err = c.Compile(node.Left)
			if err != nil {
				return err
			}
			c.emit(OpGreaterThan)
			return nil
		}
		// TODO: Handle <= if/when we have it, or transform it.
		// For now we only have OpGreaterThan in opcodes?
		// Spec has:
		// 0x26 | GT | - | Pop b, Pop a, Push a > b.
		// 0x27 | GTE | - | Pop b, Pop a, Push a >= b.
		// It doesn't seem to have LT or LTE. The standard way is to swap operands.
		// a < b  <=> b > a
		// a <= b <=> b >= a

		err := c.Compile(node.Left)
		if err != nil {
			return err
		}

		err = c.Compile(node.Right)
		if err != nil {
			return err
		}

		switch node.Operator {
		case "+":
			c.emit(OpAdd)
		case "-":
			c.emit(OpSub)
		case "*":
			c.emit(OpMul)
		case "/":
			c.emit(OpDiv)
		case ">":
			c.emit(OpGreaterThan)
		case "==":
			c.emit(OpEqual)
		case "!=":
			c.emit(OpNotEqual)
		default:
			return fmt.Errorf("unknown operator %s", node.Operator)
		}

	case *parser.IntegerLiteral:
		integer := &object.Integer{Value: node.Value}
		c.emit(OpLoadConst, c.addConstant(integer))
	}

	return nil
}

func (c *Compiler) Bytecode() *Bytecode {
	return &Bytecode{
		Instructions: c.instructions,
		Constants:    c.constants,
	}
}

type Bytecode struct {
	Instructions Instructions
	Constants    []object.Object
}

func (c *Compiler) addConstant(obj object.Object) int {
	c.constants = append(c.constants, obj)
	return len(c.constants) - 1
}

func (c *Compiler) emit(op Opcode, operands ...int) int {
	ins := Make(op, operands...)
	pos := c.addInstruction(ins)
	return pos
}

func (c *Compiler) addInstruction(ins []byte) int {
	posNewInstruction := len(c.instructions)
	c.instructions = append(c.instructions, ins...)
	return posNewInstruction
}
