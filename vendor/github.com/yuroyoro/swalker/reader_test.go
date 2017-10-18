package swalker

import (
	"reflect"

	"encoding/json"
	"testing"
)

type Foo struct {
	Foo   []*Bar
	Noooo []int
	Moooo []interface{}
	Ababa bool
	Cccc  float64
	xxxx  string
}

type Bar struct {
	Bar1 *Hoge
	Bar2 *Hoge
	Bar3 *Hoge
}
type Hoge struct {
	Hoge string
}

func makeStruct() *Foo {
	return &Foo{
		Foo: []*Bar{
			{
				Bar1: &Hoge{Hoge: "aaaa"},
				Bar2: &Hoge{Hoge: "bbbb"},
				Bar3: nil,
			},
			{
				Bar1: &Hoge{Hoge: "cccc"},
				Bar2: &Hoge{Hoge: "dddd"},
			},
		},
		Noooo: []int{33, 44, 55},
		Moooo: []interface{}{1},
		Ababa: true,
		Cccc:  1.234,
		xxxx:  "unexported",
	}
}

func makeHash() interface{} {
	raw := `
{ "foo" :[
   {
     "bar1" : { "hoge" : "aaaa"},
     "bar2" : { "hoge" : "bbbb"},
     "bar3" : null
    },
    {
     "bar1" : { "hoge" : "cccc"},
     "bar2" : { "hoge" : "dddd"}
    }
  ],
  "noooo" : [33,44,55],
  "ababa" : true,
  "cccc" : 1.234
}`
	var obj interface{}
	err := json.Unmarshal([]byte(raw), &obj)
	if err != nil {
		panic(err)
	}

	return obj

}

