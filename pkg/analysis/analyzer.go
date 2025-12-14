package analysis

import (
	"crypto/sha256"
	"fmt"
	"strings"
	"time"

	"github.com/VzoelFox/morphlang/pkg/parser"
)

type Analyzer struct {
	program  *parser.Program
	filename string
	input    string
	context  *Context
	currFunc string
}

func GenerateContext(program *parser.Program, filename string, input string, parserErrors []string) (*Context, error) {
	ctx := &Context{
		Version:       "1.0.0",
		File:          filename,
		Timestamp:     time.Now(),
		Symbols:       make(map[string]*Symbol),
		GlobalVars:    make(map[string]*Variable),
		LocalScopes:   make(map[string]LocalScope),
		Errors:        []ParserError{},
		Warnings:      []Warning{},
		Imports:       []string{},
		TypeInference: make(map[string][]string),
		CallGraph:     make(map[string][]string),
	}

	// Checksum
	h := sha256.New()
	h.Write([]byte(input))
	ctx.Checksum = fmt.Sprintf("sha256:%x", h.Sum(nil))

	// Stats
	analyzeStatistics(ctx, input)

	// Parse Errors
	for _, err := range parserErrors {
		ctx.Errors = append(ctx.Errors, ParserError{Message: err, File: filename})
	}

	if program == nil {
		return ctx, nil
	}

	a := &Analyzer{
		program:  program,
		filename: filename,
		input:    input,
		context:  ctx,
	}

	a.analyze()

	return ctx, nil
}

func analyzeStatistics(ctx *Context, input string) {
	lines := strings.Split(input, "\n")
	ctx.Statistics.TotalLines = len(lines)
	for _, line := range lines {
		trim := strings.TrimSpace(line)
		if trim == "" {
			ctx.Statistics.BlankLines++
		} else if strings.HasPrefix(trim, "#") {
			ctx.Statistics.CommentLines++
		} else {
			ctx.Statistics.CodeLines++
		}
	}
}

func (a *Analyzer) analyze() {
	// First pass: Global symbols and functions
	for _, stmt := range a.program.Statements {
		a.analyzeTopLevel(stmt)
	}
	// Calculate complexity summary
	a.context.Complexity.LinesOfCode = a.context.Statistics.CodeLines
}

func (a *Analyzer) analyzeTopLevel(stmt parser.Statement) {
	switch s := stmt.(type) {
	case *parser.ExpressionStatement:
		// Check for function literal: `fungsi name(...) ...`
		if fn, ok := s.Expression.(*parser.FunctionLiteral); ok {
			a.analyzeFunction(fn)
		} else {
			// Walk top level expressions (scripts)
			a.walkExpression(s.Expression, func(node parser.Node) {})
		}
	case *parser.AssignmentStatement:
		// Global variable
		name := s.Name.Value
		inferredType := "unknown"

		// Simple type inference for literals
		switch s.Value.(type) {
		case *parser.IntegerLiteral:
			inferredType = "integer"
		case *parser.StringLiteral:
			inferredType = "string"
		case *parser.BooleanLiteral:
			inferredType = "boolean"
		}

		a.context.GlobalVars[name] = &Variable{
			Line: s.Token.Line,
			Type: inferredType,
		}
		a.walkExpression(s.Value, func(node parser.Node) {})
	}
}

