package compiler

import (
	"fmt"
	"sort"

	"github.com/VzoelFox/morphlang/pkg/object"
	"github.com/VzoelFox/morphlang/pkg/parser"
)

type EmittedInstruction struct {
	Opcode   Opcode
	Position int
}

type Compiler struct {
	instructions        Instructions
	constants           []object.Object
	lastInstruction     EmittedInstruction
	previousInstruction EmittedInstruction
	symbolTable         *SymbolTable
}

func New() *Compiler {
	return &Compiler{
		instructions:        Instructions{},
		constants:           []object.Object{},
		lastInstruction:     EmittedInstruction{},
		previousInstruction: EmittedInstruction{},
		symbolTable:         NewSymbolTable(),
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
		if node.Expression == nil {
			return nil
		}
		err := c.Compile(node.Expression)
		if err != nil {
			return err
		}
		c.emit(OpPop)

	case *parser.BlockStatement:
		for _, s := range node.Statements {
			err := c.Compile(s)
			if err != nil {
				return err
			}
		}

	case *parser.AssignmentStatement:
		err := c.Compile(node.Value)
		if err != nil {
			return err
		}

		symbol, ok := c.symbolTable.Resolve(node.Name.Value)
		if !ok {
			symbol = c.symbolTable.Define(node.Name.Value)
		}
		c.emit(OpStoreGlobal, symbol.Index)

	case *parser.Identifier:
		symbol, ok := c.symbolTable.Resolve(node.Value)
		if !ok {
			return fmt.Errorf("undefined variable %s", node.Value)
		}
		c.emit(OpLoadGlobal, symbol.Index)

	case *parser.IfExpression:
		err := c.Compile(node.Condition)
		if err != nil {
			return err
		}

		// Jump over consequence if condition is false
		jumpNotTruthyPos := c.emit(OpJumpNotTruthy, 9999)

		err = c.Compile(node.Consequence)
		if err != nil {
			return err
		}

		if c.lastInstruction.Opcode == OpPop {
			c.removeLastPop()
		}

		// Jump over alternative if consequence was executed
		jumpPos := c.emit(OpJump, 9999)

		afterConsequencePos := len(c.instructions)
		c.changeOperand(jumpNotTruthyPos, afterConsequencePos)

		if node.Alternative == nil {
			// If no alternative, we need to push NULL so the expression has a value
			c.emit(OpLoadConst, c.addConstant(&object.Null{}))
		} else {
			err = c.Compile(node.Alternative)
			if err != nil {
				return err
			}

			if c.lastInstruction.Opcode == OpPop {
				c.removeLastPop()
			}
		}

		afterAlternativePos := len(c.instructions)
		c.changeOperand(jumpPos, afterAlternativePos)

	case *parser.WhileExpression:
		// 1. Initialize result with Null (in case loop never runs)
		c.emit(OpLoadConst, c.addConstant(&object.Null{}))

		loopStartPos := len(c.instructions)

		// 2. Condition
		err := c.Compile(node.Condition)
		if err != nil {
			return err
		}

		// 3. Jump if False to End
		jumpNotTruthyPos := c.emit(OpJumpNotTruthy, 9999)

		// 4. Pop previous result (prepare stack for new iteration result)
		c.emit(OpPop)

		// 5. Body
		err = c.Compile(node.Body)
		if err != nil {
			return err
		}

		if c.lastInstruction.Opcode == OpPop {
			c.removeLastPop()
		} else {
			// If body didn't end with expression, push Null
			c.emit(OpLoadConst, c.addConstant(&object.Null{}))
		}

		// 6. Jump back to Start
		c.emit(OpJump, loopStartPos)

		// 7. Patch JumpNotTruthy
		afterLoopPos := len(c.instructions)
		c.changeOperand(jumpNotTruthyPos, afterLoopPos)

	case *parser.InfixExpression:
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

	case *parser.PrefixExpression:
		err := c.Compile(node.Right)
		if err != nil {
			return err
		}

		switch node.Operator {
		case "!":
			c.emit(OpBang)
		case "-":
			c.emit(OpMinus)
		default:
			return fmt.Errorf("unknown operator %s", node.Operator)
		}

	case *parser.IntegerLiteral:
		integer := &object.Integer{Value: node.Value}
		c.emit(OpLoadConst, c.addConstant(integer))

	case *parser.BooleanLiteral:
		boolean := &object.Boolean{Value: node.Value}
		c.emit(OpLoadConst, c.addConstant(boolean))

	case *parser.StringLiteral:
		str := &object.String{Value: node.Value}
		c.emit(OpLoadConst, c.addConstant(str))

	case *parser.ArrayLiteral:
		for _, el := range node.Elements {
			err := c.Compile(el)
			if err != nil {
				return err
			}
		}
		c.emit(OpArray, len(node.Elements))

	case *parser.HashLiteral:
		keys := []parser.Expression{}
		for k := range node.Pairs {
			keys = append(keys, k)
		}
		sort.Slice(keys, func(i, j int) bool {
			return keys[i].String() < keys[j].String()
		})

		for _, key := range keys {
			err := c.Compile(key)
			if err != nil {
				return err
			}
			err = c.Compile(node.Pairs[key])
			if err != nil {
				return err
			}
		}
		c.emit(OpHash, len(node.Pairs)*2)

	case *parser.IndexExpression:
		err := c.Compile(node.Left)
		if err != nil {
			return err
		}
		err = c.Compile(node.Index)
		if err != nil {
			return err
		}
		c.emit(OpIndex)
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

	c.setLastInstruction(op, pos)

	return pos
}

func (c *Compiler) addInstruction(ins []byte) int {
	posNewInstruction := len(c.instructions)
	c.instructions = append(c.instructions, ins...)
	return posNewInstruction
}

func (c *Compiler) setLastInstruction(op Opcode, pos int) {
	previous := c.lastInstruction
	last := EmittedInstruction{Opcode: op, Position: pos}

	c.previousInstruction = previous
	c.lastInstruction = last
}

func (c *Compiler) removeLastPop() {
	c.instructions = c.instructions[:c.lastInstruction.Position]
	c.lastInstruction = c.previousInstruction
}

func (c *Compiler) replaceInstruction(pos int, newInstruction []byte) {
	for i := 0; i < len(newInstruction); i++ {
		c.instructions[pos+i] = newInstruction[i]
	}
}

func (c *Compiler) changeOperand(opPos int, operand int) {
	op := Opcode(c.instructions[opPos])
	newInstruction := Make(op, operand)

	c.replaceInstruction(opPos, newInstruction)
}
