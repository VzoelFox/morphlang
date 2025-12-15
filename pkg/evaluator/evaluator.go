package evaluator

import (
	"github.com/VzoelFox/morphlang/pkg/object"
	"github.com/VzoelFox/morphlang/pkg/parser"
)

var (
	NULL  = &object.Null{}
	TRUE  = &object.Boolean{Value: true}
	FALSE = &object.Boolean{Value: false}
)

func Eval(node parser.Node, env *object.Environment) object.Object {
	switch node := node.(type) {
	// Statements
	case *parser.Program:
		return evalProgram(node, env)
	case *parser.ExpressionStatement:
		return Eval(node.Expression, env)

	// Expressions
	case *parser.IntegerLiteral:
		return &object.Integer{Value: node.Value}
	case *parser.BooleanLiteral:
		return nativeBoolToBooleanObject(node.Value)
	}
	return nil
}

func evalProgram(program *parser.Program, env *object.Environment) object.Object {
	var result object.Object
	for _, statement := range program.Statements {
		result = Eval(statement, env)

		// Handle return value explicitly later
		if returnValue, ok := result.(*object.ReturnValue); ok {
			return returnValue.Value
		}
		if err, ok := result.(*object.Error); ok {
			return err
		}
	}
	return result
}

func nativeBoolToBooleanObject(input bool) *object.Boolean {
	if input {
		return TRUE
	}
	return FALSE
}
