# swalker

extract value from map/struct by dot notated syntax like "Foo.Bar[0].Baz"

## Usage

```
package main

import (
	"fmt"
	"github.com/yuroyoro/swalker"
)

type A struct {
	Foo *B
}
type B struct {
	Bar []*C
}
type C struct {
	Hoge string
}

func main() {
	obj := A{Foo: &B{Bar: []*C{&C{Hoge: "aaa"}, &C{Hoge: "bbb"}}}}

	// Read Foo from obj
	v, err := swalker.Read("Foo", obj)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%T : %v\n", v, v) // -> *B

	// Read Foo.Bar from obj
	v, err = swalker.Read("Foo.Bar", obj)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%T : %v\n", v, v) // -> []*C

	// Read Foo.Bar[0] from obj
	v, err = swalker.Read("Foo.Bar[0]", obj)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%T : %v\n", v, v) // -> *C

	// Read Foo.Bar[0].Hoge from obj
	v, err = swalker.Read("Foo.Bar[0].Hoge", obj)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%T : %v\n", v, v) // -> "aaaa"

	// Write "xxxx" to obj's Foo.Bar[0].Hoge

	err = swalker.Write("Foo.Bar[0].Hoge", obj, "xxxx")
	if err != nil {
		panic(err)
	}
	nv := obj.Foo.Bar[0].Hoge
	fmt.Printf("%T : %v\n", nv, nv) // -> "xxxx"
}
```

## License

MIT

## Author

Tomothio Ozaki (@yuroyoro)
