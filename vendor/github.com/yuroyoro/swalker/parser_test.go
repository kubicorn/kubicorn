package swalker

import (
	"reflect"
	"testing"
)

func TestParser(t *testing.T) {

	// Parse("Foo.Bar[1].Baz") -> valid
	exps, err := Parse("Foo.Bar[1].Baz")
	if err != nil {
		t.Fatalf("unexpected error : [%s]", err)
	}

	expected := Expressions{
		&Expression{Type: Property, Name: "Foo"},
		&Expression{Type: Indexing, Name: "Bar", Index: 1},
		&Expression{Type: Property, Name: "Baz"},
	}
	if reflect.DeepEqual(exps, expected) == false {
		t.Fatalf(`invalid value: expected %+v : actual %+v`, expected, exps)
	}

	// Parse("Fo   o.Bar") -> invalid
	exps, err = Parse("Fo   o..Bar")
	if err == nil {
		t.Fatal("should return error")
	}
	if err.Error() != `Invalid expression "Fo   o"` {
		t.Fatalf("unexpected error : [%s]", err)
	}

	// Parse("Foo..Bar") -> invalid
	exps, err = Parse("Foo..Bar")
	if err == nil {
		t.Fatal("should return error")
	}
	if err.Error() != `Invalid expression ""` {
		t.Fatalf("unexpected error : [%s]", err)
	}
	// Parse(".Foo.Bar") -> invalid
	exps, err = Parse(".Foo.Bar")
	if err == nil {
		t.Fatal("should return error")
	}
	if err.Error() != `Invalid expression ""` {
		t.Fatalf("unexpected error : [%s]", err)
	}
	// Parse("[1].Foo.Bar") -> invalid
	exps, err = Parse("[1].Foo.Bar")
	if err == nil {
		t.Fatal("should return error")
	}
	if err.Error() != `Invalid expression "[1]"` {
		t.Fatalf("unexpected error : [%s]", err)
	}
	// Parse("Foo[xxx].Bar") -> invalid
	exps, err = Parse("Foo[xxx].Bar")
	if err == nil {
		t.Fatal("should return error")
	}
	if err.Error() != `Invalid expression "Foo[xxx]"` {
		t.Fatalf("unexpected error : [%s]", err)
	}
	// Parse("Foo[.Bar") -> invalid
	exps, err = Parse("Foo[.Bar")
	if err == nil {
		t.Fatal("should return error")
	}
	if err.Error() != `Invalid expression "Foo["` {
		t.Fatalf("unexpected error : [%s]", err)
	}
	// Parse("Foo{}.Bar") -> invalid
	exps, err = Parse("Foo{}.Bar")
	if err == nil {
		t.Fatal("should return error")
	}
	if err.Error() != `Invalid expression "Foo{}"` {
		t.Fatalf("unexpected error : [%s]", err)
	}
}
