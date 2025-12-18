package object

import (
	"bufio"
	"fmt"
	"github.com/VzoelFox/morphlang/pkg/memory"
	"io"
	"math"
	"os"
	"strings"
	"time"
)

// Helper Functions for Standardized Errors

func newArgumentError(got int, expected int) *Error {
	if got < expected {
		return &Error{Code: ErrCodeMissingArgs, Message: fmt.Sprintf("argument mismatch: expected %d, got %d", expected, got)}
	}
	return &Error{Code: ErrCodeTooManyArgs, Message: fmt.Sprintf("argument mismatch: expected %d, got %d", expected, got)}
}

func newArgumentErrorRange(got int, min int, max int) *Error {
	if got < min {
		return &Error{Code: ErrCodeMissingArgs, Message: fmt.Sprintf("argument mismatch: expected %d or %d, got %d", min, max, got)}
	}
	return &Error{Code: ErrCodeTooManyArgs, Message: fmt.Sprintf("argument mismatch: expected %d or %d, got %d", min, max, got)}
}

type BuiltinDef struct {
	Name    string
	Builtin *Builtin
}

var Builtins = []BuiltinDef{
	{
		"panjang",
		&Builtin{Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newArgumentError(len(args), 1)
			}
			switch arg := args[0].(type) {
			case *Array:
				return &Integer{Value: int64(len(arg.Elements))}
			case *String:
				return &Integer{Value: int64(len(arg.Value))}
			default:
				return &Error{Code: ErrCodeTypeMismatch, Message: fmt.Sprintf("argument to `panjang` not supported, got %s", args[0].Type())}
			}
		}},
	},
	{
		"kunci",
		&Builtin{Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newArgumentError(len(args), 1)
			}
			hash, ok := args[0].(*Hash)
			if !ok {
				return &Error{Code: ErrCodeTypeMismatch, Message: fmt.Sprintf("argument to `kunci` must be HASH, got %s", args[0].Type())}
			}

			elements := make([]Object, 0, len(hash.Pairs))
			for _, pair := range hash.Pairs {
				elements = append(elements, pair.Key)
			}
			return &Array{Elements: elements}
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
				return newArgumentError(len(args), 1)
			}
			return &String{Value: string(args[0].Type())}
		}},
	},
	{
		"galat",
		&Builtin{Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newArgumentError(len(args), 1)
			}
			switch arg := args[0].(type) {
			case *String:
				return &Error{Message: arg.Value}
			default:
				return &Error{Code: ErrCodeTypeMismatch, Message: fmt.Sprintf("argument to `galat` must be STRING, got %s", args[0].Type())}
			}
		}},
	},
	{
		"adalah_galat",
		&Builtin{Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newArgumentError(len(args), 1)
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
				return newArgumentError(len(args), 1)
			}
			switch arg := args[0].(type) {
			case *Error:
				return &String{Value: arg.Message}
			default:
				return &Error{Code: ErrCodeTypeMismatch, Message: fmt.Sprintf("argument to `pesan_galat` must be ERROR, got %s", args[0].Type())}
			}
		}},
	},
	{
		"baca_file",
		&Builtin{Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newArgumentError(len(args), 1)
			}
			path, ok := args[0].(*String)
			if !ok {
				return &Error{Code: ErrCodeTypeMismatch, Message: fmt.Sprintf("argument to `baca_file` must be STRING, got %s", args[0].Type())}
			}

			content, err := os.ReadFile(path.Value)
			if err != nil {
				return &Error{Code: ErrCodeRuntime, Message: fmt.Sprintf("baca_file error: %s", err.Error())}
			}
			return &String{Value: string(content)}
		}},
	},
	{
		"tulis_file",
		&Builtin{Fn: func(args ...Object) Object {
			if len(args) != 2 {
				return newArgumentError(len(args), 2)
			}
			path, ok := args[0].(*String)
			if !ok {
				return &Error{Code: ErrCodeTypeMismatch, Message: fmt.Sprintf("first argument to `tulis_file` must be STRING, got %s", args[0].Type())}
			}
			content, ok := args[1].(*String)
			if !ok {
				return &Error{Code: ErrCodeTypeMismatch, Message: fmt.Sprintf("second argument to `tulis_file` must be STRING, got %s", args[1].Type())}
			}

			err := os.WriteFile(path.Value, []byte(content.Value), 0644)
			if err != nil {
				return &Error{Code: ErrCodeRuntime, Message: fmt.Sprintf("tulis_file error: %s", err.Error())}
			}
			return &Null{}
		}},
	},
	{
		"buka_file",
		&Builtin{Fn: func(args ...Object) Object {
			if len(args) != 2 {
				return newArgumentError(len(args), 2)
			}
			path, ok := args[0].(*String)
			if !ok {
				return &Error{Code: ErrCodeTypeMismatch, Message: fmt.Sprintf("first argument to `buka_file` must be STRING, got %s", args[0].Type())}
			}
			mode, ok := args[1].(*String)
			if !ok {
				return &Error{Code: ErrCodeTypeMismatch, Message: fmt.Sprintf("second argument to `buka_file` must be STRING, got %s", args[1].Type())}
			}

			var flag int
			switch mode.Value {
			case "b": // Baca (Read)
				flag = os.O_RDONLY
			case "t": // Tulis (Write/Truncate)
				flag = os.O_WRONLY | os.O_CREATE | os.O_TRUNC
			case "tb": // Tulis Baru (Truncate - same as t)
				flag = os.O_WRONLY | os.O_CREATE | os.O_TRUNC
			case "t+": // Tulis Tambah (Append)
				flag = os.O_WRONLY | os.O_CREATE | os.O_APPEND
			default:
				return &Error{Code: ErrCodeTypeMismatch, Message: "unknown file mode: " + mode.Value}
			}

			f, err := os.OpenFile(path.Value, flag, 0644)
			if err != nil {
				return &Error{Code: ErrCodeRuntime, Message: "failed to open file: " + err.Error()}
			}
			return &File{File: f, Mode: mode.Value}
		}},
	},
	{
		"tutup_file",
		&Builtin{Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newArgumentError(len(args), 1)
			}
			f, ok := args[0].(*File)
			if !ok {
				return &Error{Code: ErrCodeTypeMismatch, Message: fmt.Sprintf("argument to `tutup_file` must be FILE, got %s", args[0].Type())}
			}
			err := f.File.Close()
			if err != nil {
				return &Error{Code: ErrCodeRuntime, Message: "failed to close file: " + err.Error()}
			}
			return &Null{}
		}},
	},
	{
		"baca",
		&Builtin{Fn: func(args ...Object) Object {
			if len(args) < 1 || len(args) > 2 {
				return newArgumentErrorRange(len(args), 1, 2)
			}
			f, ok := args[0].(*File)
			if !ok {
				return &Error{Code: ErrCodeTypeMismatch, Message: fmt.Sprintf("first argument to `baca` must be FILE, got %s", args[0].Type())}
			}

			var limit int64 = -1
			if len(args) == 2 {
				n, ok := args[1].(*Integer)
				if !ok {
					return &Error{Code: ErrCodeTypeMismatch, Message: fmt.Sprintf("second argument to `baca` must be INTEGER, got %s", args[1].Type())}
				}
				limit = n.Value
			}

			var content []byte
			var err error

			if limit < 0 {
				content, err = io.ReadAll(f.File)
			} else {
				content = make([]byte, limit)
				var n int
				n, err = f.File.Read(content)
				if n < int(limit) {
					content = content[:n]
				}
			}

			if err != nil && err != io.EOF {
				return &Error{Code: ErrCodeRuntime, Message: "baca error: " + err.Error()}
			}
			return &String{Value: string(content)}
		}},
	},
	{
		"tulis",
		&Builtin{Fn: func(args ...Object) Object {
			if len(args) != 2 {
				return newArgumentError(len(args), 2)
			}
			f, ok := args[0].(*File)
			if !ok {
				return &Error{Code: ErrCodeTypeMismatch, Message: fmt.Sprintf("first argument to `tulis` must be FILE, got %s", args[0].Type())}
			}

			var content []byte
			if str, ok := args[1].(*String); ok {
				content = []byte(str.Value)
			} else {
				content = []byte(args[1].Inspect())
			}

			_, err := f.File.Write(content)
			if err != nil {
				return &Error{Code: ErrCodeRuntime, Message: "tulis error: " + err.Error()}
			}
			return &Null{}
		}},
	},
	{
		"input",
		&Builtin{Fn: func(args ...Object) Object {
			if len(args) > 1 {
				return newArgumentErrorRange(len(args), 0, 1)
			}
			if len(args) == 1 {
				prompt, ok := args[0].(*String)
				if !ok {
					return &Error{Code: ErrCodeTypeMismatch, Message: fmt.Sprintf("argument to `input` must be STRING, got %s", args[0].Type())}
				}
				fmt.Print(prompt.Value)
			}

			reader := bufio.NewReader(os.Stdin)
			text, err := reader.ReadString('\n')
			if err != nil {
				return &Error{Code: ErrCodeRuntime, Message: fmt.Sprintf("input error: %s", err.Error())}
			}
			return &String{Value: strings.TrimSpace(text)}
		}},
	},
	{
		"huruf_besar",
		&Builtin{Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newArgumentError(len(args), 1)
			}
			str, ok := args[0].(*String)
			if !ok {
				return &Error{Code: ErrCodeTypeMismatch, Message: fmt.Sprintf("argument to `huruf_besar` must be STRING, got %s", args[0].Type())}
			}
			return &String{Value: strings.ToUpper(str.Value)}
		}},
	},
	{
		"huruf_kecil",
		&Builtin{Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newArgumentError(len(args), 1)
			}
			str, ok := args[0].(*String)
			if !ok {
				return &Error{Code: ErrCodeTypeMismatch, Message: fmt.Sprintf("argument to `huruf_kecil` must be STRING, got %s", args[0].Type())}
			}
			return &String{Value: strings.ToLower(str.Value)}
		}},
	},
	{
		"pisah",
		&Builtin{Fn: func(args ...Object) Object {
			if len(args) != 2 {
				return newArgumentError(len(args), 2)
			}
			str, ok := args[0].(*String)
			if !ok {
				return &Error{Code: ErrCodeTypeMismatch, Message: fmt.Sprintf("first argument to `pisah` must be STRING, got %s", args[0].Type())}
			}
			delim, ok := args[1].(*String)
			if !ok {
				return &Error{Code: ErrCodeTypeMismatch, Message: fmt.Sprintf("second argument to `pisah` must be STRING, got %s", args[1].Type())}
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
				return newArgumentError(len(args), 2)
			}
			arr, ok := args[0].(*Array)
			if !ok {
				return &Error{Code: ErrCodeTypeMismatch, Message: fmt.Sprintf("first argument to `gabung` must be ARRAY, got %s", args[0].Type())}
			}
			delim, ok := args[1].(*String)
			if !ok {
				return &Error{Code: ErrCodeTypeMismatch, Message: fmt.Sprintf("second argument to `gabung` must be STRING, got %s", args[1].Type())}
			}

			parts := make([]string, len(arr.Elements))
			for i, el := range arr.Elements {
				str, ok := el.(*String)
				if !ok {
					return &Error{Code: ErrCodeTypeMismatch, Message: fmt.Sprintf("array elements must be STRING, got %s", el.Type())}
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
				return newArgumentError(len(args), 1)
			}
			num, ok := args[0].(*Integer)
			if !ok {
				return &Error{Code: ErrCodeTypeMismatch, Message: fmt.Sprintf("argument to `abs` must be INTEGER, got %s", args[0].Type())}
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
				return newArgumentError(len(args), 2)
			}
			a, ok := args[0].(*Integer)
			if !ok {
				return &Error{Code: ErrCodeTypeMismatch, Message: fmt.Sprintf("first argument to `max` must be INTEGER, got %s", args[0].Type())}
			}
			b, ok := args[1].(*Integer)
			if !ok {
				return &Error{Code: ErrCodeTypeMismatch, Message: fmt.Sprintf("second argument to `max` must be INTEGER, got %s", args[1].Type())}
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
				return newArgumentError(len(args), 2)
			}
			a, ok := args[0].(*Integer)
			if !ok {
				return &Error{Code: ErrCodeTypeMismatch, Message: fmt.Sprintf("first argument to `min` must be INTEGER, got %s", args[0].Type())}
			}
			b, ok := args[1].(*Integer)
			if !ok {
				return &Error{Code: ErrCodeTypeMismatch, Message: fmt.Sprintf("second argument to `min` must be INTEGER, got %s", args[1].Type())}
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
				return newArgumentError(len(args), 2)
			}

			xVal, err := getFloatValue(args[0])
			if err != nil {
				return &Error{Code: ErrCodeTypeMismatch, Message: "first argument to `pow` " + err.Error()}
			}
			yVal, err := getFloatValue(args[1])
			if err != nil {
				return &Error{Code: ErrCodeTypeMismatch, Message: "second argument to `pow` " + err.Error()}
			}

			res := math.Pow(xVal, yVal)
			return &Float{Value: res}
		}},
	},
	{
		"sqrt",
		&Builtin{Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newArgumentError(len(args), 1)
			}

			val, err := getFloatValue(args[0])
			if err != nil {
				return &Error{Code: ErrCodeTypeMismatch, Message: "argument to `sqrt` " + err.Error()}
			}

			if val < 0 {
				return &Error{Code: ErrCodeRuntime, Message: "cannot calculate square root of negative number"}
			}
			res := math.Sqrt(val)
			return &Float{Value: res}
		}},
	},
	{
		"sin",
		&Builtin{Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newArgumentError(len(args), 1)
			}
			val, err := getFloatValue(args[0])
			if err != nil { return &Error{Code: ErrCodeTypeMismatch, Message: "argument to `sin` " + err.Error()} }
			return &Float{Value: math.Sin(val)}
		}},
	},
	{
		"cos",
		&Builtin{Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newArgumentError(len(args), 1)
			}
			val, err := getFloatValue(args[0])
			if err != nil { return &Error{Code: ErrCodeTypeMismatch, Message: "argument to `cos` " + err.Error()} }
			return &Float{Value: math.Cos(val)}
		}},
	},
	{
		"tan",
		&Builtin{Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newArgumentError(len(args), 1)
			}
			val, err := getFloatValue(args[0])
			if err != nil { return &Error{Code: ErrCodeTypeMismatch, Message: "argument to `tan` " + err.Error()} }
			return &Float{Value: math.Tan(val)}
		}},
	},
	{
		"asin",
		&Builtin{Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newArgumentError(len(args), 1)
			}
			val, err := getFloatValue(args[0])
			if err != nil { return &Error{Code: ErrCodeTypeMismatch, Message: "argument to `asin` " + err.Error()} }
			return &Float{Value: math.Asin(val)}
		}},
	},
	{
		"acos",
		&Builtin{Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newArgumentError(len(args), 1)
			}
			val, err := getFloatValue(args[0])
			if err != nil { return &Error{Code: ErrCodeTypeMismatch, Message: "argument to `acos` " + err.Error()} }
			return &Float{Value: math.Acos(val)}
		}},
	},
	{
		"atan",
		&Builtin{Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newArgumentError(len(args), 1)
			}
			val, err := getFloatValue(args[0])
			if err != nil { return &Error{Code: ErrCodeTypeMismatch, Message: "argument to `atan` " + err.Error()} }
			return &Float{Value: math.Atan(val)}
		}},
	},
	{
		"pi",
		&Builtin{Fn: func(args ...Object) Object {
			return &Float{Value: math.Pi}
		}},
	},
}

