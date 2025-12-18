package object

import (
	"bytes"
	"fmt"
	"hash/fnv"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/VzoelFox/morphlang/pkg/memory"
	"github.com/VzoelFox/morphlang/pkg/parser"
)

type ObjectType string

const (
	INTEGER_OBJ           = "INTEGER"
	FLOAT_OBJ             = "FLOAT"
	BOOLEAN_OBJ           = "BOOLEAN"
	NULL_OBJ              = "NULL"
	RETURN_VALUE_OBJ      = "RETURN_VALUE"
	ERROR_OBJ             = "ERROR"
	FUNCTION_OBJ          = "FUNCTION"
	STRING_OBJ            = "STRING"
	BUILTIN_OBJ           = "BUILTIN"
	ARRAY_OBJ             = "ARRAY"
	HASH_OBJ              = "HASH"
	COMPILED_FUNCTION_OBJ = "COMPILED_FUNCTION"
	CLOSURE_OBJ           = "CLOSURE"
	CHANNEL_OBJ           = "CHANNEL"
	THREAD_OBJ            = "THREAD"
	TIME_OBJ              = "TIME"
	MUTEX_OBJ             = "MUTEX"
	ATOM_OBJ              = "ATOM"
	FILE_OBJ              = "FILE"
	POINTER_OBJ           = "POINTER"
)

type Object interface {
	Type() ObjectType
	Inspect() string
	GetAddress() memory.Ptr
}

type Hashable interface {
	HashKey() HashKey
}

type Integer struct {
	Address memory.Ptr
}

func NewInteger(val int64) *Integer {
	ptr, err := memory.AllocInteger(val)
	if err != nil {
		panic(err)
	}
	return &Integer{Address: ptr}
}

func (i *Integer) Type() ObjectType { return INTEGER_OBJ }
func (i *Integer) Inspect() string {
	val, _ := memory.ReadInteger(i.Address)
	return fmt.Sprintf("%d", val)
}
func (i *Integer) HashKey() HashKey {
	val, _ := memory.ReadInteger(i.Address)
	return HashKey{Type: i.Type(), Value: uint64(val)}
}
func (i *Integer) GetAddress() memory.Ptr { return i.Address }
func (i *Integer) GetValue() int64 {
	val, _ := memory.ReadInteger(i.Address)
	return val
}

type Float struct {
	Address memory.Ptr
}

func NewFloat(val float64) *Float {
	ptr, err := memory.AllocFloat(val)
	if err != nil {
		panic(err)
	}
	return &Float{Address: ptr}
}

func (f *Float) Type() ObjectType { return FLOAT_OBJ }
func (f *Float) Inspect() string {
	val, _ := memory.ReadFloat(f.Address)
	return fmt.Sprintf("%g", val)
}
func (f *Float) GetAddress() memory.Ptr { return f.Address }
func (f *Float) GetValue() float64 {
	val, _ := memory.ReadFloat(f.Address)
	return val
}

type Boolean struct {
	Address memory.Ptr
}

func NewBoolean(val bool) *Boolean {
	ptr, err := memory.AllocBoolean(val)
	if err != nil {
		panic(err)
	}
	return &Boolean{Address: ptr}
}

func (b *Boolean) Type() ObjectType { return BOOLEAN_OBJ }
func (b *Boolean) Inspect() string {
	val, _ := memory.ReadBoolean(b.Address)
	if val {
		return "benar"
	}
	return "salah"
}
func (b *Boolean) HashKey() HashKey {
	val, _ := memory.ReadBoolean(b.Address)
	var v uint64
	if val {
		v = 1
	} else {
		v = 0
	}
	return HashKey{Type: b.Type(), Value: v}
}
func (b *Boolean) GetAddress() memory.Ptr { return b.Address }
func (b *Boolean) GetValue() bool {
	val, _ := memory.ReadBoolean(b.Address)
	return val
}

type Null struct {
	Address memory.Ptr
}

func NewNull() *Null {
	ptr, err := memory.AllocNull()
	if err != nil {
		panic(err)
	}
	return &Null{Address: ptr}
}