func TestReadHash(t *testing.T) {
	obj := makeHash()
	hash := obj.(map[string]interface{})

	// Read("foo") -> []interface{}
	ret, err := Read("foo", obj)
	if err != nil {
		t.Fatal(err)
	}

	v, ok := ret.([]interface{})
	if !ok {
		t.Fatalf(`Read("foo") returns invalid type : %T`, ret)
	}

	fooval := hash["foo"]
	if reflect.DeepEqual(fooval, v) == false {
		t.Fatalf(`Read("foo") returns invalid value: expected %+v : actual %+v`, fooval, v)
	}

	// Read("foo[0]") -> map[string]interface{}
	ret, err = Read("foo[0]", obj)
	if err != nil {
		t.Fatal(err)
	}

	v1, ok := ret.(map[string]interface{})
	if !ok {
		t.Fatalf(`Read("foo[0]") returns invalid type : %T`, ret)
	}

	foo0 := fooval.([]interface{})[0]
	if reflect.DeepEqual(foo0, v1) == false {
		t.Fatalf(`Read("foo[0]") returns invalid value: expected %+v : actual %+v`, foo0, v1)
	}

	// Read("foo[0].bar1") -> map[string]interface{}
	ret, err = Read("foo[0].bar1", obj)
	if err != nil {
		t.Fatal(err)
	}

	v2, ok := ret.(map[string]interface{})
	if !ok {
		t.Fatalf(`Read("foo[0].bar1") returns invalid type : %T`, ret)
	}

	bar1 := foo0.(map[string]interface{})["bar1"]
	if reflect.DeepEqual(bar1, v2) == false {
		t.Fatalf(`Read("foo[0].bar1") returns invalid value: expected %+v : actual %+v`, bar1, v2)
	}

	// Read("foo[0].bar1.hoge") -> sttring
	ret, err = Read("foo[0].bar1.hoge", obj)
	if err != nil {
		t.Fatal(err)
	}

	v3, ok := ret.(string)
	if !ok {
		t.Fatalf(`Read("foo[0].bar1.hoge") returns invalid type : %T`, ret)
	}

	hoge := bar1.(map[string]interface{})["hoge"]
	if reflect.DeepEqual(hoge, v3) == false {
		t.Fatalf(`Read("foo[0].bar1.hoge") returns invalid value: expected %+v : actual %+v`, hoge, v3)
	}

	// Read("foo[0].bar1.hoge") -> sttring
	ret, err = Read("foo[0].bar1.hoge", obj)
	if err != nil {
		t.Fatal(err)
	}

	// Read("foo[3]") -> index out of range
	ret, err = Read("foo[3]", obj)
	if err == nil {
		t.Fatal("should return error")
	}
	if err.Error() != "field foo len 2 : index 3 is out of range" {
		t.Fatalf("unexpected error : [%s]", err)
	}

	// Read("foo[0].bar[1]") -> index access to non-slice
	ret, err = Read("foo[0].bar1[1]", obj)
	if err == nil {
		t.Fatal("should return error")
	}
	if err.Error() != "field bar1 is not array or slice : map" {
		t.Fatalf("unexpected error : [%s]", err)
	}

	// Read("foo[1].aaaa") -> unknown key
	ret, err = Read("foo[0].aaaa", obj)
	if err == nil {
		t.Fatal("should return error")
	}
	if err.Error() != "map has no entry for key \"aaaa\"" {
		t.Fatalf("unexpected error : [%s]", err)
	}

	// Read("foo[0].bar3.hoge") -> nil access
	ret, err = Read("foo[0].bar3.hoge", obj)
	if err == nil {
		t.Fatal("should return error")
	}
	if err.Error() != "field bar3 is nil" {
		t.Fatalf("unexpected error : [%s]", err)
	}

	// ReadString("foo[0].bar1.hoge") -> string
	str, err := ReadString("foo[0].bar1.hoge", obj)
	if err != nil {
		t.Fatalf("unexpected error : [%s]", err)
	}
	if str != "aaaa" {
		t.Fatalf("unexpected value : [%s]", str)
	}

	// ReadSlice("noooo") -> []interface{}
	arr, err := ReadSlice("noooo", obj)
	if err != nil {
		t.Fatalf("unexpected error : [%s]", err)
	}
	if reflect.DeepEqual(arr, hash["noooo"]) == false {
		t.Fatalf("unexpected value : [%v]", arr)
	}

	// ReadInt("noooo[1]") -> int
	// iv, err := ReadInt("noooo[1]", obj)
	// if err != nil {
	// t.Fatalf("unexpected error : [%s]", err)
	// }
	// if iv != 44 {
	// t.Fatalf("unexpected value : [%v]", arr)
	// }

	// ReadFloat("noooo[1]") -> int
	iv, err := ReadFloat("noooo[1]", obj)
	if err != nil {
		t.Fatalf("unexpected error : [%s]", err)
	}
	if iv != 44 {
		t.Fatalf("unexpected value : [%v]", arr)
	}

	// ReadBool("ababa") -> bool
	bv, err := ReadBool("ababa", obj)
	if err != nil {
		t.Fatalf("unexpected error : [%s]", err)
	}
	if !bv {
		t.Fatalf("unexpected value : [%v]", arr)
	}

	// ReadString("noooo[1]") -> type mismatch
	ret, err = ReadString("noooo[1]", obj)
	if err == nil {
		t.Fatal("should return error")
	}
	if err.Error() != "extracted value is not string : float64" {
		t.Fatalf("unexpected error : [%s]", err)
	}

}

