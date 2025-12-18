package evaluator

import (
	"fmt"

	"github.com/VzoelFox/morphlang/pkg/object"
	"github.com/VzoelFox/morphlang/pkg/parser"
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

		switch name := node.Name.(type) {
		case *parser.Identifier:
			env.Set(name.Value, val)
		default:
			return newError(node.Name, "assignment not supported in evaluator for %T", node.Name)
		}

		// We return NULL to indicate the statement executed successfully
		return object.NewNull()

	// Expressions
	case *parser.IntegerLiteral:
		return object.NewInteger(node.Value)
	case *parser.StringLiteral:
		return object.NewString(node.Value)
	case *parser.BooleanLiteral:
		return nativeBoolToBooleanObject(node.Value)
	case *parser.PrefixExpression:
		right := Eval(node.Right, env)
		if isError(right) {
			return right
		}
		return evalPrefixExpression(node, node.Operator, right)
	case *parser.InfixExpression:
		left := Eval(node.Left, env)
		if isError(left) {
			return left
		}
		right := Eval(node.Right, env)
		if isError(right) {
			return right
		}
		return evalInfixExpression(node, node.Operator, left, right)
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
		if builtin, ok := fn.(*object.Builtin); ok {
			return builtin.Fn(args...)
		}
		return args[0]
	}

	switch function := fn.(type) {
	case *object.Function:
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

func evalPrefixExpression(node parser.Node, operator string, right object.Object) object.Object {
	switch operator {
	case "!":
		return evalBangOperatorExpression(right)
	case "-":
		return evalMinusPrefixOperatorExpression(right, node)
	default:
		return newError(node, "unknown operator: %s%s", operator, right.Type())
	}
}

func evalInfixExpression(node parser.Node, operator string, left, right object.Object) object.Object {
	switch {
	case left.Type() == object.INTEGER_OBJ && right.Type() == object.INTEGER_OBJ:
		return evalIntegerInfixExpression(node, operator, left, right)
	case left.Type() == object.STRING_OBJ && right.Type() == object.STRING_OBJ:
		return evalStringInfixExpression(node, operator, left, right)
	case operator == "==":
		// Basic pointer equality fallback, but better to compare values if possible
		// For primitives, we should compare values.
		// However, nativeBoolToBooleanObject expects bool.
		// If they are pointers to different addresses but same value?
		// We should implement object.Equals?
		// For now, assume strict check:
		if left.Type() != right.Type() {
			return object.NewBoolean(false)
		}
		// If Integer/String/Boolean, check value
		switch l := left.(type) {
		case *object.Integer:
			r := right.(*object.Integer)
			return object.NewBoolean(l.GetValue() == r.GetValue())
		case *object.Boolean:
			r := right.(*object.Boolean)
			return object.NewBoolean(l.GetValue() == r.GetValue())
		case *object.String:
			r := right.(*object.String)
			return object.NewBoolean(l.GetValue() == r.GetValue())
		case *object.Null:
			return object.NewBoolean(true)
		default:
			return object.NewBoolean(left == right)
		}
	case operator == "!=":
		// Similar logic inverted
		if left.Type() != right.Type() {
			return object.NewBoolean(true)
		}
		switch l := left.(type) {
		case *object.Integer:
			r := right.(*object.Integer)
			return object.NewBoolean(l.GetValue() != r.GetValue())
		case *object.Boolean:
			r := right.(*object.Boolean)
			return object.NewBoolean(l.GetValue() != r.GetValue())
		case *object.String:
			r := right.(*object.String)
			return object.NewBoolean(l.GetValue() != r.GetValue())
		case *object.Null:
			return object.NewBoolean(false)
		default:
			return object.NewBoolean(left != right)
		}
	case left.Type() != right.Type():
		return newError(node, "type mismatch: %s %s %s", left.Type(), operator, right.Type())
	default:
		return newError(node, "unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

func evalIntegerInfixExpression(node parser.Node, operator string, left, right object.Object) object.Object {
	leftVal := left.(*object.Integer).GetValue()
	rightVal := right.(*object.Integer).GetValue()

	switch operator {
	case "+":
		return object.NewInteger(leftVal + rightVal)
	case "-":
		return object.NewInteger(leftVal - rightVal)
	case "*":
		return object.NewInteger(leftVal * rightVal)
	case "/":
		if rightVal == 0 {
			return newError(node, "division by zero")
		}
		return object.NewInteger(leftVal / rightVal)
	case "<":
		return nativeBoolToBooleanObject(leftVal < rightVal)
	case ">":
		return nativeBoolToBooleanObject(leftVal > rightVal)
	case "==":
		return nativeBoolToBooleanObject(leftVal == rightVal)
	case "!=":
		return nativeBoolToBooleanObject(leftVal != rightVal)
	default:
		return newError(node, "unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

func evalStringInfixExpression(node parser.Node, operator string, left, right object.Object) object.Object {
	leftVal := left.(*object.String).GetValue()
	rightVal := right.(*object.String).GetValue()

	switch operator {
	case "+":
		return object.NewString(leftVal + rightVal)
	case "==":
		return nativeBoolToBooleanObject(leftVal == rightVal)
	case "!=":
		return nativeBoolToBooleanObject(leftVal != rightVal)
	default:
		return newError(node, "unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

func evalBangOperatorExpression(right object.Object) object.Object {
	// !true -> false
	// !false -> true
	// !null -> true
	// !5 -> false
	if isTruthy(right) {
		return object.NewBoolean(false)
	}
	return object.NewBoolean(true)
}

func evalMinusPrefixOperatorExpression(right object.Object, node parser.Node) object.Object {
	if right.Type() != object.INTEGER_OBJ {
		return newError(node, "unknown operator: -%s", right.Type())
	}

	value := right.(*object.Integer).GetValue()
	return object.NewInteger(-value)
}

func evalIdentifier(node *parser.Identifier, env *object.Environment) object.Object {
	val, ok := env.Get(node.Value)
	if ok {
		return val
	}

	if index := object.GetBuiltinByName(node.Value); index != -1 {
		return object.Builtins[index].Builtin
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
		return object.NewNull()
	}
}

func evalWhileExpression(we *parser.WhileExpression, env *object.Environment) object.Object {
	var result object.Object = object.NewNull()

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
	switch obj := obj.(type) {
	case *object.Null:
		return false
	case *object.Boolean:
		return obj.GetValue()
	default:
		return true
	}
}

func nativeBoolToBooleanObject(input bool) *object.Boolean {
	return object.NewBoolean(input)
}

func newError(node parser.Node, format string, a ...interface{}) *object.Error {
	msg := fmt.Sprintf(format, a...)
	err := &object.Error{
		Message: msg,
		File:    "unknown",
		Line:    0,
		Column:  0,
	}

	if node != nil {
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