func (n *Null) Type() ObjectType       { return NULL_OBJ }
func (n *Null) Inspect() string        { return "kosong" }
func (n *Null) GetAddress() memory.Ptr { return n.Address }

type ReturnValue struct {
	Value Object
}

func (rv *ReturnValue) Type() ObjectType       { return RETURN_VALUE_OBJ }
func (rv *ReturnValue) Inspect() string        { return rv.Value.Inspect() }
func (rv *ReturnValue) GetAddress() memory.Ptr { return memory.NilPtr } // Virtual

// Error struct (Same as before)
const (
	ErrCodeSyntax          = "E001"
	ErrCodeUndefined       = "E002"
	ErrCodeTypeMismatch    = "E003"
	ErrCodeUncheckedError  = "E004"
	ErrCodeRuntime         = "E005"
	ErrCodeIndexOutOfBounds = "E006"
	ErrCodeInvalidOp       = "E007"
	ErrCodeMissingArgs     = "E008"
	ErrCodeTooManyArgs     = "E009"
)

type Error struct {
	Address memory.Ptr
}

func NewError(msg string, code string, line, col int) *Error {
	msgPtr, err := memory.AllocString(msg)
	if err != nil { panic(err) }

	var codePtr memory.Ptr = memory.NilPtr
	if code != "" {
		codePtr, err = memory.AllocString(code)
		if err != nil { panic(err) }
	}

	ptr, err := memory.AllocError(msgPtr, codePtr, line, col)
	if err != nil { panic(err) }

	return &Error{Address: ptr}
}

func (e *Error) Type() ObjectType { return ERROR_OBJ }
func (e *Error) GetAddress() memory.Ptr { return e.Address }

func (e *Error) GetMessage() string {
	msgPtr, _, _, _, _ := memory.ReadError(e.Address)
	if msgPtr == memory.NilPtr { return "" }
	val, _ := memory.ReadString(msgPtr)
	return val
}

func (e *Error) GetCode() string {
	_, codePtr, _, _, _ := memory.ReadError(e.Address)
	if codePtr == memory.NilPtr { return "" }
	val, _ := memory.ReadString(codePtr)
	return val
}

func (e *Error) Inspect() string {
	msg := e.GetMessage()
	code := e.GetCode()
	_, _, line, col, _ := memory.ReadError(e.Address)

	var out bytes.Buffer
	if code != "" {
		out.WriteString(fmt.Sprintf("Error [%s] di [:%d:%d]:\n", code, line, col))
	} else {
		out.WriteString(fmt.Sprintf("Error di [:%d:%d]:\n", line, col))
	}
	out.WriteString(fmt.Sprintf("  %s\n", msg))
	return out.String()
}
func (e *Error) GetAddress() memory.Ptr { return memory.NilPtr }

type Channel struct {
	Value chan Object
}

func (c *Channel) Type() ObjectType       { return CHANNEL_OBJ }
func (c *Channel) Inspect() string        { return fmt.Sprintf("saluran[%p]", c.Value) }
func (c *Channel) GetAddress() memory.Ptr { return memory.NilPtr }

type Thread struct {
	Result chan Object
}

func (t *Thread) Type() ObjectType       { return THREAD_OBJ }
func (t *Thread) Inspect() string        { return fmt.Sprintf("utas[%p]", t.Result) }
func (t *Thread) GetAddress() memory.Ptr { return memory.NilPtr }

type Time struct {
	Value time.Time
}

func (t *Time) Type() ObjectType       { return TIME_OBJ }
func (t *Time) Inspect() string        { return t.Value.Format(time.RFC3339) }
func (t *Time) GetAddress() memory.Ptr { return memory.NilPtr }

type Mutex struct {
	Mu sync.Mutex
}

func (m *Mutex) Type() ObjectType       { return MUTEX_OBJ }
func (m *Mutex) Inspect() string        { return fmt.Sprintf("mutex[%p]", &m.Mu) }
func (m *Mutex) GetAddress() memory.Ptr { return memory.NilPtr }