func getFloatValue(obj Object) (float64, error) {
	switch val := obj.(type) {
	case *Integer:
		return float64(val.Value), nil
	case *Float:
		return val.Value, nil
	default:
		return 0, fmt.Errorf("must be INTEGER or FLOAT, got %s", val.Type())
	}
}

func init() {
	// waktu_sekarang() -> Time
	RegisterBuiltin("waktu_sekarang", func(args ...Object) Object {
		return &Time{Value: time.Now()}
	})

	// waktu_unix() -> Integer
	RegisterBuiltin("waktu_unix", func(args ...Object) Object {
		return &Integer{Value: time.Now().Unix()}
	})

	// tidur(milidetik) -> Null
	RegisterBuiltin("tidur", func(args ...Object) Object {
		if len(args) != 1 {
			return newArgumentError(len(args), 1)
		}

		ms, ok := args[0].(*Integer)
		if !ok {
			return &Error{Code: ErrCodeTypeMismatch, Message: fmt.Sprintf("argument to `tidur` must be INTEGER, got %s", args[0].Type())}
		}

		time.Sleep(time.Duration(ms.Value) * time.Millisecond)
		return &Null{}
	})

	// format_waktu(waktu, format_str) -> String
	RegisterBuiltin("format_waktu", func(args ...Object) Object {
		if len(args) != 2 {
			return newArgumentError(len(args), 2)
		}

		tObj, ok := args[0].(*Time)
		if !ok {
			return &Error{Code: ErrCodeTypeMismatch, Message: fmt.Sprintf("first argument to `format_waktu` must be TIME, got %s", args[0].Type())}
		}

		formatObj, ok := args[1].(*String)
		if !ok {
			return &Error{Code: ErrCodeTypeMismatch, Message: fmt.Sprintf("second argument to `format_waktu` must be STRING, got %s", args[1].Type())}
		}

		layout := formatObj.Value
		switch layout {
		case "RFC3339":
			layout = time.RFC3339
		case "ANSIC":
			layout = time.ANSIC
		case "UnixDate":
			layout = time.UnixDate
		}

		return &String{Value: tObj.Value.Format(layout)}
	})

	// System Signals (Intercepted by VM)
	RegisterBuiltin("potret", func(args ...Object) Object {
		return &Error{Message: "SIGNAL:SNAPSHOT"}
	})

	RegisterBuiltin("pulih", func(args ...Object) Object {
		msg := "Manual Rollback"
		if len(args) > 0 {
			if str, ok := args[0].(*String); ok {
				msg = str.Value
			}
		}
		return &Error{Message: "SIGNAL:ROLLBACK:" + msg}
	})

	RegisterBuiltin("simpan", func(args ...Object) Object {
		return &Error{Message: "SIGNAL:COMMIT"}
	})

	// Atom Primitives
	RegisterBuiltin("atom_baru", func(args ...Object) Object {
		if len(args) != 1 {
			return newArgumentError(len(args), 1)
		}
		return &Atom{Value: args[0]}
	})

	RegisterBuiltin("atom_baca", func(args ...Object) Object {
		if len(args) != 1 {
			return newArgumentError(len(args), 1)
		}
		atom, ok := args[0].(*Atom)
		if !ok {
			return &Error{Code: ErrCodeTypeMismatch, Message: fmt.Sprintf("argument must be ATOM, got %s", args[0].Type())}
		}
		atom.Mu.Lock()
		defer atom.Mu.Unlock()
		return atom.Value
	})

	RegisterBuiltin("atom_tulis", func(args ...Object) Object {
		if len(args) != 2 {
			return newArgumentError(len(args), 2)
		}
		atom, ok := args[0].(*Atom)
		if !ok {
			return &Error{Code: ErrCodeTypeMismatch, Message: fmt.Sprintf("first argument must be ATOM, got %s", args[0].Type())}
		}
		atom.Mu.Lock()
		defer atom.Mu.Unlock()
		atom.Value = args[1]
		return &Null{}
	})

	RegisterBuiltin("atom_tukar", func(args ...Object) Object {
		if len(args) != 3 {
			return newArgumentError(len(args), 3)
		}
		atom, ok := args[0].(*Atom)
		if !ok {
			return &Error{Code: ErrCodeTypeMismatch, Message: fmt.Sprintf("first argument must be ATOM, got %s", args[0].Type())}
		}

		oldVal := args[1]
		newVal := args[2]

		atom.Mu.Lock()
		defer atom.Mu.Unlock()

		if atom.Value == oldVal {
			atom.Value = newVal
			return &Boolean{Value: true}
		}
		return &Boolean{Value: false}
	})

	RegisterBuiltin("alokasi", func(args ...Object) Object {
		if len(args) != 1 {
			return newArgumentError(len(args), 1)
		}
		size, ok := args[0].(*Integer)
		if !ok {
			return &Error{Code: ErrCodeTypeMismatch, Message: fmt.Sprintf("argument to `alokasi` must be INTEGER, got %s", args[0].Type())}
		}

		ptr, err := memory.Lemari.Alloc(int(size.Value))
		if err != nil {
			return &Error{Code: ErrCodeRuntime, Message: "allocation failed: " + err.Error()}
		}
		return &Pointer{Address: uint64(ptr)}
	})

	RegisterBuiltin("tulis_mem", func(args ...Object) Object {
		if len(args) != 2 {
			return newArgumentError(len(args), 2)
		}
		ptr, ok := args[0].(*Pointer)
		if !ok {
			return &Error{Code: ErrCodeTypeMismatch, Message: fmt.Sprintf("first argument to `tulis_mem` must be POINTER, got %s", args[0].Type())}
		}

		var data []byte
		switch val := args[1].(type) {
		case *String:
			data = []byte(val.Value)
		default:
			return &Error{Code: ErrCodeTypeMismatch, Message: fmt.Sprintf("second argument to `tulis_mem` must be STRING, got %s", args[1].Type())}
		}

		err := memory.Write(memory.Ptr(ptr.Address), data)
		if err != nil {
			return &Error{Code: ErrCodeRuntime, Message: "tulis_mem error: " + err.Error()}
		}
		return &Null{}
	})

	RegisterBuiltin("baca_mem", func(args ...Object) Object {
		if len(args) != 2 {
			return newArgumentError(len(args), 2)
		}
		ptr, ok := args[0].(*Pointer)
		if !ok {
			return &Error{Code: ErrCodeTypeMismatch, Message: fmt.Sprintf("first argument to `baca_mem` must be POINTER, got %s", args[0].Type())}
		}
		size, ok := args[1].(*Integer)
		if !ok {
			return &Error{Code: ErrCodeTypeMismatch, Message: fmt.Sprintf("second argument to `baca_mem` must be INTEGER, got %s", args[1].Type())}
		}

		data, err := memory.Read(memory.Ptr(ptr.Address), int(size.Value))
		if err != nil {
			return &Error{Code: ErrCodeRuntime, Message: "baca_mem error: " + err.Error()}
		}
		return &String{Value: string(data)}
	})

	RegisterBuiltin("alamat", func(args ...Object) Object {
		if len(args) != 1 {
			return newArgumentError(len(args), 1)
		}
		ptr, ok := args[0].(*Pointer)
		if !ok {
			return &Error{Code: ErrCodeTypeMismatch, Message: fmt.Sprintf("argument to `alamat` must be POINTER, got %s", args[0].Type())}
		}
		return &Integer{Value: int64(ptr.Address)}
	})

	RegisterBuiltin("ptr_dari", func(args ...Object) Object {
		if len(args) != 1 {
			return newArgumentError(len(args), 1)
		}
		addr, ok := args[0].(*Integer)
		if !ok {
			return &Error{Code: ErrCodeTypeMismatch, Message: fmt.Sprintf("argument to `ptr_dari` must be INTEGER, got %s", args[0].Type())}
		}
		return &Pointer{Address: uint64(addr.Value)}
	})
}

// RegisterBuiltin registers a new builtin function dynamically.
// Useful for adding builtins from other files (e.g. system info).
func RegisterBuiltin(name string, fn BuiltinFunction) {
	// Registry Check: Check for name collision
	for i, def := range Builtins {
		if def.Name == name {
			// Collision detected! Use composition to support overloading.
			originalFn := def.Builtin.Fn
			newFn := fn

			composedFn := func(args ...Object) Object {
				// Try the NEW function first
				result := newFn(args...)

				// If it fails with a type/arg mismatch, try the OLD function
				if err, ok := result.(*Error); ok {
					// Check specific error codes for fallback
					if err.Code == ErrCodeMissingArgs || err.Code == ErrCodeTooManyArgs || err.Code == ErrCodeTypeMismatch {
						return originalFn(args...)
					}
				}
				return result
			}

			// Update the existing builtin
			Builtins[i].Builtin.Fn = composedFn
			return
		}
	}

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
