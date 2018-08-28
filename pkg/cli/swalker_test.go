// Copyright © 2018 The Kubicorn Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cli

import (
	"fmt"
	"testing"

	"github.com/yuroyoro/swalker"
)

func TestSwalkerWrite(t *testing.T) {
	type myStruct struct {
		Int    int
		Int8   int8
		Int16  int16
		Int32  int32
		Int64  int64
		Uint   uint
		Uint8  uint8
		Uint16 uint16
		Uint32 uint32
		Uint64 uint64
		String string
		Bool   bool
	}

	s := &myStruct{
		Int:    1,
		Int8:   2,
		Int16:  3,
		Int32:  4,
		Int64:  5,
		Uint:   6,
		Uint8:  7,
		Uint16: 8,
		Uint32: 9,
		Uint64: 10,
	}

	keys := []string{
		"Int",
		"Int8",
		"Int16",
		"Int32",
		"Int64",
		"Uint",
		"Uint8",
		"Uint16",
		"Uint32",
		"Uint64",
	}

	for _, key := range keys {
		t.Run(key, func(t *testing.T) {
			if err := SwalkerWrite(key, s, "100"); err != nil {
				t.Errorf("unexpected write error: %v", err)
			}

			a, err := swalker.Read(key, s)
			if err != nil {
				t.Errorf("unexpected read error: %v", err)
			}
			if e, a := "100", fmt.Sprintf("%d", a); e != a {
				t.Errorf("read: expected %s, got %s", e, a)
			}
		})
	}

	// check non-int types too
	if err := SwalkerWrite("String", s, "abc"); err != nil {
		t.Errorf("unexpected write error: %v", err)
	}
	if e, a := "abc", s.String; e != a {
		t.Errorf("String: expected %q, got %q", e, a)
	}
	if err := SwalkerWrite("Bool", s, "true"); err != nil {
		t.Errorf("unexpected write error: %v", err)
	}
	if e, a := true, s.Bool; e != a {
		t.Errorf("Bool: expected %t, got %t", e, a)
	}
}