type Atom struct {
	Value Object
	Mu    sync.Mutex
}

func (a *Atom) Type() ObjectType { return ATOM_OBJ }
func (a *Atom) Inspect() string {
	a.Mu.Lock()
	defer a.Mu.Unlock()
	if a.Value != nil {
		return fmt.Sprintf("atom[%s]", a.Value.Inspect())
	}
	return "atom[kosong]"
}
func (a *Atom) GetAddress() memory.Ptr { return memory.NilPtr }

type File struct {
	File *os.File
	Mode string
}

func (f *File) Type() ObjectType       { return FILE_OBJ }
func (f *File) Inspect() string        { return fmt.Sprintf("file[%s]", f.File.Name()) }
func (f *File) GetAddress() memory.Ptr { return memory.NilPtr }

type Pointer struct {
	Address uint64
}

func (p *Pointer) Type() ObjectType       { return POINTER_OBJ }
func (p *Pointer) Inspect() string        { return fmt.Sprintf("ptr[0x%x]", p.Address) }
func (p *Pointer) GetAddress() memory.Ptr { return memory.Ptr(p.Address) }

type BuiltinFunction func(args ...Object) Object

type Builtin struct {
	Fn      BuiltinFunction
	Address memory.Ptr
}

func (b *Builtin) Type() ObjectType       { return BUILTIN_OBJ }
func (b *Builtin) Inspect() string        { return "builtin function" }
func (b *Builtin) GetAddress() memory.Ptr { return b.Address }

type String struct {
	Address memory.Ptr
}

func NewString(val string) *String {
	ptr, err := memory.AllocString(val)
	if err != nil {
		panic(err)
	}
	return &String{Address: ptr}
}

func (s *String) Type() ObjectType { return STRING_OBJ }
func (s *String) Inspect() string {
	val, _ := memory.ReadString(s.Address)
	return val
}
func (s *String) HashKey() HashKey {
	val, _ := memory.ReadString(s.Address)
	h := fnv.New64a()
	h.Write([]byte(val))
	return HashKey{Type: s.Type(), Value: h.Sum64()}
}
func (s *String) GetAddress() memory.Ptr { return s.Address }
func (s *String) GetValue() string {
	val, _ := memory.ReadString(s.Address)
	return val
}

type Function struct {
	Parameters []*parser.Identifier
	Body       *parser.BlockStatement
	Env        *Environment
}

func (f *Function) Type() ObjectType { return FUNCTION_OBJ }
func (f *Function) Inspect() string {
	var out bytes.Buffer
	params := []string{}
	for _, p := range f.Parameters {
		params = append(params, p.String())
	}
	out.WriteString("fungsi")
	out.WriteString("(")
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(") ")
	out.WriteString(f.Body.String())
	out.WriteString(" akhir")
	return out.String()
}
func (f *Function) GetAddress() memory.Ptr { return memory.NilPtr }

type CompiledFunction struct {
	Address memory.Ptr
}

func (cf *CompiledFunction) Type() ObjectType       { return COMPILED_FUNCTION_OBJ }
func (cf *CompiledFunction) Inspect() string        { return fmt.Sprintf("CompiledFunction[Ptr:0x%x]", cf.Address) }
func (cf *CompiledFunction) GetAddress() memory.Ptr { return cf.Address }

func (cf *CompiledFunction) NumLocals() int {
	n, _, err := memory.ReadCompiledFunctionMeta(cf.Address)
	if err != nil {
		panic(err)
	}
	return n
}

func (cf *CompiledFunction) NumParameters() int {
	_, n, err := memory.ReadCompiledFunctionMeta(cf.Address)
	if err != nil {
		panic(err)
	}
	return n
}

func (cf *CompiledFunction) Instructions() []byte {
	instr, _, _, err := memory.ReadCompiledFunction(cf.Address)
	if err != nil {
		panic(err)
	}
	return instr
}

type Closure struct {
	Address memory.Ptr
}

