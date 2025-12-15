package object

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/VzoelFox/morphlang/pkg/parser"
)

type ObjectType string

const (
	INTEGER_OBJ      = "INTEGER"
	BOOLEAN_OBJ      = "BOOLEAN"
	NULL_OBJ         = "NULL"
	RETURN_VALUE_OBJ = "RETURN_VALUE"
	ERROR_OBJ        = "ERROR"
	FUNCTION_OBJ     = "FUNCTION"
	STRING_OBJ       = "STRING"
)

type Object interface {
	Type() ObjectType
	Inspect() string
}

type Integer struct {
	Value int64
}

func (i *Integer) Type() ObjectType { return INTEGER_OBJ }
func (i *Integer) Inspect() string  { return fmt.Sprintf("%d", i.Value) }

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

type Null struct{}

func (n *Null) Type() ObjectType { return NULL_OBJ }
func (n *Null) Inspect() string  { return "kosong" }

type ReturnValue struct {
	Value Object
}

func (rv *ReturnValue) Type() ObjectType { return RETURN_VALUE_OBJ }
func (rv *ReturnValue) Inspect() string  { return rv.Value.Inspect() }

type Error struct {
	Message    string
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
	out.WriteString(fmt.Sprintf("Error di [%s:%d:%d]:\n", e.File, e.Line, e.Column))
	out.WriteString(fmt.Sprintf("  %s\n", e.Message))

	// Arrow Pointer (Context) logic would theoretically go here if we had source access
	// For now, we stick to the formatted output required by AGENTS.md

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

type String struct {
	Value string
}

func (s *String) Type() ObjectType { return STRING_OBJ }
func (s *String) Inspect() string  { return s.Value }

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
