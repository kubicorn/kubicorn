package swalker

import (
	"reflect"
	"testing"
)

func TestWriteHash(t *testing.T) {

	obj := makeHash()
	hash := obj.(map[string]interface{})
	fooval := hash["foo"]
	foo0 := fooval.([]interface{})[0]
	bar1 := foo0.(map[string]interface{})["bar1"]
	hbar1 := bar1.(map[string]interface{})

	// Write(obj, "foo.bar1.hoge", "xxxx") string value
	err := Write("foo[0].bar1.hoge", obj, "xxxx")
	if err != nil {
		t.Fatalf("unexpected error : [%s]", err)
	}
	if hbar1["hoge"] != "xxxx" {
		t.Fatalf("could not write : value %v", hbar1["hoge"])
	}
	// Write(obj, "foo.bar1.hoge", "xxxx") new key
	err = Write("foo[0].bar1.fuga", obj, "wwww")
	if err != nil {
		t.Fatalf("unexpected error : [%s]", err)
	}
	if hbar1["fuga"] != "wwww" {
		t.Fatalf("could not write : value %v", hbar1["hoge"])
	}

	// Write(obj, "foo.bar1", mapvalue) map value
	newhoge := map[string]interface{}{"hoge": "yyyy"}
	err = Write("foo[0].bar1", obj, newhoge)
	if err != nil {
		t.Fatalf("unexpected error : [%s]", err)
	}
	bar1 = foo0.(map[string]interface{})["bar1"]
	if reflect.DeepEqual(bar1, newhoge) == false {
		t.Fatalf("could not write : value %v", bar1)
	}

	// Write(obj, "noooo[1]", 99) replace slice value
	var nv interface{} = 99.0
	err = Write("noooo[1]", obj, nv)
	if err != nil {
		t.Fatalf("unexpected error : [%s]", err)
	}
	noooo := hash["noooo"].([]interface{})
	if reflect.DeepEqual(nv, noooo[1]) == false {
		t.Fatalf("unexpected value : [%v]", noooo[1])
	}

	// Write("foo[3]", obj, "xxxx") -> index out of range
	err = Write("foo[3]", obj, "xxxx")
	if err == nil {
		t.Fatal("should return error")
	}
	if err.Error() != "field foo len 2 : index 3 is out of range" {
		t.Fatalf("unexpected error : [%s]", err)
	}

	// Write("foo[0].bar[1]", obj, "xxxx") -> index access to non-slice
	err = Write("foo[0].bar1[1]", obj, "xxxx")
	if err == nil {
		t.Fatal("should return error")
	}
	if err.Error() != "field bar1 is not array or slice : map" {
		t.Fatalf("unexpected error : [%s]", err)
	}

	// Write("foo[0].bar3.hoge") -> nil access
	err = Write("foo[0].bar3.hoge", obj, "xxxx")
	if err == nil {
		t.Fatal("should return error")
	}
	if err.Error() != "field bar3 is nil" {
		t.Fatalf("unexpected error : [%s]", err)
	}
}
func TestWriteStruct(t *testing.T) {
	obj := makeStruct()

	// Write(obj, "Foo.Bar1.Hoge", "xxxx") string value
	err := Write("Foo[0].Bar1.Hoge", obj, "xxxx")
	if err != nil {
		t.Fatalf("unexpected error : [%s]", err)
	}
	if obj.Foo[0].Bar1.Hoge != "xxxx" {
		t.Fatalf("could not write : value %v", obj.Foo[0].Bar1.Hoge)
	}

	// Write(obj, "Foo.Bar1", mapvalue) map value
	newHoge := &Hoge{Hoge: "yyyy"}
	err = Write("Foo[0].Bar1", obj, newHoge)
	if err != nil {
		t.Fatalf("unexpected error : [%s]", err)
	}
	if reflect.DeepEqual(obj.Foo[0].Bar1, newHoge) == false {
		t.Fatalf("could not write : value %v", obj.Foo[0].Bar1)
	}

	// Write(obj, "noooo[1]", 99) replace slice value
	err = Write("Noooo[1]", obj, 99)
	if err != nil {
		t.Fatalf("unexpected error : [%s]", err)
	}
	if reflect.DeepEqual(99, obj.Noooo[1]) == false {
		t.Fatalf("unexpected value : [%v]", obj.Noooo[1])
	}

	// Write("Foo[3]", obj, "xxxx") -> index out of range
	err = Write("Foo[3]", obj, "xxxx")
	if err == nil {
		t.Fatal("should return error")
	}
	if err.Error() != "field Foo len 2 : index 3 is out of range" {
		t.Fatalf("unexpected error : [%s]", err)
	}

	// Write("Foo[0].Bar[1]", obj, "xxxx") -> index access to non-slice
	err = Write("Foo[0].Bar1[1]", obj, "xxxx")
	if err == nil {
		t.Fatal("should return error")
	}
	if err.Error() != "field Bar1 is not array or slice : ptr" {
		t.Fatalf("unexpected error : [%s]", err)
	}

	// Write(obj, "Foo.Bar1.Fuga", "xxxx") unknown member
	err = Write("Foo[0].Bar1.Fuga", obj, "wwww")
	if err == nil {
		t.Fatal("should return error")
	}
	if err.Error() != "Fuga is not a field of struct type swalker.Hoge" {
		t.Fatalf("unexpected error : [%s]", err)
	}

	// Write("Foo[0].Bar3.Hoge") -> nil access
	err = Write("Foo[0].Bar3.Hoge", obj, "xxxx")
	if err == nil {
		t.Fatal("should return error")
	}
	if err.Error() != "field Bar3 is nil" {
		t.Fatalf("unexpected error : [%s]", err)
	}

	// Write("Ababa", obj."xxxx") -> type mismatch
	err = Write("Ababa", obj, "xxxx")
	if err == nil {
		t.Fatal("should return error")
	}
	if err.Error() != "cannot write string to field Ababa(bool) : xxxx" {
		t.Fatalf("unexpected error : [%s]", err)
	}

	// Write("Noooo[1]", obj."xxxx") -> type mismatch
	err = Write("Noooo[1]", obj, "xxxx")
	if err == nil {
		t.Fatal("should return error")
	}
	if err.Error() != "cannot write string to field Noooo[1](int) : xxxx" {
		t.Fatalf("unexpected error : [%s]", err)
	}
}