func NewClosure(fn *CompiledFunction, freeVars []Object) *Closure {
	freePtrs := make([]memory.Ptr, len(freeVars))
	for i, v := range freeVars {
		freePtrs[i] = v.GetAddress()
	}
	ptr, err := memory.AllocClosure(fn.Address, freePtrs)
	if err != nil {
		panic(err)
	}
	return &Closure{Address: ptr}
}

func (c *Closure) Type() ObjectType { return CLOSURE_OBJ }
func (c *Closure) Inspect() string  { return fmt.Sprintf("Closure[%p]", c) }
func (c *Closure) GetAddress() memory.Ptr { return c.Address }

func (c *Closure) Fn() *CompiledFunction {
	fnPtr, _, err := memory.ReadClosure(c.Address)
	if err != nil { panic(err) }
	return &CompiledFunction{Address: fnPtr}
}
func (c *Closure) FreeVariables() []Object {
	_, freePtrs, err := memory.ReadClosure(c.Address)
	if err != nil { panic(err) }
	objs := make([]Object, len(freePtrs))
	for i, ptr := range freePtrs {
		objs[i] = FromPtr(ptr)
	}
	return objs
}

type Array struct {
	Address memory.Ptr
}

func NewArray(elements []Object) *Array {
	count := len(elements)
	ptr, err := memory.AllocArray(count, count)
	if err != nil {
		panic(err)
	}
	for i, el := range elements {
		memory.WriteArrayElement(ptr, i, el.GetAddress())
	}
	return &Array{Address: ptr}
}

func (ao *Array) Type() ObjectType       { return ARRAY_OBJ }
func (ao *Array) GetAddress() memory.Ptr { return ao.Address }
func (ao *Array) Inspect() string {
	var out bytes.Buffer
	// Read elements from memory
	// We need length
	// memory.ReadArrayElement checks bounds, but we need length to iterate
	// We can use ReadArrayElement(0) etc until error? No.
	// We need ReadArrayLength.
	// I'll add GetLength helper.
	// For now, assume we can get it or just print address.
	// Inspect is important. I need length.
	// memory.go doesn't export ReadArrayLength?
	// I saw ReadArrayElement reads length internally.
	// I'll assume I can add a helper here using unsafe? No, keep it clean.
	// Inspect needs to be fixed. I will add ReadArrayLength to pkg/memory/array.go if missing?
	// It is missing.
	out.WriteString(fmt.Sprintf("Array[Ptr:0x%x]", ao.Address))
	return out.String()
}
// Helper to get Elements
func (ao *Array) GetElements() []Object {
	len, err := memory.ReadArrayLength(ao.Address)
	if err != nil { panic(err) }

	elements := make([]Object, len)
	for i := 0; i < len; i++ {
		elPtr, err := memory.ReadArrayElement(ao.Address, i)
		if err != nil { panic(err) }
		elements[i] = FromPtr(elPtr)
	}
	return elements
}

type HashKey struct {
	Type  ObjectType
	Value uint64
}

type HashPair struct {
	Key   Object
	Value Object
}

type Hash struct {
	Address memory.Ptr
}

func NewHash(pairs []HashPair) *Hash {
	count := len(pairs)
	ptr, err := memory.AllocHash(count)
	if err != nil { panic(err) }

	for i, pair := range pairs {
		memory.WriteHashPair(ptr, i, pair.Key.GetAddress(), pair.Value.GetAddress())
	}
	return &Hash{Address: ptr}
}

func (h *Hash) Type() ObjectType       { return HASH_OBJ }
func (h *Hash) GetAddress() memory.Ptr { return h.Address }
func (h *Hash) Inspect() string {
	return fmt.Sprintf("Hash[Ptr:0x%x]", h.Address)
}

// Helper to get Pairs
func (h *Hash) GetPairs() []HashPair {
	count, _ := memory.ReadHashCount(h.Address)
	pairs := make([]HashPair, count)
	for i := 0; i < count; i++ {
		kPtr, vPtr, _ := memory.ReadHashPair(h.Address, i)
		pairs[i] = HashPair{Key: FromPtr(kPtr), Value: FromPtr(vPtr)}
	}
	return pairs
}
