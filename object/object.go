package object

import (
	"bytes"
	"fmt"
	"monkey/ast"
	"strings"
)

// ObjectType es el tipo de dato base para todos los objetos.
type ObjectType string

// constantes para los tipos de datos del lenguaje interpretado.
const (
	INTEGER_OBJ      = "INTEGER"
	BOOLEAN_OBJ      = "BOOLEAN"
	NULL_OBJ         = "NULL"
	RETURN_VALUE_OBJ = "RETURN_VALUE"
	ERROR_OBJ        = "ERROR"
	FUNCTION_OBJ     = "FUNCTION"
	STRING_OBJ       = "STRING"
	BUILTIN_OBJ      = "BUILTIN"
)

// Object es una interface que comprende todos los valores
// que pueden almacenar los tipos de datos definidos en el
// lenguaje de programación interpretado. Se le puede llamar:
// 1. Sistema de objetos.
// 2. Sistema de tipos.
// 3. Representación de Objetos.
// En resumen: es el tipo de dato "nativo" que nuestro
// lenguaje interpretado será capaz de exponer. El objetivo
// de tener un sistema de tipos es "envolver" el valor real
// que el lenguaje necesita.
type Object interface {
	Type() ObjectType
	Inspect() string
}

type Integer struct {
	Value int64
}

// Type() retorna el tipo de objeto.
func (i *Integer) Type() ObjectType {
	return INTEGER_OBJ
}

// Inspect() retorna el valor literal.
func (i *Integer) Inspect() string {
	return fmt.Sprintf("%d", i.Value)
}

// Tipo de dato Boolean que soportará nuestro
// lenguaje interpretado Monkey. (iox en la segunda implementación)
type Boolean struct {
	Value bool
}

func (b *Boolean) Type() ObjectType { return BOOLEAN_OBJ }
func (b *Boolean) Inspect() string {
	return fmt.Sprintf("%t", b.Value)
}

// Tipo de dato Null soportado para nuestro lenguaje Monkey
// este tipo es una estructura vacía.
type Null struct {
}

// Cumple con la interface Object
func (n *Null) Type() ObjectType { return NULL_OBJ }
func (n *Null) Inspect() string  { return "null" }

// Objeto Return
type ReturnValue struct {
	Value Object
}

func (rv *ReturnValue) Type() ObjectType { return RETURN_VALUE_OBJ }
func (rv *ReturnValue) Inspect() string  { return rv.Value.Inspect() }

type Error struct {
	Message string
}

func (e *Error) Type() ObjectType { return ERROR_OBJ }
func (e *Error) Inspect() string  { return "ERROR: " + e.Message }

// Objeto función.
type Function struct {
	Parameters []*ast.Identifier
	Body       *ast.BlockStatement
	Env        *Environment
}

func (f *Function) Type() ObjectType { return FUNCTION_OBJ }
func (f *Function) Inspect() string {
	var out bytes.Buffer
	params := []string{}
	for _, p := range f.Parameters {
		params = append(params, p.String())
	}
	out.WriteString("fn")
	out.WriteString("(")
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(") {\n")
	out.WriteString(f.Body.String())
	out.WriteString("\n}")
	return out.String()
}

// Objeto String
type String struct {
	Value string
}

func (s *String) Type() ObjectType { return STRING_OBJ }
func (s *String) Inspect() string  { return s.Value }

// ... significa que la función acepta 0+ parametros
// del tipo especificado a la derecha.
type BuiltinFunction func(args ...Object) Object

// Objeto Builtin
type Builtin struct {
	Fn BuiltinFunction
}

func (b *Builtin) Type() ObjectType { return BUILTIN_OBJ }
func (b *Builtin) Inspect() string  { return "builtin function" }
