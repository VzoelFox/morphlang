package evaluator

import (
	"fmt"

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
	case *parser.BlockStatement:
		return evalBlockStatement(node, env)
	case *parser.ReturnStatement:
		val := Eval(node.ReturnValue, env)
		if isError(val) {
			return val
		}
		return &object.ReturnValue{Value: val}
	case *parser.AssignmentStatement:
		val := Eval(node.Value, env)
		// Error as Value: we allow assigning Error objects to variables
		env.Set(node.Name.Value, val)

		// We return NULL to indicate the statement executed successfully (even if the assigned value is an error).
		// This prevents the main evaluation loop from aborting execution when a variable is assigned an error value,
		// allowing the user to check the variable with `adalah_galat(x)`.
		return NULL

	// Expressions
	case *parser.IntegerLiteral:
		return &object.Integer{Value: node.Value}
	case *parser.StringLiteral:
		return &object.String{Value: node.Value}
	case *parser.BooleanLiteral:
		return nativeBoolToBooleanObject(node.Value)
	case *parser.PrefixExpression:
		right := Eval(node.Right, env)
		if isError(right) {
			return right
		}
		return evalPrefixExpression(node.Operator, right)
	case *parser.InfixExpression:
		left := Eval(node.Left, env)
		if isError(left) {
			return left
		}
		right := Eval(node.Right, env)
		if isError(right) {
			return right
		}
		return evalInfixExpression(node.Operator, left, right)
	case *parser.IfExpression:
		return evalIfExpression(node, env)
	case *parser.WhileExpression:
		return evalWhileExpression(node, env)
	case *parser.Identifier:
		return evalIdentifier(node, env)
	case *parser.FunctionLiteral:
		params := node.Parameters
		body := node.Body
		fn := &object.Function{Parameters: params, Env: env, Body: body}
		if node.Name != "" {
			env.Set(node.Name, fn)
		}
		return fn
	case *parser.CallExpression:
		function := Eval(node.Function, env)
		if isError(function) {
			return function
		}
		args := evalExpressions(node.Arguments, env)
		// We REMOVE the check that aborts if an arg is an error.
		// This allows functions (like adalah_galat) to receive Error objects.
		return applyFunction(function, args)
	}
	return nil
}

func evalExpressions(exps []parser.Expression, env *object.Environment) []object.Object {
	var result []object.Object

	for _, e := range exps {
		evaluated := Eval(e, env)
		if isError(evaluated) {
			return []object.Object{evaluated}
		}
		result = append(result, evaluated)
	}

	return result
}

func applyFunction(fn object.Object, args []object.Object) object.Object {
	// Check for propagated error in arguments
	if len(args) == 1 && isError(args[0]) {
		// If the function is a built-in that specifically handles errors, let it pass.
		// Otherwise, propagate the error up.
		if builtin, ok := fn.(*object.Builtin); ok {
			// List of built-ins that accept errors
			// TODO: Maybe add a flag to Builtin struct?
			// For now, hardcode check or rely on Builtin logic?
			// But Builtins are stored in map, we don't know their name here easily unless we look it up.
			// Actually, Builtin struct only has Fn.
			// We can try to run it. But if it's `panjang(error)`, `panjang` implementation checks type.
			// If `panjang` doesn't support error, it returns error.
			// But what if it's a User Function?
			return builtin.Fn(args...)
		}

		// For user functions, we don't support passing Error objects as arguments yet
		// (unless we add specific syntax for it, but AGENTS.md implies we check return values).
		// So we propagate the error.
		return args[0]
	}

	switch function := fn.(type) {
	case *object.Function:
		// Safety check for argument count
		if len(args) != len(function.Parameters) {
			return newError(nil, "argument mismatch: expected %d, got %d", len(function.Parameters), len(args))
		}
		extendedEnv := extendFunctionEnv(function, args)
		evaluated := Eval(function.Body, extendedEnv)
		return unwrapReturnValue(evaluated)
	case *object.Builtin:
		return function.Fn(args...)
	default:
		return newError(nil, "not a function: %s", fn.Type())
	}
}

func extendFunctionEnv(fn *object.Function, args []object.Object) *object.Environment {
	env := object.NewEnclosedEnvironment(fn.Env)

	for i, param := range fn.Parameters {
		env.Set(param.Value, args[i])
	}

	return env
}

func unwrapReturnValue(obj object.Object) object.Object {
	if returnValue, ok := obj.(*object.ReturnValue); ok {
		return returnValue.Value
	}

	return obj
}

func evalProgram(program *parser.Program, env *object.Environment) object.Object {
	var result object.Object
	for _, statement := range program.Statements {
		result = Eval(statement, env)

		switch result := result.(type) {
		case *object.ReturnValue:
			return result.Value
		case *object.Error:
			return result
		}
	}
	return result
}

func evalBlockStatement(block *parser.BlockStatement, env *object.Environment) object.Object {
	var result object.Object
	for _, statement := range block.Statements {
		result = Eval(statement, env)

		if result != nil {
			rt := result.Type()
			if rt == object.RETURN_VALUE_OBJ || rt == object.ERROR_OBJ {
				return result
			}
		}
	}
	return result
}

