package object

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"strings"
)

type BuiltinDef struct {
	Name    string
	Builtin *Builtin
}

var Builtins = []BuiltinDef{
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
	{
		"baca_file",
		&Builtin{Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return &Error{Message: fmt.Sprintf("argument mismatch: expected 1, got %d", len(args))}
			}
			path, ok := args[0].(*String)
			if !ok {
				return &Error{Message: fmt.Sprintf("argument to `baca_file` must be STRING, got %s", args[0].Type())}
			}

			content, err := os.ReadFile(path.Value)
			if err != nil {
				return &Error{Message: fmt.Sprintf("baca_file error: %s", err.Error())}
			}
			return &String{Value: string(content)}
		}},
	},
	{
		"tulis_file",
		&Builtin{Fn: func(args ...Object) Object {
			if len(args) != 2 {
				return &Error{Message: fmt.Sprintf("argument mismatch: expected 2, got %d", len(args))}
			}
			path, ok := args[0].(*String)
			if !ok {
				return &Error{Message: fmt.Sprintf("first argument to `tulis_file` must be STRING, got %s", args[0].Type())}
			}
			content, ok := args[1].(*String)
			if !ok {
				return &Error{Message: fmt.Sprintf("second argument to `tulis_file` must be STRING, got %s", args[1].Type())}
			}

			err := os.WriteFile(path.Value, []byte(content.Value), 0644)
			if err != nil {
				return &Error{Message: fmt.Sprintf("tulis_file error: %s", err.Error())}
			}
			return &Null{}
		}},
	},
	{
		"input",
		&Builtin{Fn: func(args ...Object) Object {
			if len(args) > 1 {
				return &Error{Message: fmt.Sprintf("argument mismatch: expected 0 or 1, got %d", len(args))}
			}
			if len(args) == 1 {
				prompt, ok := args[0].(*String)
				if !ok {
					return &Error{Message: fmt.Sprintf("argument to `input` must be STRING, got %s", args[0].Type())}
				}
				fmt.Print(prompt.Value)
			}

			reader := bufio.NewReader(os.Stdin)
			text, err := reader.ReadString('\n')
			if err != nil {
				return &Error{Message: fmt.Sprintf("input error: %s", err.Error())}
			}
			return &String{Value: strings.TrimSpace(text)}
		}},
	},
	{
		"huruf_besar",
		&Builtin{Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return &Error{Message: fmt.Sprintf("argument mismatch: expected 1, got %d", len(args))}
			}
			str, ok := args[0].(*String)
			if !ok {
				return &Error{Message: fmt.Sprintf("argument to `huruf_besar` must be STRING, got %s", args[0].Type())}
			}
			return &String{Value: strings.ToUpper(str.Value)}
		}},
	},
	{
		"huruf_kecil",
		&Builtin{Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return &Error{Message: fmt.Sprintf("argument mismatch: expected 1, got %d", len(args))}
			}
			str, ok := args[0].(*String)
			if !ok {
				return &Error{Message: fmt.Sprintf("argument to `huruf_kecil` must be STRING, got %s", args[0].Type())}
			}
			return &String{Value: strings.ToLower(str.Value)}
		}},
	},
	{
		"pisah",
		&Builtin{Fn: func(args ...Object) Object {
			if len(args) != 2 {
				return &Error{Message: fmt.Sprintf("argument mismatch: expected 2, got %d", len(args))}
			}
			str, ok := args[0].(*String)
			if !ok {
				return &Error{Message: fmt.Sprintf("first argument to `pisah` must be STRING, got %s", args[0].Type())}
			}
			delim, ok := args[1].(*String)
			if !ok {
				return &Error{Message: fmt.Sprintf("second argument to `pisah` must be STRING, got %s", args[1].Type())}
			}

			parts := strings.Split(str.Value, delim.Value)
			elements := make([]Object, len(parts))
			for i, p := range parts {
				elements[i] = &String{Value: p}
			}
			return &Array{Elements: elements}
		}},
	},
	{
		"gabung",
		&Builtin{Fn: func(args ...Object) Object {
			if len(args) != 2 {
				return &Error{Message: fmt.Sprintf("argument mismatch: expected 2, got %d", len(args))}
			}
			arr, ok := args[0].(*Array)
			if !ok {
				return &Error{Message: fmt.Sprintf("first argument to `gabung` must be ARRAY, got %s", args[0].Type())}
			}
			delim, ok := args[1].(*String)
			if !ok {
				return &Error{Message: fmt.Sprintf("second argument to `gabung` must be STRING, got %s", args[1].Type())}
			}

			parts := make([]string, len(arr.Elements))
			for i, el := range arr.Elements {
				str, ok := el.(*String)
				if !ok {
					return &Error{Message: fmt.Sprintf("array elements must be STRING, got %s", el.Type())}
				}
				parts[i] = str.Value
			}
			return &String{Value: strings.Join(parts, delim.Value)}
		}},
	},
	{
		"abs",
		&Builtin{Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return &Error{Message: fmt.Sprintf("argument mismatch: expected 1, got %d", len(args))}
			}
			num, ok := args[0].(*Integer)
			if !ok {
				return &Error{Message: fmt.Sprintf("argument to `abs` must be INTEGER, got %s", args[0].Type())}
			}
			if num.Value < 0 {
				return &Integer{Value: -num.Value}
			}
			return num
		}},
	},
	{
		"max",
		&Builtin{Fn: func(args ...Object) Object {
			if len(args) != 2 {
				return &Error{Message: fmt.Sprintf("argument mismatch: expected 2, got %d", len(args))}
			}
			a, ok := args[0].(*Integer)
			if !ok {
				return &Error{Message: fmt.Sprintf("first argument to `max` must be INTEGER, got %s", args[0].Type())}
			}
			b, ok := args[1].(*Integer)
			if !ok {
				return &Error{Message: fmt.Sprintf("second argument to `max` must be INTEGER, got %s", args[1].Type())}
			}
			if a.Value > b.Value {
				return a
			}
			return b
		}},
	},
	{
		"min",
		&Builtin{Fn: func(args ...Object) Object {
			if len(args) != 2 {
				return &Error{Message: fmt.Sprintf("argument mismatch: expected 2, got %d", len(args))}
			}
			a, ok := args[0].(*Integer)
			if !ok {
				return &Error{Message: fmt.Sprintf("first argument to `min` must be INTEGER, got %s", args[0].Type())}
			}
			b, ok := args[1].(*Integer)
			if !ok {
				return &Error{Message: fmt.Sprintf("second argument to `min` must be INTEGER, got %s", args[1].Type())}
			}
			if a.Value < b.Value {
				return a
			}
			return b
		}},
	},
	{
		"pow",
		&Builtin{Fn: func(args ...Object) Object {
			if len(args) != 2 {
				return &Error{Message: fmt.Sprintf("argument mismatch: expected 2, got %d", len(args))}
			}
			x, ok := args[0].(*Integer)
			if !ok {
				return &Error{Message: fmt.Sprintf("first argument to `pow` must be INTEGER, got %s", args[0].Type())}
			}
			y, ok := args[1].(*Integer)
			if !ok {
				return &Error{Message: fmt.Sprintf("second argument to `pow` must be INTEGER, got %s", args[1].Type())}
			}
			res := math.Pow(float64(x.Value), float64(y.Value))
			return &Integer{Value: int64(res)}
		}},
	},
	{
		"sqrt",
		&Builtin{Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return &Error{Message: fmt.Sprintf("argument mismatch: expected 1, got %d", len(args))}
			}
			x, ok := args[0].(*Integer)
			if !ok {
				return &Error{Message: fmt.Sprintf("argument to `sqrt` must be INTEGER, got %s", args[0].Type())}
			}
			if x.Value < 0 {
				return &Error{Message: "cannot calculate square root of negative number"}
			}
			res := math.Sqrt(float64(x.Value))
			return &Integer{Value: int64(res)}
		}},
	},
}

// RegisterBuiltin registers a new builtin function dynamically.
// Useful for adding builtins from other files (e.g. system info).
func RegisterBuiltin(name string, fn BuiltinFunction) {
	Builtins = append(Builtins, BuiltinDef{
		Name:    name,
		Builtin: &Builtin{Fn: fn},
	})
}

func GetBuiltinByName(name string) int {
	for i, b := range Builtins {
		if b.Name == name {
			return i
		}
	}
	return -1
}
