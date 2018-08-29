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
	"reflect"
	"strconv"

	"github.com/yuroyoro/swalker"
)

// SwalkerWrite is a convenience wrapper around swalker.Write that automatically converts value to
// an appropriate int, uint, or bool type based on the destination field's type, if appropriate.
func SwalkerWrite(exp string, obj interface{}, value string) error {
	v, err := swalker.Read(exp, obj)
	if err != nil {
		return err
	}

	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Int:
		i, err := strconv.ParseInt(value, 10, 0)
		if err != nil {
			return err
		}
		return swalker.Write(exp, obj, int(i))
	case reflect.Int8:
		i, err := strconv.ParseInt(value, 10, 8)
		if err != nil {
			return err
		}
		return swalker.Write(exp, obj, int8(i))
	case reflect.Int16:
		i, err := strconv.ParseInt(value, 10, 16)
		if err != nil {
			return err
		}
		return swalker.Write(exp, obj, int16(i))
	case reflect.Int32:
		i, err := strconv.ParseInt(value, 10, 32)
		if err != nil {
			return err
		}
		return swalker.Write(exp, obj, int32(i))
	case reflect.Int64:
		i, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return err
		}
		return swalker.Write(exp, obj, i)
	case reflect.Uint:
		i, err := strconv.ParseUint(value, 10, 0)
		if err != nil {
			return err
		}
		return swalker.Write(exp, obj, uint(i))
	case reflect.Uint8:
		i, err := strconv.ParseUint(value, 10, 8)
		if err != nil {
			return err
		}
		return swalker.Write(exp, obj, uint8(i))
	case reflect.Uint16:
		i, err := strconv.ParseUint(value, 10, 16)
		if err != nil {
			return err
		}
		return swalker.Write(exp, obj, uint16(i))
	case reflect.Uint32:
		i, err := strconv.ParseUint(value, 10, 32)
		if err != nil {
			return err
		}
		return swalker.Write(exp, obj, uint32(i))
	case reflect.Uint64:
		i, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return err
		}
		return swalker.Write(exp, obj, i)
	case reflect.Bool:
		b, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		return swalker.Write(exp, obj, b)
	}

	return swalker.Write(exp, obj, value)
}
