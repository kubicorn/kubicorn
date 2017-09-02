package swalker

import (
	"fmt"
	"reflect"
)

// Write : write value by expression
func Write(exp string, obj, value interface{}) error {
	exps, err := Parse(exp)
	if err != nil {
		return err
	}

	return exps.Write(obj, value)
}

func (exps Expressions) Write(obj, val interface{}) error {
	v := reflect.ValueOf(obj)
	target := &v
	var err error

	for i, exp := range exps {
		if i == len(exps)-1 {
			return exp.write(target, val)
		}
		target, err = exp.read(target)
		if err != nil {
			return err
		}
		if isNil(target) && i != len(exps) {
			return fmt.Errorf("field %s is nil", exp.Name)
		}
	}

	panic("not reached")

}

func (exp *Expression) write(obj *reflect.Value, val interface{}) error {
	switch exp.Type {
	case Property:
		return exp.writeProperty(obj, val)
	case Indexing:
		return exp.writeIndexing(obj, val)
	}

	return fmt.Errorf("unknow expression type %v", exp.Type)
}

func (exp *Expression) writeIndexing(obj *reflect.Value, val interface{}) error {
	arr, err := exp.readProperty(obj)
	if err != nil {
		return err
	}

	switch arr.Kind() {
	case reflect.Array, reflect.Slice:
		if arr.Len() < exp.Index {
			return fmt.Errorf("field %s len %d : index %d is out of range", exp.Name, arr.Len(), exp.Index)
		}
		dst := arr.Index(exp.Index)

		return exp.set(&dst, val)
	}

	return fmt.Errorf("field %s is not array or slice : %s", exp.Name, arr.Kind())
}

func (exp *Expression) writeProperty(obj *reflect.Value, val interface{}) error {
	name := exp.Name
	v := indirecte(obj)
	typ := v.Type()
	switch v.Kind() {
	case reflect.Struct:
		tf, ok := v.Type().FieldByName(exp.Name)
		if ok {
			field := v.FieldByIndex(tf.Index)
			if tf.PkgPath != "" { // field is unexported
				return fmt.Errorf("%s is an unexported field of struct type %s", name, typ)
			}

			return exp.set(&field, val)
		}
		return fmt.Errorf("%s is not a field of struct type %s", name, typ)
	case reflect.Map:
		// If it's a map, attempt to use the field name as a key.
		nameVal := reflect.ValueOf(name)
		if nameVal.Type().AssignableTo(v.Type().Key()) {
			vv := reflect.ValueOf(val)
			v.SetMapIndex(nameVal, vv)
			return nil
		}
	}
	return fmt.Errorf("can't evaluate field %s in type %s (%s)", name, typ, v.Kind())
}

func (exp *Expression) set(dst *reflect.Value, val interface{}) error {

	if dst.CanSet() == false {
		return fmt.Errorf("cannot write ot field %s[%d]", exp.Name, exp.Index)
	}

	v := reflect.ValueOf(val)

	if dst.Kind() == reflect.Interface {
		v = *copy(dst, &v)
	}

	if dst.Type().AssignableTo(v.Type()) == false {
		switch exp.Type {
		case Indexing:
			return fmt.Errorf("cannot write %s to field %s[%d](%s) : %v", v.Type(), exp.Name, exp.Index, dst.Type(), val)
		case Property:
			return fmt.Errorf("cannot write %s to field %s(%s) : %v", v.Type(), exp.Name, dst.Type(), val)
		}
	}
	dst.Set(v)
	return nil
}

func copy(dst, v *reflect.Value) *reflect.Value {
	if v.CanAddr() == false {
		copy := reflect.New(dst.Type()).Elem()
		copy.Set(*v)
		return &copy
	}

	return v
}