func evalPrefixExpression(operator string, right object.Object) object.Object {
	switch operator {
	case "!":
		return evalBangOperatorExpression(right)
	case "-":
		return evalMinusPrefixOperatorExpression(right)
	default:
		return newError(nil, "unknown operator: %s%s", operator, right.Type())
	}
}

func evalInfixExpression(operator string, left, right object.Object) object.Object {
	switch {
	case left.Type() == object.INTEGER_OBJ && right.Type() == object.INTEGER_OBJ:
		return evalIntegerInfixExpression(operator, left, right)
	case operator == "==":
		return nativeBoolToBooleanObject(left == right)
	case operator == "!=":
		return nativeBoolToBooleanObject(left != right)
	case left.Type() != right.Type():
		return newError(nil, "type mismatch: %s %s %s", left.Type(), operator, right.Type())
	default:
		return newError(nil, "unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

func evalIntegerInfixExpression(operator string, left, right object.Object) object.Object {
	leftVal := left.(*object.Integer).Value
	rightVal := right.(*object.Integer).Value

	switch operator {
	case "+":
		return &object.Integer{Value: leftVal + rightVal}
	case "-":
		return &object.Integer{Value: leftVal - rightVal}
	case "*":
		return &object.Integer{Value: leftVal * rightVal}
	case "/":
		if rightVal == 0 {
			return newError(nil, "division by zero")
		}
		return &object.Integer{Value: leftVal / rightVal}
	case "<":
		return nativeBoolToBooleanObject(leftVal < rightVal)
	case ">":
		return nativeBoolToBooleanObject(leftVal > rightVal)
	case "==":
		return nativeBoolToBooleanObject(leftVal == rightVal)
	case "!=":
		return nativeBoolToBooleanObject(leftVal != rightVal)
	default:
		return newError(nil, "unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

func evalBangOperatorExpression(right object.Object) object.Object {
	switch right {
	case TRUE:
		return FALSE
	case FALSE:
		return TRUE
	case NULL:
		return TRUE
	default:
		return FALSE
	}
}

func evalMinusPrefixOperatorExpression(right object.Object) object.Object {
	if right.Type() != object.INTEGER_OBJ {
		return newError(nil, "unknown operator: -%s", right.Type())
	}

	value := right.(*object.Integer).Value
	return &object.Integer{Value: -value}
}

func evalIdentifier(node *parser.Identifier, env *object.Environment) object.Object {
	val, ok := env.Get(node.Value)
	if ok {
		return val
	}

	if builtin, ok := builtins[node.Value]; ok {
		return builtin
	}

	return newError(node, "identifier not found: %s", node.Value)
}

func evalIfExpression(ie *parser.IfExpression, env *object.Environment) object.Object {
	condition := Eval(ie.Condition, env)
	if isError(condition) {
		return condition
	}

	if isTruthy(condition) {
		return Eval(ie.Consequence, env)
	} else if ie.Alternative != nil {
		return Eval(ie.Alternative, env)
	} else {
		return NULL
	}
}

func evalWhileExpression(we *parser.WhileExpression, env *object.Environment) object.Object {
	var result object.Object = NULL

	for {
		condition := Eval(we.Condition, env)
		if isError(condition) {
			return condition
		}

		if !isTruthy(condition) {
			break
		}

		result = Eval(we.Body, env)
		if result != nil && (result.Type() == object.RETURN_VALUE_OBJ || result.Type() == object.ERROR_OBJ) {
			return result
		}
	}
	return result
}

func isTruthy(obj object.Object) bool {
	switch obj {
	case NULL:
		return false
	case TRUE:
		return true
	case FALSE:
		return false
	default:
		return true
	}
}

func nativeBoolToBooleanObject(input bool) *object.Boolean {
	if input {
		return TRUE
	}
	return FALSE
}

func newError(node parser.Node, format string, a ...interface{}) *object.Error {
	msg := fmt.Sprintf(format, a...)
	err := &object.Error{
		Message: msg,
		File:    "unknown", // Defaults, ideally we propagate file context too
		Line:    0,
		Column:  0,
	}

	if node != nil {
		// Try to extract Token from known types or if we extended Node interface
		// For now we check specifically for types we know have Tokens in AST
		switch n := node.(type) {
		case *parser.Identifier:
			err.Line = n.Token.Line
			err.Column = n.Token.Column
		case *parser.IntegerLiteral:
			err.Line = n.Token.Line
			err.Column = n.Token.Column
		case *parser.BooleanLiteral:
			err.Line = n.Token.Line
			err.Column = n.Token.Column
		case *parser.PrefixExpression:
			err.Line = n.Token.Line
			err.Column = n.Token.Column
		case *parser.InfixExpression:
			err.Line = n.Token.Line
			err.Column = n.Token.Column
		}
	}

	return err
}

func isError(obj object.Object) bool {
	if obj != nil {
		return obj.Type() == object.ERROR_OBJ
	}
	return false
}
