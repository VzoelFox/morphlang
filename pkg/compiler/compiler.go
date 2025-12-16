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

type CompilationScope struct {
	instructions        Instructions
	lastInstruction     EmittedInstruction
	previousInstruction EmittedInstruction
}

type Compiler struct {
	instructions        Instructions
	constants           []object.Object
	lastInstruction     EmittedInstruction
	previousInstruction EmittedInstruction
	symbolTable         *SymbolTable
	scopes              []CompilationScope
	scopeIndex          int
}

func New() *Compiler {
	mainScope := CompilationScope{
		instructions:        Instructions{},
		lastInstruction:     EmittedInstruction{},
		previousInstruction: EmittedInstruction{},
	}

	return &Compiler{
		constants:   []object.Object{},
		symbolTable: NewSymbolTable(),
		scopes:      []CompilationScope{mainScope},
		scopeIndex:  0,
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

		// Handle named function declarations: define variable and store it
		if fn, ok := node.Expression.(*parser.FunctionLiteral); ok && fn.Name != "" {
			symbol := c.symbolTable.Define(fn.Name)

			err := c.Compile(fn)
			if err != nil {
				return err
			}

			if symbol.Scope == GlobalScope {
				c.emit(OpStoreGlobal, symbol.Index)
			} else {
				c.emit(OpStoreLocal, symbol.Index)
			}
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

		if symbol.Scope == GlobalScope {
			c.emit(OpStoreGlobal, symbol.Index)
		} else {
			c.emit(OpStoreLocal, symbol.Index)
		}

	case *parser.Identifier:
		symbol, ok := c.symbolTable.Resolve(node.Value)
		if !ok {
			builtinIndex := object.GetBuiltinByName(node.Value)
			if builtinIndex >= 0 {
				c.emit(OpGetBuiltin, builtinIndex)
				return nil
			}
			return fmt.Errorf("undefined variable %s", node.Value)
		}

		if symbol.Scope == GlobalScope {
			c.emit(OpLoadGlobal, symbol.Index)
		} else if symbol.Scope == LocalScope {
			c.emit(OpLoadLocal, symbol.Index)
		} else if symbol.Scope == FreeScope {
			c.emit(OpGetFree, symbol.Index)
		}

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

		if c.scopes[c.scopeIndex].lastInstruction.Opcode == OpPop {
			c.removeLastPop()
		}

		// Jump over alternative if consequence was executed
		jumpPos := c.emit(OpJump, 9999)

		afterConsequencePos := len(c.currentInstructions())
		c.changeOperand(jumpNotTruthyPos, afterConsequencePos)

		if node.Alternative == nil {
			// If no alternative, we need to push NULL so the expression has a value
			c.emit(OpLoadConst, c.addConstant(&object.Null{}))
		} else {
			err = c.Compile(node.Alternative)
			if err != nil {
				return err
			}

			if c.scopes[c.scopeIndex].lastInstruction.Opcode == OpPop {
				c.removeLastPop()
			}
		}

		afterAlternativePos := len(c.currentInstructions())
		c.changeOperand(jumpPos, afterAlternativePos)

	case *parser.WhileExpression:
		// 1. Initialize result with Null
		c.emit(OpLoadConst, c.addConstant(&object.Null{}))

		loopStartPos := len(c.currentInstructions())

		// 2. Condition
		err := c.Compile(node.Condition)
		if err != nil {
			return err
		}

		// 3. Jump if False to End
		jumpNotTruthyPos := c.emit(OpJumpNotTruthy, 9999)

		// 4. Pop previous result
		c.emit(OpPop)

		// 5. Body
		err = c.Compile(node.Body)
		if err != nil {
			return err
		}

		if c.scopes[c.scopeIndex].lastInstruction.Opcode == OpPop {
			c.removeLastPop()
		} else {
			c.emit(OpLoadConst, c.addConstant(&object.Null{}))
		}

		// 6. Jump back to Start
		c.emit(OpJump, loopStartPos)

		// 7. Patch JumpNotTruthy
		afterLoopPos := len(c.currentInstructions())
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

	case *parser.FunctionLiteral:
		c.EnterScope()

		for _, p := range node.Parameters {
			c.symbolTable.Define(p.Value)
		}

		err := c.Compile(node.Body)
		if err != nil {
			return err
		}

		if c.scopes[c.scopeIndex].lastInstruction.Opcode == OpPop {
			c.removeLastPop()
			c.emit(OpReturnValue)
		} else {
			c.emit(OpReturn)
		}

		numLocals := c.symbolTable.numDefinitions
		freeSymbols := c.symbolTable.FreeSymbols
		instructions := c.LeaveScope()

		for _, s := range freeSymbols {
			symbol, ok := c.symbolTable.Resolve(s.Name)
			if !ok {
				return fmt.Errorf("free variable %s could not be resolved", s.Name)
			}

			if symbol.Scope == GlobalScope {
				c.emit(OpLoadGlobal, symbol.Index)
			} else if symbol.Scope == LocalScope {
				c.emit(OpLoadLocal, symbol.Index)
			} else if symbol.Scope == FreeScope {
				c.emit(OpGetFree, symbol.Index)
			}
		}

		compiledFn := &object.CompiledFunction{
			Instructions:  instructions,
			NumLocals:     numLocals,
			NumParameters: len(node.Parameters),
		}

		fnIndex := c.addConstant(compiledFn)
		c.emit(OpClosure, fnIndex, len(freeSymbols))

	case *parser.ReturnStatement:
		err := c.Compile(node.ReturnValue)
		if err != nil {
			return err
		}
		c.emit(OpReturnValue)

	case *parser.CallExpression:
		err := c.Compile(node.Function)
		if err != nil {
			return err
		}

		for _, a := range node.Arguments {
			err := c.Compile(a)
			if err != nil {
				return err
			}
		}

		c.emit(OpCall, len(node.Arguments))
	}

	return nil
}

func (c *Compiler) Bytecode() *Bytecode {
	return &Bytecode{
		Instructions: c.scopes[0].instructions,
		Constants:    c.constants,
	}
}

func (c *Compiler) currentInstructions() Instructions {
	return c.scopes[c.scopeIndex].instructions
}

func (c *Compiler) EnterScope() {
	scope := CompilationScope{
		instructions:        Instructions{},
		lastInstruction:     EmittedInstruction{},
		previousInstruction: EmittedInstruction{},
	}
	c.scopes = append(c.scopes, scope)
	c.scopeIndex++
	c.symbolTable = NewEnclosedSymbolTable(c.symbolTable)
}

func (c *Compiler) LeaveScope() Instructions {
	instructions := c.currentInstructions()

	c.scopes = c.scopes[:len(c.scopes)-1]
	c.scopeIndex--
	c.symbolTable = c.symbolTable.Outer

	return instructions
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
	posNewInstruction := len(c.currentInstructions())
	updatedInstructions := append(c.currentInstructions(), ins...)
	c.scopes[c.scopeIndex].instructions = updatedInstructions
	return posNewInstruction
}

func (c *Compiler) setLastInstruction(op Opcode, pos int) {
	previous := c.scopes[c.scopeIndex].lastInstruction
	last := EmittedInstruction{Opcode: op, Position: pos}

	c.scopes[c.scopeIndex].previousInstruction = previous
	c.scopes[c.scopeIndex].lastInstruction = last
}

func (c *Compiler) removeLastPop() {
	last := c.scopes[c.scopeIndex].lastInstruction
	previous := c.scopes[c.scopeIndex].previousInstruction

	old := c.currentInstructions()
	if last.Position >= len(old) {
		return
	}
	new := old[:last.Position]

	c.scopes[c.scopeIndex].instructions = new
	c.scopes[c.scopeIndex].lastInstruction = previous
}

func (c *Compiler) replaceInstruction(pos int, newInstruction []byte) {
	ins := c.currentInstructions()
	for i := 0; i < len(newInstruction); i++ {
		ins[pos+i] = newInstruction[i]
	}
}

func (c *Compiler) changeOperand(opPos int, operand int) {
	op := Opcode(c.currentInstructions()[opPos])
	newInstruction := Make(op, operand)

	c.replaceInstruction(opPos, newInstruction)
}
