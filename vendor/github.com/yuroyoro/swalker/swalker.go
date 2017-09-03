package swalker

import (
	"reflect"
)

// ExpType : type of Expression
type ExpType int

const (
	// Property : ExpType
	Property ExpType = iota
	// Indexing : Exptype
	Indexing
)

// Expressions : parsed expressions
type Expressions []*Expression

// Expression : element of expressions
type Expression struct {
	Type  ExpType
	Name  string
	Index int
}

func unwrap(v *reflect.Value) *reflect.Value {
	if v.Kind() == reflect.Interface {
		org := v.Elem() //  Get rid of the wrapping interface
		return &org
	}
	return v
}

func indirecte(v *reflect.Value) *reflect.Value {
	indirected := reflect.Indirect(*v)
	return &indirected
}

func isNil(v *reflect.Value) bool {
	if v == nil {
		return true
	}

	v = indirecte(v)
	switch v.Kind() {
	case reflect.Invalid:
		return true
	case reflect.Array, reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Slice:
		return v.IsNil()
	}
	return false
}
