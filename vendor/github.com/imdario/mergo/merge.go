// Copyright 2013 Dario Castañé. All rights reserved.
// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Based on src/pkg/reflect/deepequal.go from official
// golang's stdlib.

package mergo

<<<<<<< HEAD
import "reflect"

func hasExportedField(dst reflect.Value) (exported bool) {
	for i, n := 0, dst.NumField(); i < n; i++ {
		field := dst.Type().Field(i)
		if field.Anonymous && dst.Field(i).Kind() == reflect.Struct {
			exported = exported || hasExportedField(dst.Field(i))
		} else {
			exported = exported || len(field.PkgPath) == 0
		}
	}
	return
}

type config struct {
	overwrite    bool
	transformers transformers
}

type transformers interface {
	Transformer(reflect.Type) func(dst, src reflect.Value) error
}
=======
import (
	"reflect"
)
>>>>>>> Working on getting compiling

func hasExportedField(dst reflect.Value) (exported bool) {
	for i, n := 0, dst.NumField(); i < n; i++ {
		field := dst.Type().Field(i)
		if field.Anonymous {
			exported = exported || hasExportedField(dst.Field(i))
		} else {
			exported = exported || len(field.PkgPath) == 0
		}
	}
	return
}

// Traverses recursively both values, assigning src's fields values to dst.
// The map argument tracks comparisons that have already been seen, which allows
// short circuiting on recursive types.
<<<<<<< HEAD
<<<<<<< HEAD
func deepMerge(dst, src reflect.Value, visited map[uintptr]*visit, depth int, config *config) (err error) {
	overwrite := config.overwrite

=======
func deepMerge(dst, src reflect.Value, visited map[uintptr]*visit, depth int) (err error) {
>>>>>>> Working on getting compiling
=======
func deepMerge(dst, src reflect.Value, visited map[uintptr]*visit, depth int, overwrite bool) (err error) {
>>>>>>> moar deps
	if !src.IsValid() {
		return
	}
	if dst.CanAddr() {
		addr := dst.UnsafeAddr()
		h := 17 * addr
		seen := visited[h]
		typ := dst.Type()
		for p := seen; p != nil; p = p.next {
			if p.ptr == addr && p.typ == typ {
				return nil
			}
		}
		// Remember, remember...
		visited[h] = &visit{addr, typ, seen}
	}
<<<<<<< HEAD

	if config.transformers != nil {
		if fn := config.transformers.Transformer(dst.Type()); fn != nil {
			err = fn(dst, src)
			return
		}
	}

	switch dst.Kind() {
	case reflect.Struct:
		if hasExportedField(dst) {
			for i, n := 0, dst.NumField(); i < n; i++ {
				if err = deepMerge(dst.Field(i), src.Field(i), visited, depth+1, config); err != nil {
					return
				}
			}
		} else {
			if dst.CanSet() && !isEmptyValue(src) && (overwrite || isEmptyValue(dst)) {
				dst.Set(src)
			}
		}
	case reflect.Map:
		if len(src.MapKeys()) == 0 && !src.IsNil() && len(dst.MapKeys()) == 0 {
			dst.Set(reflect.MakeMap(dst.Type()))
			return
		}
=======
	switch dst.Kind() {
	case reflect.Struct:
		if hasExportedField(dst) {
			for i, n := 0, dst.NumField(); i < n; i++ {
				if err = deepMerge(dst.Field(i), src.Field(i), visited, depth+1, overwrite); err != nil {
					return
				}
			}
		} else {
			if dst.CanSet() && !isEmptyValue(src) && (overwrite || isEmptyValue(dst)) {
				dst.Set(src)
			}
		}
	case reflect.Map:
<<<<<<< HEAD
>>>>>>> Working on getting compiling
=======
		if len(src.MapKeys()) == 0 && !src.IsNil() && len(dst.MapKeys()) == 0 {
			dst.Set(reflect.MakeMap(dst.Type()))
			return
		}
>>>>>>> moar deps
		for _, key := range src.MapKeys() {
			srcElement := src.MapIndex(key)
			if !srcElement.IsValid() {
				continue
			}
			dstElement := dst.MapIndex(key)
<<<<<<< HEAD
<<<<<<< HEAD
=======
>>>>>>> moar deps
			switch srcElement.Kind() {
			case reflect.Chan, reflect.Func, reflect.Map, reflect.Ptr, reflect.Interface, reflect.Slice:
				if srcElement.IsNil() {
					continue
				}
<<<<<<< HEAD
				fallthrough
			default:
				if !srcElement.CanInterface() {
					continue
				}
				switch reflect.TypeOf(srcElement.Interface()).Kind() {
				case reflect.Struct:
					fallthrough
				case reflect.Ptr:
					fallthrough
				case reflect.Map:
					if err = deepMerge(dstElement, srcElement, visited, depth+1, config); err != nil {
						return
					}
				case reflect.Slice:
					srcSlice := reflect.ValueOf(srcElement.Interface())

					var dstSlice reflect.Value
					if !dstElement.IsValid() || dstElement.IsNil() {
						dstSlice = reflect.MakeSlice(srcSlice.Type(), 0, srcSlice.Len())
					} else {
						dstSlice = reflect.ValueOf(dstElement.Interface())
					}

					dstSlice = reflect.AppendSlice(dstSlice, srcSlice)
					dst.SetMapIndex(key, dstSlice)
				}
			}
			if dstElement.IsValid() && reflect.TypeOf(srcElement.Interface()).Kind() == reflect.Map {
				continue
			}

			if !isEmptyValue(srcElement) && (overwrite || (!dstElement.IsValid() || isEmptyValue(dst))) {
				if dst.IsNil() {
					dst.Set(reflect.MakeMap(dst.Type()))
				}
				dst.SetMapIndex(key, srcElement)
			}
		}
	case reflect.Slice:
		dst.Set(reflect.AppendSlice(dst, src))
=======
			switch reflect.TypeOf(srcElement.Interface()).Kind() {
			case reflect.Struct:
=======
>>>>>>> moar deps
				fallthrough
			default:
				if !srcElement.CanInterface() {
					continue
				}
				switch reflect.TypeOf(srcElement.Interface()).Kind() {
				case reflect.Struct:
					fallthrough
				case reflect.Ptr:
					fallthrough
				case reflect.Map:
					if err = deepMerge(dstElement, srcElement, visited, depth+1, overwrite); err != nil {
						return
					}
				}
			}
			if dstElement.IsValid() && reflect.TypeOf(srcElement.Interface()).Kind() == reflect.Map {
				continue
			}

			if !isEmptyValue(srcElement) && (overwrite || (!dstElement.IsValid() || isEmptyValue(dst))) {
				if dst.IsNil() {
					dst.Set(reflect.MakeMap(dst.Type()))
				}
				dst.SetMapIndex(key, srcElement)
			}
		}
>>>>>>> Working on getting compiling
	case reflect.Ptr:
		fallthrough
	case reflect.Interface:
		if src.Kind() != reflect.Interface {
			if dst.IsNil() || overwrite {
				if dst.CanSet() && (overwrite || isEmptyValue(dst)) {
					dst.Set(src)
				}
			} else if src.Kind() == reflect.Ptr {
				if err = deepMerge(dst.Elem(), src.Elem(), visited, depth+1, overwrite); err != nil {
					return
				}
			} else if dst.Elem().Type() == src.Type() {
				if err = deepMerge(dst.Elem(), src, visited, depth+1, overwrite); err != nil {
					return
				}
			} else {
				return ErrDifferentArgumentsTypes
			}
			break
		}
		if src.IsNil() {
			break
<<<<<<< HEAD
<<<<<<< HEAD
		}
		if src.Kind() != reflect.Interface {
			if dst.IsNil() || overwrite {
				if dst.CanSet() && (overwrite || isEmptyValue(dst)) {
					dst.Set(src)
				}
			} else if src.Kind() == reflect.Ptr {
				if err = deepMerge(dst.Elem(), src.Elem(), visited, depth+1, config); err != nil {
					return
				}
			} else if dst.Elem().Type() == src.Type() {
				if err = deepMerge(dst.Elem(), src, visited, depth+1, config); err != nil {
					return
				}
			} else {
				return ErrDifferentArgumentsTypes
			}
			break
		}
		if dst.IsNil() || overwrite {
			if dst.CanSet() && (overwrite || isEmptyValue(dst)) {
				dst.Set(src)
			}
		} else if err = deepMerge(dst.Elem(), src.Elem(), visited, depth+1, config); err != nil {
			return
		}
	default:
		if dst.CanSet() && !isEmptyValue(src) && (overwrite || isEmptyValue(dst)) {
=======
		} else if dst.IsNil() {
			if dst.CanSet() && isEmptyValue(dst) {
=======
		} else if dst.IsNil() || overwrite {
			if dst.CanSet() && (overwrite || isEmptyValue(dst)) {
>>>>>>> moar deps
				dst.Set(src)
			}
		} else if err = deepMerge(dst.Elem(), src.Elem(), visited, depth+1, overwrite); err != nil {
			return
		}
	default:
<<<<<<< HEAD
		if dst.CanSet() && !isEmptyValue(src) {
>>>>>>> Working on getting compiling
=======
		if dst.CanSet() && !isEmptyValue(src) && (overwrite || isEmptyValue(dst)) {
>>>>>>> moar deps
			dst.Set(src)
		}
	}
	return
}

<<<<<<< HEAD
<<<<<<< HEAD
=======
>>>>>>> moar deps
// Merge will fill any empty for value type attributes on the dst struct using corresponding
// src attributes if they themselves are not empty. dst and src must be valid same-type structs
// and dst must be a pointer to struct.
// It won't merge unexported (private) fields and will do recursively any exported field.
<<<<<<< HEAD
func Merge(dst, src interface{}, opts ...func(*config)) error {
	return merge(dst, src, opts...)
}

// MergeWithOverwrite will do the same as Merge except that non-empty dst attributes will be overriden by
// non-empty src attribute values.
// Deprecated: use Merge(…) with WithOverride
func MergeWithOverwrite(dst, src interface{}, opts ...func(*config)) error {
	return merge(dst, src, append(opts, WithOverride)...)
}

// WithTransformers adds transformers to merge, allowing to customize the merging of some types.
func WithTransformers(transformers transformers) func(*config) {
	return func(config *config) {
		config.transformers = transformers
	}
}

// WithOverride will make merge override non-empty dst attributes with non-empty src attributes values.
func WithOverride(config *config) {
	config.overwrite = true
}

func merge(dst, src interface{}, opts ...func(*config)) error {
=======
// Merge sets fields' values in dst from src if they have a zero
// value of their type.
// dst and src must be valid same-type structs and dst must be
// a pointer to struct.
// It won't merge unexported (private) fields and will do recursively
// any exported field.
func Merge(dst, src interface{}) error {
>>>>>>> Working on getting compiling
=======
func Merge(dst, src interface{}) error {
	return merge(dst, src, false)
}

// MergeWithOverwrite will do the same as Merge except that non-empty dst attributes will be overriden by
// non-empty src attribute values.
func MergeWithOverwrite(dst, src interface{}) error {
	return merge(dst, src, true)
}

func merge(dst, src interface{}, overwrite bool) error {
>>>>>>> moar deps
	var (
		vDst, vSrc reflect.Value
		err        error
	)
<<<<<<< HEAD

	config := &config{}

	for _, opt := range opts {
		opt(config)
	}

=======
>>>>>>> Working on getting compiling
	if vDst, vSrc, err = resolveValues(dst, src); err != nil {
		return err
	}
	if vDst.Type() != vSrc.Type() {
		return ErrDifferentArgumentsTypes
	}
<<<<<<< HEAD
<<<<<<< HEAD
	return deepMerge(vDst, vSrc, make(map[uintptr]*visit), 0, config)
=======
	return deepMerge(vDst, vSrc, make(map[uintptr]*visit), 0)
>>>>>>> Working on getting compiling
=======
	return deepMerge(vDst, vSrc, make(map[uintptr]*visit), 0, overwrite)
>>>>>>> moar deps
}
