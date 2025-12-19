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
		return NewError(fmt.Sprintf("argument mismatch: expected %d, got %d", expected, got), ErrCodeMissingArgs, 0, 0)
	}
	return NewError(fmt.Sprintf("argument mismatch: expected %d, got %d", expected, got), ErrCodeTooManyArgs, 0, 0)
}

func newArgumentErrorRange(got int, min int, max int) *Error {
	if got < min {
		return NewError(fmt.Sprintf("argument mismatch: expected %d or %d, got %d", min, max, got), ErrCodeMissingArgs, 0, 0)
	}
	return NewError(fmt.Sprintf("argument mismatch: expected %d or %d, got %d", min, max, got), ErrCodeTooManyArgs, 0, 0)
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
			if err != nil { return NewError(err.Error(), ErrCodeRuntime, 0, 0) }
			return NewInteger(int64(len))
		case *String:
			val := arg.GetValue()
			return NewInteger(int64(len(val)))
		default:
			return NewError(fmt.Sprintf("argument to `panjang` not supported, got %s", args[0].Type()), ErrCodeTypeMismatch, 0, 0)
		}
	})

	RegisterBuiltin("kunci", func(args ...Object) Object {
		if len(args) != 1 {
			return newArgumentError(len(args), 1)
		}
		hash, ok := args[0].(*Hash)
		if !ok {
			return NewError(fmt.Sprintf("argument to `kunci` must be HASH, got %s", args[0].Type()), ErrCodeTypeMismatch, 0, 0)
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
			return NewError(arg.GetValue(), "", 0, 0)
		default:
			return NewError(fmt.Sprintf("argument to `galat` must be STRING, got %s", args[0].Type()), ErrCodeTypeMismatch, 0, 0)
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
			return NewString(arg.GetMessage())
		default:
			return NewError(fmt.Sprintf("argument to `pesan_galat` must be ERROR, got %s", args[0].Type()), ErrCodeTypeMismatch, 0, 0)
		}
	})

	RegisterBuiltin("baca_file", func(args ...Object) Object {
		if len(args) != 1 {
			return newArgumentError(len(args), 1)
		}
		path, ok := args[0].(*String)
		if !ok {
			return NewError(fmt.Sprintf("argument to `baca_file` must be STRING, got %s", args[0].Type()), ErrCodeTypeMismatch, 0, 0)
		}

		content, err := os.ReadFile(path.GetValue())
		if err != nil {
			return NewError(fmt.Sprintf("baca_file error: %s", err.Error()), ErrCodeRuntime, 0, 0)
		}
		return NewString(string(content))
	})

	RegisterBuiltin("tulis_file", func(args ...Object) Object {
		if len(args) != 2 {
			return newArgumentError(len(args), 2)
		}
		path, ok := args[0].(*String)
		if !ok {
			return NewError(fmt.Sprintf("first argument to `tulis_file` must be STRING, got %s", args[0].Type()), ErrCodeTypeMismatch, 0, 0)
		}
		content, ok := args[1].(*String)
		if !ok {
			return NewError(fmt.Sprintf("second argument to `tulis_file` must be STRING, got %s", args[1].Type()), ErrCodeTypeMismatch, 0, 0)
		}

		err := os.WriteFile(path.GetValue(), []byte(content.GetValue()), 0644)
		if err != nil {
			return NewError(fmt.Sprintf("tulis_file error: %s", err.Error()), ErrCodeRuntime, 0, 0)
		}
		return NewNull()
	})

	RegisterBuiltin("buka_file", func(args ...Object) Object {
		if len(args) != 2 {
			return newArgumentError(len(args), 2)
		}
		path, ok := args[0].(*String)
		if !ok {
			return NewError(fmt.Sprintf("first argument to `buka_file` must be STRING, got %s", args[0].Type()), ErrCodeTypeMismatch, 0, 0)
		}
		mode, ok := args[1].(*String)
		if !ok {
			return NewError(fmt.Sprintf("second argument to `buka_file` must be STRING, got %s", args[1].Type()), ErrCodeTypeMismatch, 0, 0)
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
			return NewError("unknown file mode: "+modeVal, ErrCodeTypeMismatch, 0, 0)
		}

		f, err := os.OpenFile(pathVal, flag, 0644)
		if err != nil {
			return NewError("failed to open file: "+err.Error(), ErrCodeRuntime, 0, 0)
		}
		return NewFile(f, modeVal)
	})

	RegisterBuiltin("tutup_file", func(args ...Object) Object {
		if len(args) != 1 {
			return newArgumentError(len(args), 1)
		}
		f, ok := args[0].(*File)
		if !ok {
			return NewError(fmt.Sprintf("argument to `tutup_file` must be FILE, got %s", args[0].Type()), ErrCodeTypeMismatch, 0, 0)
		}
		err := f.File.Close()
		if err != nil {
			return NewError("failed to close file: "+err.Error(), ErrCodeRuntime, 0, 0)
		}
		return NewNull()
	})

	RegisterBuiltin("baca", func(args ...Object) Object {
		if len(args) < 1 || len(args) > 2 {
			return newArgumentErrorRange(len(args), 1, 2)
		}
		f, ok := args[0].(*File)
		if !ok {
			return NewError(fmt.Sprintf("first argument to `baca` must be FILE, got %s", args[0].Type()), ErrCodeTypeMismatch, 0, 0)
		}

		var limit int64 = -1
		if len(args) == 2 {
			n, ok := args[1].(*Integer)
			if !ok {
				return NewError(fmt.Sprintf("second argument to `baca` must be INTEGER, got %s", args[1].Type()), ErrCodeTypeMismatch, 0, 0)
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
			return NewError("baca error: "+err.Error(), ErrCodeRuntime, 0, 0)
		}
		return NewString(string(content))
	})

	RegisterBuiltin("tulis", func(args ...Object) Object {
		if len(args) != 2 {
			return newArgumentError(len(args), 2)
		}
		f, ok := args[0].(*File)
		if !ok {
			return NewError(fmt.Sprintf("first argument to `tulis` must be FILE, got %s", args[0].Type()), ErrCodeTypeMismatch, 0, 0)
		}

		var content []byte
		if str, ok := args[1].(*String); ok {
			content = []byte(str.GetValue())
		} else {
			content = []byte(args[1].Inspect())
		}

		_, err := f.File.Write(content)
		if err != nil {
			return NewError("tulis error: "+err.Error(), ErrCodeRuntime, 0, 0)
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
				return NewError(fmt.Sprintf("argument to `input` must be STRING, got %s", args[0].Type()), ErrCodeTypeMismatch, 0, 0)
			}
			fmt.Print(prompt.GetValue())
		}

		reader := bufio.NewReader(os.Stdin)
		text, err := reader.ReadString('\n')
		if err != nil {
			return NewError(fmt.Sprintf("input error: %s", err.Error()), ErrCodeRuntime, 0, 0)
		}
		return NewString(strings.TrimSpace(text))
	})

	RegisterBuiltin("huruf_besar", func(args ...Object) Object {
		if len(args) != 1 {
			return newArgumentError(len(args), 1)
		}
		str, ok := args[0].(*String)
		if !ok {
			return NewError(fmt.Sprintf("argument to `huruf_besar` must be STRING, got %s", args[0].Type()), ErrCodeTypeMismatch, 0, 0)
		}
		return NewString(strings.ToUpper(str.GetValue()))
	})

	RegisterBuiltin("huruf_kecil", func(args ...Object) Object {
		if len(args) != 1 {
			return newArgumentError(len(args), 1)
		}
		str, ok := args[0].(*String)
		if !ok {
			return NewError(fmt.Sprintf("argument to `huruf_kecil` must be STRING, got %s", args[0].Type()), ErrCodeTypeMismatch, 0, 0)
		}
		return NewString(strings.ToLower(str.GetValue()))
	})

	RegisterBuiltin("pisah", func(args ...Object) Object {
		if len(args) != 2 {
			return newArgumentError(len(args), 2)
		}
		str, ok := args[0].(*String)
		if !ok {
			return NewError(fmt.Sprintf("first argument to `pisah` must be STRING, got %s", args[0].Type()), ErrCodeTypeMismatch, 0, 0)
		}
		delim, ok := args[1].(*String)
		if !ok {
			return NewError(fmt.Sprintf("second argument to `pisah` must be STRING, got %s", args[1].Type()), ErrCodeTypeMismatch, 0, 0)
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
			return NewError(fmt.Sprintf("first argument to `gabung` must be ARRAY, got %s", args[0].Type()), ErrCodeTypeMismatch, 0, 0)
		}
		delim, ok := args[1].(*String)
		if !ok {
			return NewError(fmt.Sprintf("second argument to `gabung` must be STRING, got %s", args[1].Type()), ErrCodeTypeMismatch, 0, 0)
		}

		len, _ := memory.ReadArrayLength(arr.Address)
		parts := make([]string, len)
		for i := 0; i < len; i++ {
			elPtr, _ := memory.ReadArrayElement(arr.Address, i)
			el := FromPtr(elPtr)
			str, ok := el.(*String)
			if !ok {
				return NewError(fmt.Sprintf("array elements must be STRING, got %s", el.Type()), ErrCodeTypeMismatch, 0, 0)
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
			return NewError(fmt.Sprintf("argument to `abs` must be INTEGER, got %s", args[0].Type()), ErrCodeTypeMismatch, 0, 0)
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
			return NewError(fmt.Sprintf("first argument to `max` must be INTEGER, got %s", args[0].Type()), ErrCodeTypeMismatch, 0, 0)
		}
		b, ok := args[1].(*Integer)
		if !ok {
			return NewError(fmt.Sprintf("second argument to `max` must be INTEGER, got %s", args[1].Type()), ErrCodeTypeMismatch, 0, 0)
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
			return NewError(fmt.Sprintf("first argument to `min` must be INTEGER, got %s", args[0].Type()), ErrCodeTypeMismatch, 0, 0)
		}
		b, ok := args[1].(*Integer)
		if !ok {
			return NewError(fmt.Sprintf("second argument to `min` must be INTEGER, got %s", args[1].Type()), ErrCodeTypeMismatch, 0, 0)
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
			return NewError("first argument to `pow` "+err.Error(), ErrCodeTypeMismatch, 0, 0)
		}
		yVal, err := getFloatValue(args[1])
		if err != nil {
			return NewError("second argument to `pow` "+err.Error(), ErrCodeTypeMismatch, 0, 0)
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
			return NewError("argument to `sqrt` "+err.Error(), ErrCodeTypeMismatch, 0, 0)
		}

		if val < 0 {
			return NewError("cannot calculate square root of negative number", ErrCodeRuntime, 0, 0)
		}
		res := math.Sqrt(val)
		return NewFloat(res)
	})

	RegisterBuiltin("sin", func(args ...Object) Object {
		if len(args) != 1 { return newArgumentError(len(args), 1) }
		val, err := getFloatValue(args[0])
		if err != nil { return NewError("argument to `sin` "+err.Error(), ErrCodeTypeMismatch, 0, 0) }
		return NewFloat(math.Sin(val))
	})

	RegisterBuiltin("cos", func(args ...Object) Object {
		if len(args) != 1 { return newArgumentError(len(args), 1) }
		val, err := getFloatValue(args[0])
		if err != nil { return NewError("argument to `cos` "+err.Error(), ErrCodeTypeMismatch, 0, 0) }
		return NewFloat(math.Cos(val))
	})

	RegisterBuiltin("tan", func(args ...Object) Object {
		if len(args) != 1 { return newArgumentError(len(args), 1) }
		val, err := getFloatValue(args[0])
		if err != nil { return NewError("argument to `tan` "+err.Error(), ErrCodeTypeMismatch, 0, 0) }
		return NewFloat(math.Tan(val))
	})

	RegisterBuiltin("asin", func(args ...Object) Object {
		if len(args) != 1 { return newArgumentError(len(args), 1) }
		val, err := getFloatValue(args[0])
		if err != nil { return NewError("argument to `asin` "+err.Error(), ErrCodeTypeMismatch, 0, 0) }
		return NewFloat(math.Asin(val))
	})

	RegisterBuiltin("acos", func(args ...Object) Object {
		if len(args) != 1 { return newArgumentError(len(args), 1) }
		val, err := getFloatValue(args[0])
		if err != nil { return NewError("argument to `acos` "+err.Error(), ErrCodeTypeMismatch, 0, 0) }
		return NewFloat(math.Acos(val))
	})

	RegisterBuiltin("atan", func(args ...Object) Object {
		if len(args) != 1 { return newArgumentError(len(args), 1) }
		val, err := getFloatValue(args[0])
		if err != nil { return NewError("argument to `atan` "+err.Error(), ErrCodeTypeMismatch, 0, 0) }
		return NewFloat(math.Atan(val))
	})

	RegisterBuiltin("pi", func(args ...Object) Object {
		return NewFloat(math.Pi)
	})

	// ...
	RegisterBuiltin("waktu_sekarang", func(args ...Object) Object {
		return NewTime(time.Now())
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
			return NewError(fmt.Sprintf("argument to `tidur` must be INTEGER, got %s", args[0].Type()), ErrCodeTypeMismatch, 0, 0)
		}

		time.Sleep(time.Duration(ms.GetValue()) * time.Millisecond)
		return NewNull()
	})

	RegisterBuiltin("format_waktu", func(args ...Object) Object {
		if len(args) != 2 { return newArgumentError(len(args), 2) }
		tObj, ok := args[0].(*Time)
		if !ok { return NewError(fmt.Sprintf("first argument to `format_waktu` must be TIME, got %s", args[0].Type()), ErrCodeTypeMismatch, 0, 0) }
		formatObj, ok := args[1].(*String)
		if !ok { return NewError(fmt.Sprintf("second argument to `format_waktu` must be STRING, got %s", args[1].Type()), ErrCodeTypeMismatch, 0, 0) }

		layout := formatObj.GetValue()
		switch layout {
		case "RFC3339": layout = time.RFC3339
		case "ANSIC": layout = time.ANSIC
		case "UnixDate": layout = time.UnixDate
		}

		return NewString(tObj.Value.Format(layout))
	})

	RegisterBuiltin("potret", func(args ...Object) Object {
		return NewError("SIGNAL:SNAPSHOT", "", 0, 0)
	})

	RegisterBuiltin("pulih", func(args ...Object) Object {
		msg := "Manual Rollback"
		if len(args) > 0 {
			if str, ok := args[0].(*String); ok {
				msg = str.GetValue()
			}
		}
		return NewError("SIGNAL:ROLLBACK:"+msg, "", 0, 0)
	})

	RegisterBuiltin("simpan", func(args ...Object) Object {
		return NewError("SIGNAL:COMMIT", "", 0, 0)
	})

	RegisterBuiltin("atom_baru", func(args ...Object) Object {
		if len(args) != 1 { return newArgumentError(len(args), 1) }
		return NewAtom(args[0])
	})

	RegisterBuiltin("atom_baca", func(args ...Object) Object {
		if len(args) != 1 { return newArgumentError(len(args), 1) }
		atom, ok := args[0].(*Atom)
		if !ok { return NewError(fmt.Sprintf("argument must be ATOM, got %s", args[0].Type()), ErrCodeTypeMismatch, 0, 0) }
		atom.Mu.Lock()
		defer atom.Mu.Unlock()
		return atom.Value
	})

	RegisterBuiltin("atom_tulis", func(args ...Object) Object {
		if len(args) != 2 { return newArgumentError(len(args), 2) }
		atom, ok := args[0].(*Atom)
		if !ok { return NewError(fmt.Sprintf("first argument must be ATOM, got %s", args[0].Type()), ErrCodeTypeMismatch, 0, 0) }
		atom.Mu.Lock()
		defer atom.Mu.Unlock()
		atom.Value = args[1]
		return NewNull()
	})

	RegisterBuiltin("atom_tukar", func(args ...Object) Object {
		if len(args) != 3 { return newArgumentError(len(args), 3) }
		atom, ok := args[0].(*Atom)
		if !ok { return NewError(fmt.Sprintf("first argument must be ATOM, got %s", args[0].Type()), ErrCodeTypeMismatch, 0, 0) }

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
		if !ok { return NewError(fmt.Sprintf("argument to `alokasi` must be INTEGER, got %s", args[0].Type()), ErrCodeTypeMismatch, 0, 0) }

		ptr, err := memory.Lemari.Alloc(int(size.GetValue()))
		if err != nil {
			return NewError("allocation failed: "+err.Error(), ErrCodeRuntime, 0, 0)
		}
		return NewPointer(uint64(ptr))
	})

	RegisterBuiltin("tulis_mem", func(args ...Object) Object {
		if len(args) != 2 { return newArgumentError(len(args), 2) }
		ptr, ok := args[0].(*Pointer)
		if !ok { return NewError(fmt.Sprintf("first argument to `tulis_mem` must be POINTER, got %s", args[0].Type()), ErrCodeTypeMismatch, 0, 0) }

		var data []byte
		switch val := args[1].(type) {
		case *String:
			data = []byte(val.GetValue())
		default:
			return NewError(fmt.Sprintf("second argument to `tulis_mem` must be STRING, got %s", args[1].Type()), ErrCodeTypeMismatch, 0, 0)
		}

		err := memory.Write(memory.Ptr(ptr.GetValue()), data)
		if err != nil {
			return NewError("tulis_mem error: "+err.Error(), ErrCodeRuntime, 0, 0)
		}
		return NewNull()
	})

	RegisterBuiltin("baca_mem", func(args ...Object) Object {
		if len(args) != 2 { return newArgumentError(len(args), 2) }
		ptr, ok := args[0].(*Pointer)
		if !ok { return NewError(fmt.Sprintf("first argument to `baca_mem` must be POINTER, got %s", args[0].Type()), ErrCodeTypeMismatch, 0, 0) }
		size, ok := args[1].(*Integer)
		if !ok { return NewError(fmt.Sprintf("second argument to `baca_mem` must be INTEGER, got %s", args[1].Type()), ErrCodeTypeMismatch, 0, 0) }

		data, err := memory.Read(memory.Ptr(ptr.GetValue()), int(size.GetValue()))
		if err != nil {
			return NewError("baca_mem error: "+err.Error(), ErrCodeRuntime, 0, 0)
		}
		return NewString(string(data))
	})

	RegisterBuiltin("alamat", func(args ...Object) Object {
		if len(args) != 1 { return newArgumentError(len(args), 1) }
		ptr, ok := args[0].(*Pointer)
		if !ok { return NewError(fmt.Sprintf("argument to `alamat` must be POINTER, got %s", args[0].Type()), ErrCodeTypeMismatch, 0, 0) }
		return NewInteger(int64(ptr.GetValue()))
	})

	RegisterBuiltin("ptr_dari", func(args ...Object) Object {
		if len(args) != 1 { return newArgumentError(len(args), 1) }
		addr, ok := args[0].(*Integer)
		if !ok { return NewError(fmt.Sprintf("argument to `ptr_dari` must be INTEGER, got %s", args[0].Type()), ErrCodeTypeMismatch, 0, 0) }
		return NewPointer(uint64(addr.GetValue()))
	})

	RegisterBuiltin("chr", func(args ...Object) Object {
		if len(args) != 1 { return newArgumentError(len(args), 1) }
		code, ok := args[0].(*Integer)
		if !ok { return NewError(fmt.Sprintf("argument to `chr` must be INTEGER, got %s", args[0].Type()), ErrCodeTypeMismatch, 0, 0) }
		return NewString(string(rune(code.GetValue())))
	})

	RegisterBuiltin("ord", func(args ...Object) Object {
		if len(args) != 1 { return newArgumentError(len(args), 1) }
		str, ok := args[0].(*String)
		if !ok { return NewError(fmt.Sprintf("argument to `ord` must be STRING, got %s", args[0].Type()), ErrCodeTypeMismatch, 0, 0) }

		val := str.GetValue()
		if len(val) == 0 { return NewError("empty string", ErrCodeRuntime, 0, 0) }
		r := []rune(val)[0]
		return NewInteger(int64(r))
	})

	RegisterBuiltin("baca_byte", func(args ...Object) Object {
		if len(args) != 2 { return newArgumentError(len(args), 2) }
		ptr, ok := args[0].(*Pointer)
		if !ok { return NewError(fmt.Sprintf("first argument to `baca_byte` must be POINTER, got %s", args[0].Type()), ErrCodeTypeMismatch, 0, 0) }
		offset, ok := args[1].(*Integer)
		if !ok { return NewError(fmt.Sprintf("second argument to `baca_byte` must be INTEGER, got %s", args[1].Type()), ErrCodeTypeMismatch, 0, 0) }

		target := memory.Ptr(ptr.GetValue()).Add(uint32(offset.GetValue()))
		data, err := memory.Read(target, 1)
		if err != nil { return NewError("baca_byte error: "+err.Error(), ErrCodeRuntime, 0, 0) }

		return NewInteger(int64(data[0]))
	})

	RegisterBuiltin("tulis_byte", func(args ...Object) Object {
		if len(args) != 3 { return newArgumentError(len(args), 3) }
		ptr, ok := args[0].(*Pointer)
		if !ok { return NewError(fmt.Sprintf("first argument to `tulis_byte` must be POINTER, got %s", args[0].Type()), ErrCodeTypeMismatch, 0, 0) }
		offset, ok := args[1].(*Integer)
		if !ok { return NewError(fmt.Sprintf("second argument to `tulis_byte` must be INTEGER, got %s", args[1].Type()), ErrCodeTypeMismatch, 0, 0) }
		val, ok := args[2].(*Integer)
		if !ok { return NewError(fmt.Sprintf("third argument to `tulis_byte` must be INTEGER, got %s", args[2].Type()), ErrCodeTypeMismatch, 0, 0) }

		target := memory.Ptr(ptr.GetValue()).Add(uint32(offset.GetValue()))
		data := []byte{byte(val.GetValue())}
		err := memory.Write(target, data)
		if err != nil { return NewError("tulis_byte error: "+err.Error(), ErrCodeRuntime, 0, 0) }
		return NewNull()
	})
}

// ...

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

func RegisterBuiltin(name string, fn BuiltinFunction) {
	for i, def := range Builtins {
		if def.Name == name {
			originalFn := def.Builtin.Fn
			newFn := fn
			composedFn := func(args ...Object) Object {
				result := newFn(args...)
				if err, ok := result.(*Error); ok {
					code := err.GetCode()
					if code == ErrCodeMissingArgs || code == ErrCodeTooManyArgs || code == ErrCodeTypeMismatch {
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
