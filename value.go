package main

import (
	"bytes"
	"fmt"
	"strconv"
)

const (
	kValueTypeString = iota
	kValueTypeNumber
	kValueTypeBoolean
	kValueTypeNull
	kValueTypeObject
	kValueTypeList
)

type ValueType int

// Value object represents a generic typped value
type Value struct {
	Type    ValueType
	String  string
	Number  float64
	Boolean bool
	Object  *Object
	List    *List
}

type Object struct {
	Value map[string]Value
}

type List struct {
	Value []Value
}

func (v ValueType) GetName() string {
	switch v {
	case kValueTypeString:
		return "string"
	case kValueTypeNumber:
		return "number"
	case kValueTypeBoolean:
		return "boolean"
	case kValueTypeNull:
		return "null"
	case kValueTypeObject:
		return "object"
	case kValueTypeList:
		return "list"
	default:
		panic("unreachable!")
	}
}

func NewNull() Value {
	return Value{Type: kValueTypeNull}
}

func NewObject() *Object {
	return &Object{Value: make(map[string]Value)}
}

func NewList() *List {
	return &List{Value: []Value{}}
}

func indent(buf *bytes.Buffer, level int) *bytes.Buffer {
	for i := 0; i < level; i++ {
		buf.WriteString("  ")
	}

	return buf
}

func (v *Value) toJson(b *bytes.Buffer, indent int) {

	switch v.Type {
	case kValueTypeString:
		b.WriteString(PrintQuotedString(v.String))
	case kValueTypeNumber:
		b.WriteString(fmt.Sprintf("%s", strconv.FormatFloat(v.Number, 'f', -1, 64)))
	case kValueTypeBoolean:
		b.WriteString(BooleanToString(v.Boolean))
	case kValueTypeNull:
		b.WriteString("null")
	case kValueTypeObject:
		v.Object.toJson(b, indent)
	case kValueTypeList:
		v.List.toJson(b, indent)
	default:
		panic("unreachable!")
	}
}

func (obj *Object) toJson(b *bytes.Buffer, idt int) {
	b.WriteString("{\n")

	idx := 0
	for k, v := range obj.Value {
		indent(b, idt+1).WriteString(PrintQuotedString(k))
		b.WriteString(" : ")
		v.toJson(b, idt+1)
		if idx < len(obj.Value)-1 {
			b.WriteString(",")
		}
		b.WriteString("\n")
		idx++
	}

	indent(b, idt).WriteString("}")
}

func (l *List) toJson(b *bytes.Buffer, idt int) {
	b.WriteString("[\n")

	idx := 0
	for _, v := range l.Value {
		indent(b, idt+1)
		v.toJson(b, idt+1)
		if idx < len(l.Value)-1 {
			b.WriteString(",")
		}
		b.WriteString("\n")
		idx++
	}

	indent(b, idt).WriteString("]")
}

func (v *Value) ToJson() string {
	b := bytes.Buffer{}
	v.toJson(&b, 0)
	return b.String()
}
