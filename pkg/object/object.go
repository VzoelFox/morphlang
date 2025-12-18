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
}

type Hashable interface {
	HashKey() HashKey
}

type Integer struct {
	Value   int64
	Address memory.Ptr
}

func (i *Integer) Type() ObjectType { return INTEGER_OBJ }
func (i *Integer) Inspect() string  { return fmt.Sprintf("%d", i.Value) }
func (i *Integer) HashKey() HashKey {
	return HashKey{Type: i.Type(), Value: uint64(i.Value)}
}

type Float struct {
	Value   float64
	Address memory.Ptr
}

func (f *Float) Type() ObjectType { return FLOAT_OBJ }
func (f *Float) Inspect() string {
	// Use %g for clean output (removes trailing zeros)
	return fmt.Sprintf("%g", f.Value)
}

type Boolean struct {
	Value bool
}

func (b *Boolean) Type() ObjectType { return BOOLEAN_OBJ }
func (b *Boolean) Inspect() string {
	if b.Value {
		return "benar"
	}
	return "salah"
}
func (b *Boolean) HashKey() HashKey {
	var value uint64

	if b.Value {
		value = 1
	} else {
		value = 0
	}

	return HashKey{Type: b.Type(), Value: value}
}

type Null struct{}

func (n *Null) Type() ObjectType { return NULL_OBJ }
func (n *Null) Inspect() string  { return "kosong" }

type ReturnValue struct {
	Value Object
}

func (rv *ReturnValue) Type() ObjectType { return RETURN_VALUE_OBJ }
func (rv *ReturnValue) Inspect() string  { return rv.Value.Inspect() }

const (
	ErrCodeSyntax          = "E001"
	ErrCodeUndefined       = "E002"
	ErrCodeTypeMismatch    = "E003"
	ErrCodeUncheckedError  = "E004"
	ErrCodeRuntime         = "E005" // Generic Runtime (DivByZero etc)
	ErrCodeIndexOutOfBounds = "E006"
	ErrCodeInvalidOp       = "E007"
	ErrCodeMissingArgs     = "E008"
	ErrCodeTooManyArgs     = "E009"
)

type Error struct {
	Message    string
	Code       string
	Line       int
	Column     int
	File       string
	StackTrace []string
	Hint       string
}

func (e *Error) Type() ObjectType { return ERROR_OBJ }
func (e *Error) Inspect() string {
	var out bytes.Buffer

	// Basic Error Header
	if e.Code != "" {
		out.WriteString(fmt.Sprintf("Error [%s] di [%s:%d:%d]:\n", e.Code, e.File, e.Line, e.Column))
	} else {
		out.WriteString(fmt.Sprintf("Error di [%s:%d:%d]:\n", e.File, e.Line, e.Column))
	}
	out.WriteString(fmt.Sprintf("  %s\n", e.Message))

	if len(e.StackTrace) > 0 {
		out.WriteString("\nStack trace:\n")
		for _, trace := range e.StackTrace {
			out.WriteString(fmt.Sprintf("  di %s\n", trace))
		}
	}

	if e.Hint != "" {
		out.WriteString(fmt.Sprintf("\nHint: %s", e.Hint))
	}

	return out.String()
}

type Channel struct {
	Value chan Object
}

func (c *Channel) Type() ObjectType { return CHANNEL_OBJ }
func (c *Channel) Inspect() string  { return fmt.Sprintf("saluran[%p]", c.Value) }

type Thread struct {
	Result chan Object
}

func (t *Thread) Type() ObjectType { return THREAD_OBJ }
func (t *Thread) Inspect() string  { return fmt.Sprintf("utas[%p]", t.Result) }

type Time struct {
	Value time.Time
}

func (t *Time) Type() ObjectType { return TIME_OBJ }
func (t *Time) Inspect() string  { return t.Value.Format(time.RFC3339) }

type Mutex struct {
	Mu sync.Mutex
}

func (m *Mutex) Type() ObjectType { return MUTEX_OBJ }
func (m *Mutex) Inspect() string  { return fmt.Sprintf("mutex[%p]", &m.Mu) }

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

type File struct {
	File *os.File
	Mode string
}

func (f *File) Type() ObjectType { return FILE_OBJ }
func (f *File) Inspect() string {
	return fmt.Sprintf("file[%s]", f.File.Name())
}

type Pointer struct {
	Address uint64
}

func (p *Pointer) Type() ObjectType { return POINTER_OBJ }
func (p *Pointer) Inspect() string {
	return fmt.Sprintf("ptr[0x%x]", p.Address)
}

type BuiltinFunction func(args ...Object) Object

type Builtin struct {
	Fn BuiltinFunction
}

func (b *Builtin) Type() ObjectType { return BUILTIN_OBJ }
func (b *Builtin) Inspect() string  { return "builtin function" }

type String struct {
	Value string
}

func (s *String) Type() ObjectType { return STRING_OBJ }
func (s *String) Inspect() string  { return s.Value }
func (s *String) HashKey() HashKey {
	h := fnv.New64a()
	h.Write([]byte(s.Value))

	return HashKey{Type: s.Type(), Value: h.Sum64()}
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

type CompiledFunction struct {
	Instructions  []byte
	NumLocals     int
	NumParameters int
}

func (cf *CompiledFunction) Type() ObjectType { return COMPILED_FUNCTION_OBJ }
func (cf *CompiledFunction) Inspect() string {
	return fmt.Sprintf("CompiledFunction[%d]", len(cf.Instructions))
}

type Closure struct {
	Fn            *CompiledFunction
	FreeVariables []Object
}

func (c *Closure) Type() ObjectType { return CLOSURE_OBJ }
func (c *Closure) Inspect() string {
	return fmt.Sprintf("Closure[%p]", c)
}

type Array struct {
	Elements []Object
}

func (ao *Array) Type() ObjectType { return ARRAY_OBJ }
func (ao *Array) Inspect() string {
	var out bytes.Buffer

	elements := []string{}
	for _, e := range ao.Elements {
		elements = append(elements, e.Inspect())
	}

	out.WriteString("[")
	out.WriteString(strings.Join(elements, ", "))
	out.WriteString("]")

	return out.String()
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
	Pairs map[HashKey]HashPair
}

func (h *Hash) Type() ObjectType { return HASH_OBJ }
func (h *Hash) Inspect() string {
	var out bytes.Buffer

	pairs := []string{}
	for _, pair := range h.Pairs {
		pairs = append(pairs, fmt.Sprintf("%s: %s", pair.Key.Inspect(), pair.Value.Inspect()))
	}

	out.WriteString("{")
	out.WriteString(strings.Join(pairs, ", "))
	out.WriteString("}")

	return out.String()
}