func (a *Analyzer) analyzeFunction(fn *parser.FunctionLiteral) {
	name := fn.Name
	if name == "" {
		name = "<anonymous>"
	}

	sym := &Symbol{
		Type:       "function",
		Line:       fn.Token.Line,
		Column:     fn.Token.Column,
		Parameters: []Parameter{},
		LocalVars:  []string{},
		Calls:      []string{},
	}

	for _, p := range fn.Parameters {
		sym.Parameters = append(sym.Parameters, Parameter{
			Name:         p.Value,
			InferredType: "any",
			Line:         p.Token.Line,
			Column:       p.Token.Column,
		})
	}

	a.context.Symbols[name] = sym
	a.context.CallGraph[name] = []string{}
	a.context.LocalScopes[name] = make(LocalScope)

	prevFunc := a.currFunc
	a.currFunc = name

	canError := false
	complexity := 1 // Base complexity

	a.walkBlock(fn.Body, func(node parser.Node) {
		// Check for calls
		if call, ok := node.(*parser.CallExpression); ok {
			if ident, ok := call.Function.(*parser.Identifier); ok {
				sym.Calls = append(sym.Calls, ident.Value)
				a.context.CallGraph[name] = append(a.context.CallGraph[name], ident.Value)
			}
		}
		// Check for returns galat
		if ret, ok := node.(*parser.ReturnStatement); ok {
			if ret.ReturnValue != nil {
				if call, ok := ret.ReturnValue.(*parser.CallExpression); ok {
					if ident, ok := call.Function.(*parser.Identifier); ok {
						if ident.Value == "galat" {
							canError = true
						}
					}
				}
			}
		}
		// Local vars
		if assign, ok := node.(*parser.AssignmentStatement); ok {
			varName := assign.Name.Value
			// Avoid duplicates
			found := false
			for _, v := range sym.LocalVars {
				if v == varName {
					found = true
					break
				}
			}
			if !found {
				sym.LocalVars = append(sym.LocalVars, varName)
				a.context.LocalScopes[name][varName] = &Variable{
					Line: assign.Token.Line,
					Type: "inferred",
				}
			}
		}
		// Complexity
		if _, ok := node.(*parser.IfExpression); ok {
			complexity++
		}
		if _, ok := node.(*parser.WhileExpression); ok {
			complexity++
		}
	})

	sym.CanError = canError
	if canError {
		sym.Returns = &TypeInfo{Type: "union", Types: []string{"any", "error"}}
	}

	if complexity > a.context.Complexity.Cyclomatic {
		a.context.Complexity.Cyclomatic += complexity
	}

	a.currFunc = prevFunc
	a.context.Complexity.Functions++
}

func (a *Analyzer) walkBlock(block *parser.BlockStatement, visitor func(parser.Node)) {
	if block == nil {
		return
	}
	for _, stmt := range block.Statements {
		visitor(stmt)
		a.walkStatement(stmt, visitor)
	}
}

func (a *Analyzer) walkStatement(stmt parser.Statement, visitor func(parser.Node)) {
	switch s := stmt.(type) {
	case *parser.ExpressionStatement:
		a.walkExpression(s.Expression, visitor)
	case *parser.ReturnStatement:
		if s.ReturnValue != nil {
			a.walkExpression(s.ReturnValue, visitor)
		}
	case *parser.AssignmentStatement:
		a.walkExpression(s.Value, visitor)
	}
}

func (a *Analyzer) walkExpression(expr parser.Expression, visitor func(parser.Node)) {
	if expr == nil {
		return
	}
	visitor(expr)

	switch e := expr.(type) {
	case *parser.InfixExpression:
		a.walkExpression(e.Left, visitor)
		a.walkExpression(e.Right, visitor)
	case *parser.PrefixExpression:
		a.walkExpression(e.Right, visitor)
	case *parser.IfExpression:
		a.walkExpression(e.Condition, visitor)
		a.walkBlock(e.Consequence, visitor)
		a.walkBlock(e.Alternative, visitor)
	case *parser.WhileExpression:
		a.walkExpression(e.Condition, visitor)
		a.walkBlock(e.Body, visitor)
	case *parser.CallExpression:
		a.walkExpression(e.Function, visitor)
		for _, arg := range e.Arguments {
			a.walkExpression(arg, visitor)
		}
	case *parser.InterpolatedString:
		for _, part := range e.Parts {
			a.walkExpression(part, visitor)
		}
	case *parser.FunctionLiteral:
		a.analyzeFunction(e)
	}
}