func TestReadStruct(t *testing.T) {
	obj := makeStruct()

	// Read("Foo") -> []*Bar
	ret, err := Read("Foo", obj)
	if err != nil {
		t.Fatal(err)
	}

	v, ok := ret.([]*Bar)
	if !ok {
		t.Fatalf(`Read("Foo") returns invalid type : %T`, ret)
	}

	fooval := obj.Foo
	if reflect.DeepEqual(fooval, v) == false {
		t.Fatalf(`Read("Foo") returns invalid value: expected %+v : actual %+v`, fooval, v)
	}

	// Read("Foo[0]") -> *Bar
	ret, err = Read("Foo[0]", obj)
	if err != nil {
		t.Fatal(err)
	}

	v1, ok := ret.(*Bar)
	if !ok {
		t.Fatalf(`Read("Foo[0]") returns invalid type : %T`, ret)
	}

	foo0 := fooval[0]
	if reflect.DeepEqual(foo0, v1) == false {
		t.Fatalf(`Read("Foo[0]") returns invalid value: expected %+v : actual %+v`, foo0, v1)
	}

	// Read("Foo[0].Bar1") -> *Hoge
	ret, err = Read("Foo[0].Bar1", obj)
	if err != nil {
		t.Fatal(err)
	}

	v2, ok := ret.(*Hoge)
	if !ok {
		t.Fatalf(`Read("Foo[0].Bar1") returns invalid type : %T`, ret)
	}

	Bar1 := foo0.Bar1
	if reflect.DeepEqual(Bar1, v2) == false {
		t.Fatalf(`Read("Foo[0].Bar1") returns invalid value: expected %+v : actual %+v`, Bar1, v2)
	}

	// Read("Foo[0].Bar1.Hoge") -> sttring
	ret, err = Read("Foo[0].Bar1.Hoge", obj)
	if err != nil {
		t.Fatal(err)
	}

	v3, ok := ret.(string)
	if !ok {
		t.Fatalf(`Read("Foo[0].Bar1.Hoge") returns invalid type : %T`, ret)
	}

	Hoge := Bar1.Hoge
	if reflect.DeepEqual(Hoge, v3) == false {
		t.Fatalf(`Read("Foo[0].Bar1.Hoge") returns invalid value: expected %+v : actual %+v`, Hoge, v3)
	}

	// Read("Foo[0].Bar1.Hoge") -> sttring
	ret, err = Read("Foo[0].Bar1.Hoge", obj)
	if err != nil {
		t.Fatal(err)
	}

	// Read("Foo[3]") -> index out of range
	ret, err = Read("Foo[3]", obj)
	if err == nil {
		t.Fatal("should return error")
	}
	if err.Error() != "field Foo len 2 : index 3 is out of range" {
		t.Fatalf("unexpected error : [%s]", err)
	}

	// Read("Foo[0].Bar[1]") -> index access to non-slice
	ret, err = Read("Foo[0].Bar1[1]", obj)
	if err == nil {
		t.Fatal("should return error")
	}
	if err.Error() != "field Bar1 is not array or slice : ptr" {
		t.Fatalf("unexpected error : [%s]", err)
	}

	// Read("Foo[1].aaaa") -> unknown key
	ret, err = Read("Foo[0].Aaaa", obj)
	if err == nil {
		t.Fatal("should return error")
	}
	if err.Error() != "Aaaa is not a field of struct type swalker.Bar" {
		t.Fatalf("unexpected error : [%s]", err)
	}

	// Read("Foo[0].Bar3.Hoge") -> nil access
	ret, err = Read("Foo[0].Bar3.Hoge", obj)
	if err == nil {
		t.Fatal("should return error")
	}
	if err.Error() != "field Bar3 is nil" {
		t.Fatalf("unexpected error : [%s]", err)
	}
	// ReadString("Foo[0].Bar1.Hoge") -> string
	str, err := ReadString("Foo[0].Bar1.Hoge", obj)
	if err != nil {
		t.Fatalf("unexpected error : [%s]", err)
	}
	if str != "aaaa" {
		t.Fatalf("unexpected value : [%s]", str)
	}

	// ReadSlice("Moooo") -> []interface{}
	arr, err := ReadSlice("Moooo", obj)
	if err != nil {
		t.Fatalf("unexpected error : [%s]", err)
	}
	if reflect.DeepEqual(arr, obj.Moooo) == false {
		t.Fatalf("unexpected value : [%v]", arr)
	}

	// ReadInt("Noooo[1]") -> int
	iv, err := ReadInt("Noooo[1]", obj)
	if err != nil {
		t.Fatalf("unexpected error : [%s]", err)
	}
	if iv != 44 {
		t.Fatalf("unexpected value : [%v]", arr)
	}

	// ReadFloat("Cccc") -> float64
	fv, err := ReadFloat("Cccc", obj)
	if err != nil {
		t.Fatalf("unexpected error : [%s]", err)
	}
	if fv != 1.234 {
		t.Fatalf("unexpected value : [%v]", fv)
	}

	// ReadBool("Ababa") -> bool
	bv, err := ReadBool("Ababa", obj)
	if err != nil {
		t.Fatalf("unexpected error : [%s]", err)
	}
	if !bv {
		t.Fatalf("unexpected value : [%v]", arr)
	}

	// ReadString("Noooo[1]") -> type mismatch
	ret, err = ReadString("Noooo[1]", obj)
	if err == nil {
		t.Fatal("should return error")
	}
	if err.Error() != "extracted value is not string : int" {
		t.Fatalf("unexpected error : [%s]", err)
	}

}
