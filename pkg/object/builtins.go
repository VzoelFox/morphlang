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

var Builtins []BuiltinDef

func init() {
	RegisterBuiltin("panjang", func(args ...Object) Object {
		if len(args) != 1 {
			return newArgumentError(len(args), 1)
		}
		switch arg := args[0].(type) {
		case *Array:
			len, err := memory.ReadArrayLength(arg.Address)
			if err != nil { return &Error{Message: err.Error()} }
			return NewInteger(int64(len))
		case *String:
			val := arg.GetValue()
			return NewInteger(int64(len(val)))
		default:
			return &Error{Code: ErrCodeTypeMismatch, Message: fmt.Sprintf("argument to `panjang` not supported, got %s", args[0].Type())}
		}
	})

	RegisterBuiltin("kunci", func(args ...Object) Object {
		if len(args) != 1 {
			return newArgumentError(len(args), 1)
		}
		hash, ok := args[0].(*Hash)
		if !ok {
			return &Error{Code: ErrCodeTypeMismatch, Message: fmt.Sprintf("argument to `kunci` must be HASH, got %s", args[0].Type())}
		}

		pairs := hash.GetPairs()
		elements := make([]Object, 0, len(pairs))
		for _, pair := range pairs {
			elements = append(elements, pair.Key)
		}
		return NewArray(elements)
	})

	RegisterBuiltin("cetak", func(args ...Object) Object {
		for _, arg := range args {
			fmt.Println(arg.Inspect())
		}
		return NewNull()
	})

	RegisterBuiltin("tipe", func(args ...Object) Object {
		if len(args) != 1 {
			return newArgumentError(len(args), 1)
		}
		return NewString(string(args[0].Type()))
	})

	RegisterBuiltin("galat", func(args ...Object) Object {
		if len(args) != 1 {
			return newArgumentError(len(args), 1)
		}
		switch arg := args[0].(type) {
		case *String:
			return &Error{Message: arg.GetValue()}
		default:
			return &Error{Code: ErrCodeTypeMismatch, Message: fmt.Sprintf("argument to `galat` must be STRING, got %s", args[0].Type())}
		}
	})

	RegisterBuiltin("adalah_galat", func(args ...Object) Object {
		if len(args) != 1 {
			return newArgumentError(len(args), 1)
		}
		switch args[0].(type) {
		case *Error:
			return NewBoolean(true)
		default:
			return NewBoolean(false)
		}
	})

	RegisterBuiltin("pesan_galat", func(args ...Object) Object {
		if len(args) != 1 {
			return newArgumentError(len(args), 1)
		}
		switch arg := args[0].(type) {
		case *Error:
			return NewString(arg.Message)
		default:
			return &Error{Code: ErrCodeTypeMismatch, Message: fmt.Sprintf("argument to `pesan_galat` must be ERROR, got %s", args[0].Type())}
		}
	})

	RegisterBuiltin("baca_file", func(args ...Object) Object {
		if len(args) != 1 {
			return newArgumentError(len(args), 1)
		}
		path, ok := args[0].(*String)
		if !ok {
			return &Error{Code: ErrCodeTypeMismatch, Message: fmt.Sprintf("argument to `baca_file` must be STRING, got %s", args[0].Type())}
		}

		content, err := os.ReadFile(path.GetValue())
		if err != nil {
			return &Error{Code: ErrCodeRuntime, Message: fmt.Sprintf("baca_file error: %s", err.Error())}
		}
		return NewString(string(content))
	})

	RegisterBuiltin("tulis_file", func(args ...Object) Object {
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

		err := os.WriteFile(path.GetValue(), []byte(content.GetValue()), 0644)
		if err != nil {
			return &Error{Code: ErrCodeRuntime, Message: fmt.Sprintf("tulis_file error: %s", err.Error())}
		}
		return NewNull()
	})

	RegisterBuiltin("buka_file", func(args ...Object) Object {
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

		pathVal := path.GetValue()
		modeVal := mode.GetValue()
		var flag int
		switch modeVal {
		case "b":
			flag = os.O_RDONLY
		case "t":
			flag = os.O_WRONLY | os.O_CREATE | os.O_TRUNC
		case "tb":
			flag = os.O_WRONLY | os.O_CREATE | os.O_TRUNC
		case "t+":
			flag = os.O_WRONLY | os.O_CREATE | os.O_APPEND
		default:
			return &Error{Code: ErrCodeTypeMismatch, Message: "unknown file mode: " + modeVal}
		}

		f, err := os.OpenFile(pathVal, flag, 0644)
		if err != nil {
			return &Error{Code: ErrCodeRuntime, Message: "failed to open file: " + err.Error()}
		}
		return &File{File: f, Mode: modeVal}
	})

	RegisterBuiltin("tutup_file", func(args ...Object) Object {
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
		return NewNull()
	})

	RegisterBuiltin("baca", func(args ...Object) Object {
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
			limit = n.GetValue()
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
		return NewString(string(content))
	})

	RegisterBuiltin("tulis", func(args ...Object) Object {
		if len(args) != 2 {
			return newArgumentError(len(args), 2)
		}
		f, ok := args[0].(*File)
		if !ok {
			return &Error{Code: ErrCodeTypeMismatch, Message: fmt.Sprintf("first argument to `tulis` must be FILE, got %s", args[0].Type())}
		}

		var content []byte
		if str, ok := args[1].(*String); ok {
			content = []byte(str.GetValue())
		} else {
			content = []byte(args[1].Inspect())
		}

		_, err := f.File.Write(content)
		if err != nil {
			return &Error{Code: ErrCodeRuntime, Message: "tulis error: " + err.Error()}
		}
		return NewNull()
	})

	RegisterBuiltin("input", func(args ...Object) Object {
		if len(args) > 1 {
			return newArgumentErrorRange(len(args), 0, 1)
		}
		if len(args) == 1 {
			prompt, ok := args[0].(*String)
			if !ok {
				return &Error{Code: ErrCodeTypeMismatch, Message: fmt.Sprintf("argument to `input` must be STRING, got %s", args[0].Type())}
			}
			fmt.Print(prompt.GetValue())
		}

		reader := bufio.NewReader(os.Stdin)
		text, err := reader.ReadString('\n')
		if err != nil {
			return &Error{Code: ErrCodeRuntime, Message: fmt.Sprintf("input error: %s", err.Error())}
		}
		return NewString(strings.TrimSpace(text))
	})

	RegisterBuiltin("huruf_besar", func(args ...Object) Object {
		if len(args) != 1 {
			return newArgumentError(len(args), 1)
		}
		str, ok := args[0].(*String)
		if !ok {
			return &Error{Code: ErrCodeTypeMismatch, Message: fmt.Sprintf("argument to `huruf_besar` must be STRING, got %s", args[0].Type())}
		}
		return NewString(strings.ToUpper(str.GetValue()))
	})

	RegisterBuiltin("huruf_kecil", func(args ...Object) Object {
		if len(args) != 1 {
			return newArgumentError(len(args), 1)
		}
		str, ok := args[0].(*String)
		if !ok {
			return &Error{Code: ErrCodeTypeMismatch, Message: fmt.Sprintf("argument to `huruf_kecil` must be STRING, got %s", args[0].Type())}
		}
		return NewString(strings.ToLower(str.GetValue()))
	})

	RegisterBuiltin("pisah", func(args ...Object) Object {
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

		parts := strings.Split(str.GetValue(), delim.GetValue())
		elements := make([]Object, len(parts))
		for i, p := range parts {
			elements[i] = NewString(p)
		}
		return NewArray(elements)
	})

	RegisterBuiltin("gabung", func(args ...Object) Object {
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

		len, _ := memory.ReadArrayLength(arr.Address)
		parts := make([]string, len)
		for i := 0; i < len; i++ {
			elPtr, _ := memory.ReadArrayElement(arr.Address, i)
			el := FromPtr(elPtr)
			str, ok := el.(*String)
			if !ok {
				return &Error{Code: ErrCodeTypeMismatch, Message: fmt.Sprintf("array elements must be STRING, got %s", el.Type())}
			}
			parts[i] = str.GetValue()
		}
		return NewString(strings.Join(parts, delim.GetValue()))
	})

	RegisterBuiltin("abs", func(args ...Object) Object {
		if len(args) != 1 {
			return newArgumentError(len(args), 1)
		}
		num, ok := args[0].(*Integer)
		if !ok {
			return &Error{Code: ErrCodeTypeMismatch, Message: fmt.Sprintf("argument to `abs` must be INTEGER, got %s", args[0].Type())}
		}
		val := num.GetValue()
		if val < 0 {
			return NewInteger(-val)
		}
		return num
	})

	RegisterBuiltin("max", func(args ...Object) Object {
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
		if a.GetValue() > b.GetValue() {
			return a
		}
		return b
	})

	RegisterBuiltin("min", func(args ...Object) Object {
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
		if a.GetValue() < b.GetValue() {
			return a
		}
		return b
	})

	RegisterBuiltin("pow", func(args ...Object) Object {
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
		return NewFloat(res)
	})

	RegisterBuiltin("sqrt", func(args ...Object) Object {
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
		return NewFloat(res)
	})

	RegisterBuiltin("sin", func(args ...Object) Object {
		if len(args) != 1 { return newArgumentError(len(args), 1) }
		val, err := getFloatValue(args[0])
		if err != nil { return &Error{Code: ErrCodeTypeMismatch, Message: "argument to `sin` " + err.Error()} }
		return NewFloat(math.Sin(val))
	})

	RegisterBuiltin("cos", func(args ...Object) Object {
		if len(args) != 1 { return newArgumentError(len(args), 1) }
		val, err := getFloatValue(args[0])
		if err != nil { return &Error{Code: ErrCodeTypeMismatch, Message: "argument to `cos` " + err.Error()} }
		return NewFloat(math.Cos(val))
	})

	RegisterBuiltin("tan", func(args ...Object) Object {
		if len(args) != 1 { return newArgumentError(len(args), 1) }
		val, err := getFloatValue(args[0])
		if err != nil { return &Error{Code: ErrCodeTypeMismatch, Message: "argument to `tan` " + err.Error()} }
		return NewFloat(math.Tan(val))
	})

	RegisterBuiltin("asin", func(args ...Object) Object {
		if len(args) != 1 { return newArgumentError(len(args), 1) }
		val, err := getFloatValue(args[0])
		if err != nil { return &Error{Code: ErrCodeTypeMismatch, Message: "argument to `asin` " + err.Error()} }
		return NewFloat(math.Asin(val))
	})

	RegisterBuiltin("acos", func(args ...Object) Object {
		if len(args) != 1 { return newArgumentError(len(args), 1) }
		val, err := getFloatValue(args[0])
		if err != nil { return &Error{Code: ErrCodeTypeMismatch, Message: "argument to `acos` " + err.Error()} }
		return NewFloat(math.Acos(val))
	})

	RegisterBuiltin("atan", func(args ...Object) Object {
		if len(args) != 1 { return newArgumentError(len(args), 1) }
		val, err := getFloatValue(args[0])
		if err != nil { return &Error{Code: ErrCodeTypeMismatch, Message: "argument to `atan` " + err.Error()} }
		return NewFloat(math.Atan(val))
	})

	RegisterBuiltin("pi", func(args ...Object) Object {
		return NewFloat(math.Pi)
	})

	// Add others from init() here to merge
	RegisterBuiltin("waktu_sekarang", func(args ...Object) Object {
		return &Time{Value: time.Now()}
	})

	RegisterBuiltin("waktu_unix", func(args ...Object) Object {
		return NewInteger(time.Now().Unix())
	})

	RegisterBuiltin("tidur", func(args ...Object) Object {
		if len(args) != 1 {
			return newArgumentError(len(args), 1)
		}

		ms, ok := args[0].(*Integer)
		if !ok {
			return &Error{Code: ErrCodeTypeMismatch, Message: fmt.Sprintf("argument to `tidur` must be INTEGER, got %s", args[0].Type())}
		}

		time.Sleep(time.Duration(ms.GetValue()) * time.Millisecond)
		return NewNull()
	})

	RegisterBuiltin("format_waktu", func(args ...Object) Object {
		if len(args) != 2 { return newArgumentError(len(args), 2) }
		tObj, ok := args[0].(*Time)
		if !ok { return &Error{Code: ErrCodeTypeMismatch, Message: fmt.Sprintf("first argument to `format_waktu` must be TIME, got %s", args[0].Type())} }
		formatObj, ok := args[1].(*String)
		if !ok { return &Error{Code: ErrCodeTypeMismatch, Message: fmt.Sprintf("second argument to `format_waktu` must be STRING, got %s", args[1].Type())} }

		layout := formatObj.GetValue()
		switch layout {
		case "RFC3339": layout = time.RFC3339
		case "ANSIC": layout = time.ANSIC
		case "UnixDate": layout = time.UnixDate
		}

		return NewString(tObj.Value.Format(layout))
	})

	RegisterBuiltin("potret", func(args ...Object) Object {
		return &Error{Message: "SIGNAL:SNAPSHOT"}
	})

	RegisterBuiltin("pulih", func(args ...Object) Object {
		msg := "Manual Rollback"
		if len(args) > 0 {
			if str, ok := args[0].(*String); ok {
				msg = str.GetValue()
			}
		}
		return &Error{Message: "SIGNAL:ROLLBACK:" + msg}
	})

	RegisterBuiltin("simpan", func(args ...Object) Object {
		return &Error{Message: "SIGNAL:COMMIT"}
	})

	RegisterBuiltin("atom_baru", func(args ...Object) Object {
		if len(args) != 1 { return newArgumentError(len(args), 1) }
		return &Atom{Value: args[0]}
	})

	RegisterBuiltin("atom_baca", func(args ...Object) Object {
		if len(args) != 1 { return newArgumentError(len(args), 1) }
		atom, ok := args[0].(*Atom)
		if !ok { return &Error{Code: ErrCodeTypeMismatch, Message: fmt.Sprintf("argument must be ATOM, got %s", args[0].Type())} }
		atom.Mu.Lock()
		defer atom.Mu.Unlock()
		return atom.Value
	})

	RegisterBuiltin("atom_tulis", func(args ...Object) Object {
		if len(args) != 2 { return newArgumentError(len(args), 2) }
		atom, ok := args[0].(*Atom)
		if !ok { return &Error{Code: ErrCodeTypeMismatch, Message: fmt.Sprintf("first argument must be ATOM, got %s", args[0].Type())} }
		atom.Mu.Lock()
		defer atom.Mu.Unlock()
		atom.Value = args[1]
		return NewNull()
	})

	RegisterBuiltin("atom_tukar", func(args ...Object) Object {
		if len(args) != 3 { return newArgumentError(len(args), 3) }
		atom, ok := args[0].(*Atom)
		if !ok { return &Error{Code: ErrCodeTypeMismatch, Message: fmt.Sprintf("first argument must be ATOM, got %s", args[0].Type())} }

		oldVal := args[1]
		newVal := args[2]

		atom.Mu.Lock()
		defer atom.Mu.Unlock()

		if atom.Value == oldVal {
			atom.Value = newVal
			return NewBoolean(true)
		}
		return NewBoolean(false)
	})

	RegisterBuiltin("alokasi", func(args ...Object) Object {
		if len(args) != 1 { return newArgumentError(len(args), 1) }
		size, ok := args[0].(*Integer)
		if !ok { return &Error{Code: ErrCodeTypeMismatch, Message: fmt.Sprintf("argument to `alokasi` must be INTEGER, got %s", args[0].Type())} }

		ptr, err := memory.Lemari.Alloc(int(size.GetValue()))
		if err != nil {
			return &Error{Code: ErrCodeRuntime, Message: "allocation failed: " + err.Error()}
		}
		return &Pointer{Address: uint64(ptr)}
	})

	RegisterBuiltin("tulis_mem", func(args ...Object) Object {
		if len(args) != 2 { return newArgumentError(len(args), 2) }
		ptr, ok := args[0].(*Pointer)
		if !ok { return &Error{Code: ErrCodeTypeMismatch, Message: fmt.Sprintf("first argument to `tulis_mem` must be POINTER, got %s", args[0].Type())} }

		var data []byte
		switch val := args[1].(type) {
		case *String:
			data = []byte(val.GetValue())
		default:
			return &Error{Code: ErrCodeTypeMismatch, Message: fmt.Sprintf("second argument to `tulis_mem` must be STRING, got %s", args[1].Type())}
		}

		err := memory.Write(memory.Ptr(ptr.Address), data)
		if err != nil {
			return &Error{Code: ErrCodeRuntime, Message: "tulis_mem error: " + err.Error()}
		}
		return NewNull()
	})

	RegisterBuiltin("baca_mem", func(args ...Object) Object {
		if len(args) != 2 { return newArgumentError(len(args), 2) }
		ptr, ok := args[0].(*Pointer)
		if !ok { return &Error{Code: ErrCodeTypeMismatch, Message: fmt.Sprintf("first argument to `baca_mem` must be POINTER, got %s", args[0].Type())} }
		size, ok := args[1].(*Integer)
		if !ok { return &Error{Code: ErrCodeTypeMismatch, Message: fmt.Sprintf("second argument to `baca_mem` must be INTEGER, got %s", args[1].Type())} }

		data, err := memory.Read(memory.Ptr(ptr.Address), int(size.GetValue()))
		if err != nil {
			return &Error{Code: ErrCodeRuntime, Message: "baca_mem error: " + err.Error()}
		}
		return NewString(string(data))
	})

	RegisterBuiltin("alamat", func(args ...Object) Object {
		if len(args) != 1 { return newArgumentError(len(args), 1) }
		ptr, ok := args[0].(*Pointer)
		if !ok { return &Error{Code: ErrCodeTypeMismatch, Message: fmt.Sprintf("argument to `alamat` must be POINTER, got %s", args[0].Type())} }
		return NewInteger(int64(ptr.Address))
	})

	RegisterBuiltin("ptr_dari", func(args ...Object) Object {
		if len(args) != 1 { return newArgumentError(len(args), 1) }
		addr, ok := args[0].(*Integer)
		if !ok { return &Error{Code: ErrCodeTypeMismatch, Message: fmt.Sprintf("argument to `ptr_dari` must be INTEGER, got %s", args[0].Type())} }
		return &Pointer{Address: uint64(addr.GetValue())}
	})
}

func getFloatValue(obj Object) (float64, error) {
	switch val := obj.(type) {
	case *Integer:
		return float64(val.GetValue()), nil
	case *Float:
		return val.GetValue(), nil
	default:
		return 0, fmt.Errorf("must be INTEGER or FLOAT, got %s", val.Type())
	}
}

// ... RegisterBuiltin ...
func RegisterBuiltin(name string, fn BuiltinFunction) {
	for i, def := range Builtins {
		if def.Name == name {
			originalFn := def.Builtin.Fn
			newFn := fn
			composedFn := func(args ...Object) Object {
				result := newFn(args...)
				if err, ok := result.(*Error); ok {
					if err.Code == ErrCodeMissingArgs || err.Code == ErrCodeTooManyArgs || err.Code == ErrCodeTypeMismatch {
						return originalFn(args...)
					}
				}
				return result
			}
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
