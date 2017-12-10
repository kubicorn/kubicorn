// Copyright 2014 Dario Castañé. All rights reserved.
// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Based on src/pkg/reflect/deepequal.go from official
// golang's stdlib.

package mergo

import (
	"fmt"
	"reflect"
	"unicode"
	"unicode/utf8"
)

func changeInitialCase(s string, mapper func(rune) rune) string {
	if s == "" {
		return s
	}
	r, n := utf8.DecodeRuneInString(s)
	return string(mapper(r)) + s[n:]
}

func isExported(field reflect.StructField) bool {
	r, _ := utf8.DecodeRuneInString(field.Name)
	return r >= 'A' && r <= 'Z'
}

// Traverses recursively both values, assigning src's fields values to dst.
// The map argument tracks comparisons that have already been seen, which allows
// short circuiting on recursive types.
<<<<<<< HEAD
<<<<<<< HEAD
func deepMap(dst, src reflect.Value, visited map[uintptr]*visit, depth int, config *config) (err error) {
	overwrite := config.overwrite
=======
func deepMap(dst, src reflect.Value, visited map[uintptr]*visit, depth int) (err error) {
>>>>>>> Working on getting compiling
=======
func deepMap(dst, src reflect.Value, visited map[uintptr]*visit, depth int, overwrite bool) (err error) {
>>>>>>> moar deps
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
	zeroValue := reflect.Value{}
	switch dst.Kind() {
	case reflect.Map:
		dstMap := dst.Interface().(map[string]interface{})
		for i, n := 0, src.NumField(); i < n; i++ {
			srcType := src.Type()
			field := srcType.Field(i)
			if !isExported(field) {
				continue
			}
			fieldName := field.Name
			fieldName = changeInitialCase(fieldName, unicode.ToLower)
<<<<<<< HEAD
<<<<<<< HEAD
			if v, ok := dstMap[fieldName]; !ok || (isEmptyValue(reflect.ValueOf(v)) || overwrite) {
				dstMap[fieldName] = src.Field(i).Interface()
			}
		}
	case reflect.Ptr:
		if dst.IsNil() {
			v := reflect.New(dst.Type().Elem())
			dst.Set(v)
		}
		dst = dst.Elem()
		fallthrough
=======
			if v, ok := dstMap[fieldName]; !ok || isEmptyValue(reflect.ValueOf(v)) {
				dstMap[fieldName] = src.Field(i).Interface()
			}
		}
>>>>>>> Working on getting compiling
=======
			if v, ok := dstMap[fieldName]; !ok || (isEmptyValue(reflect.ValueOf(v)) || overwrite) {
				dstMap[fieldName] = src.Field(i).Interface()
			}
		}
	case reflect.Ptr:
		if dst.IsNil() {
			v := reflect.New(dst.Type().Elem())
			dst.Set(v)
		}
		dst = dst.Elem()
		fallthrough
>>>>>>> moar deps
	case reflect.Struct:
		srcMap := src.Interface().(map[string]interface{})
		for key := range srcMap {
			srcValue := srcMap[key]
			fieldName := changeInitialCase(key, unicode.ToUpper)
			dstElement := dst.FieldByName(fieldName)
			if dstElement == zeroValue {
				// We discard it because the field doesn't exist.
				continue
			}
			srcElement := reflect.ValueOf(srcValue)
			dstKind := dstElement.Kind()
			srcKind := srcElement.Kind()
			if srcKind == reflect.Ptr && dstKind != reflect.Ptr {
				srcElement = srcElement.Elem()
				srcKind = reflect.TypeOf(srcElement.Interface()).Kind()
			} else if dstKind == reflect.Ptr {
				// Can this work? I guess it can't.
				if srcKind != reflect.Ptr && srcElement.CanAddr() {
					srcPtr := srcElement.Addr()
					srcElement = reflect.ValueOf(srcPtr)
					srcKind = reflect.Ptr
				}
			}
<<<<<<< HEAD
<<<<<<< HEAD

=======
>>>>>>> Working on getting compiling
=======

>>>>>>> moar deps
			if !srcElement.IsValid() {
				continue
			}
			if srcKind == dstKind {
<<<<<<< HEAD
<<<<<<< HEAD
				if err = deepMerge(dstElement, srcElement, visited, depth+1, config); err != nil {
					return
				}
			} else if dstKind == reflect.Interface && dstElement.Kind() == reflect.Interface {
				if err = deepMerge(dstElement, srcElement, visited, depth+1, config); err != nil {
					return
				}
			} else if srcKind == reflect.Map {
				if err = deepMap(dstElement, srcElement, visited, depth+1, config); err != nil {
					return
				}
			} else {
				return fmt.Errorf("type mismatch on %s field: found %v, expected %v", fieldName, srcKind, dstKind)
=======
				if err = deepMerge(dstElement, srcElement, visited, depth+1); err != nil {
=======
				if err = deepMerge(dstElement, srcElement, visited, depth+1, overwrite); err != nil {
>>>>>>> moar deps
					return
				}
			} else if dstKind == reflect.Interface && dstElement.Kind() == reflect.Interface {
				if err = deepMerge(dstElement, srcElement, visited, depth+1, overwrite); err != nil {
					return
				}
			} else if srcKind == reflect.Map {
				if err = deepMap(dstElement, srcElement, visited, depth+1, overwrite); err != nil {
					return
				}
<<<<<<< HEAD
>>>>>>> Working on getting compiling
=======
			} else {
				return fmt.Errorf("type mismatch on %s field: found %v, expected %v", fieldName, srcKind, dstKind)
>>>>>>> moar deps
			}
		}
	}
	return
}

// Map sets fields' values in dst from src.
// src can be a map with string keys or a struct. dst must be the opposite:
// if src is a map, dst must be a valid pointer to struct. If src is a struct,
// dst must be map[string]interface{}.
// It won't merge unexported (private) fields and will do recursively
// any exported field.
// If dst is a map, keys will be src fields' names in lower camel case.
// Missing key in src that doesn't match a field in dst will be skipped. This
// doesn't apply if dst is a map.
// This is separated method from Merge because it is cleaner and it keeps sane
// semantics: merging equal types, mapping different (restricted) types.
<<<<<<< HEAD
func Map(dst, src interface{}, opts ...func(*config)) error {
	return _map(dst, src, opts...)
}

// MapWithOverwrite will do the same as Map except that non-empty dst attributes will be overriden by
// non-empty src attribute values.
// Deprecated: Use Map(…) with WithOverride
func MapWithOverwrite(dst, src interface{}, opts ...func(*config)) error {
	return _map(dst, src, append(opts, WithOverride)...)
}

func _map(dst, src interface{}, opts ...func(*config)) error {
=======
func Map(dst, src interface{}) error {
<<<<<<< HEAD
>>>>>>> Working on getting compiling
=======
	return _map(dst, src, false)
}

// MapWithOverwrite will do the same as Map except that non-empty dst attributes will be overriden by
// non-empty src attribute values.
func MapWithOverwrite(dst, src interface{}) error {
	return _map(dst, src, true)
}

func _map(dst, src interface{}, overwrite bool) error {
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
	// To be friction-less, we redirect equal-type arguments
	// to deepMerge. Only because arguments can be anything.
	if vSrc.Kind() == vDst.Kind() {
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
	switch vSrc.Kind() {
	case reflect.Struct:
		if vDst.Kind() != reflect.Map {
			return ErrExpectedMapAsDestination
		}
	case reflect.Map:
		if vDst.Kind() != reflect.Struct {
			return ErrExpectedStructAsDestination
		}
	default:
		return ErrNotSupported
	}
<<<<<<< HEAD
<<<<<<< HEAD
	return deepMap(vDst, vSrc, make(map[uintptr]*visit), 0, config)
=======
	return deepMap(vDst, vSrc, make(map[uintptr]*visit), 0)
>>>>>>> Working on getting compiling
=======
	return deepMap(vDst, vSrc, make(map[uintptr]*visit), 0, overwrite)
>>>>>>> moar deps
}
