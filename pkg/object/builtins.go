package object

import "fmt"

var Builtins = []struct {
	Name    string
	Builtin *Builtin
}{
	{
		"panjang",
		&Builtin{Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return &Error{Message: fmt.Sprintf("argument mismatch: expected 1, got %d", len(args))}
			}
			switch arg := args[0].(type) {
			case *Array:
				return &Integer{Value: int64(len(arg.Elements))}
			case *String:
				return &Integer{Value: int64(len(arg.Value))}
			default:
				return &Error{Message: fmt.Sprintf("argument to `panjang` not supported, got %s", args[0].Type())}
			}
		}},
	},
	{
		"cetak",
		&Builtin{Fn: func(args ...Object) Object {
			for _, arg := range args {
				fmt.Println(arg.Inspect())
			}
			return &Null{}
		}},
	},
	{
		"tipe",
		&Builtin{Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return &Error{Message: fmt.Sprintf("argument mismatch: expected 1, got %d", len(args))}
			}
			return &String{Value: string(args[0].Type())}
		}},
	},
	{
		"galat",
		&Builtin{Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return &Error{Message: fmt.Sprintf("argument mismatch: expected 1, got %d", len(args))}
			}
			switch arg := args[0].(type) {
			case *String:
				return &Error{Message: arg.Value}
			default:
				return &Error{Message: fmt.Sprintf("argument to `galat` must be STRING, got %s", args[0].Type())}
			}
		}},
	},
	{
		"adalah_galat",
		&Builtin{Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return &Error{Message: fmt.Sprintf("argument mismatch: expected 1, got %d", len(args))}
			}
			switch args[0].(type) {
			case *Error:
				return &Boolean{Value: true}
			default:
				return &Boolean{Value: false}
			}
		}},
	},
	{
		"pesan_galat",
		&Builtin{Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return &Error{Message: fmt.Sprintf("argument mismatch: expected 1, got %d", len(args))}
			}
			switch arg := args[0].(type) {
			case *Error:
				return &String{Value: arg.Message}
			default:
				return &Error{Message: fmt.Sprintf("argument to `pesan_galat` must be ERROR, got %s", args[0].Type())}
			}
		}},
	},
}

func GetBuiltinByName(name string) int {
	for i, b := range Builtins {
		if b.Name == name {
			return i
		}
	}
	return -1
}
