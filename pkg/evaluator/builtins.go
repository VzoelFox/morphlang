package evaluator

import (
	"fmt"
	"github.com/VzoelFox/morphlang/pkg/object"
)

var builtins = map[string]*object.Builtin{
	"panjang": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError(nil, "argument mismatch: expected 1, got %d", len(args))
			}

			if args[0] == nil {
				return newError(nil, "argument is nil")
			}

			switch arg := args[0].(type) {
			case *object.String:
				return &object.Integer{Value: int64(len(arg.Value))}
			default:
				// Debug print
				// fmt.Printf("Arg type: %T, Value: %+v\n", args[0], args[0])
				return newError(nil, "argument to `panjang` not supported, got %s", args[0].Type())
			}
		},
	},
	"cetak": {
		Fn: func(args ...object.Object) object.Object {
			for _, arg := range args {
				fmt.Println(arg.Inspect())
			}
			return NULL
		},
	},
	"tipe": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError(nil, "argument mismatch: expected 1, got %d", len(args))
			}
			return &object.String{Value: string(args[0].Type())}
		},
	},
	// Error handling built-ins
	"galat": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError(nil, "argument mismatch: expected 1, got %d", len(args))
			}

			message, ok := args[0].(*object.String)
			if !ok {
				return newError(nil, "argument to `galat` must be STRING, got %s", args[0].Type())
			}

			// Note: Context (line/column) is not available here easily without passing context to builtins
			// For now, we create a basic error.
			return &object.Error{
				Message: message.Value,
				File: "runtime",
			}
		},
	},
	"adalah_galat": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError(nil, "argument mismatch: expected 1, got %d", len(args))
			}

			if args[0].Type() == object.ERROR_OBJ {
				return TRUE
			}
			return FALSE
		},
	},
	"pesan_galat": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError(nil, "argument mismatch: expected 1, got %d", len(args))
			}

			if err, ok := args[0].(*object.Error); ok {
				return &object.String{Value: err.Message}
			}
			return NULL
		},
	},
}
