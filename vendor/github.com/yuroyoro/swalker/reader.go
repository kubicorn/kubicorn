package swalker

import (
	"fmt"
	"reflect"
)

// Read : read value from given struct or map by expressions
func Read(exp string, value interface{}) (interface{}, error) {
	exps, err := Parse(exp)
	if err != nil {
		return nil, err
	}

	ret, err := exps.Read(value)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

// ReadString : read string value from given struct or map by expressions
func ReadString(exp string, value interface{}) (string, error) {
	exps, err := Parse(exp)
	if err != nil {
		return "", err
	}

	return exps.ReadString(value)
}

// ReadInt : read int value from given struct or map by expressions
func ReadInt(exp string, value interface{}) (int, error) {
	exps, err := Parse(exp)
	if err != nil {
		return 0, err
	}

	return exps.ReadInt(value)
}

// ReadFloat : read float value from given struct or map by expressions
func ReadFloat(exp string, value interface{}) (float64, error) {
	exps, err := Parse(exp)
	if err != nil {
		return 0, err
	}

	return exps.ReadFloat(value)
}

// ReadBool : read bool value from given struct or map by expressions
func ReadBool(exp string, value interface{}) (bool, error) {
	exps, err := Parse(exp)
	if err != nil {
		return false, err
	}

	return exps.ReadBool(value)
}

// ReadSlice : read slice value as []interface{} from given struct or map by expressions
func ReadSlice(exp string, value interface{}) ([]interface{}, error) {
	exps, err := Parse(exp)
	if err != nil {
		return nil, err
	}

	return exps.ReadSlice(value)
}

// Read : evaluate expressions then return value.
// if value could not be found, return nil with error
func (exps Expressions) Read(val interface{}) (interface{}, error) {

	v := reflect.ValueOf(val)
	ret := &v
	var err error

	for i, exp := range exps {
		ret, err = exp.read(ret)
		if err != nil {
			return nil, err
		}
		if isNil(ret) && i != len(exps) {
			return nil, fmt.Errorf("field %s is nil", exp.Name)
		}
	}

	return ret.Interface(), nil
}

// ReadSlice : evaluate expressions then return slice value as []interface{}.
func (exps Expressions) ReadSlice(val interface{}) ([]interface{}, error) {
	v, err := exps.Read(val)
	if err != nil {
		return nil, err
	}

	ret, ok := v.([]interface{})
	if !ok {
		return nil, fmt.Errorf("extracted value is not []interface : %T", v)
	}
	return ret, nil
}

// ReadString : evaluate expressions then return string value
func (exps Expressions) ReadString(val interface{}) (string, error) {
	v, err := exps.Read(val)
	if err != nil {
		return "", err
	}

	ret, ok := v.(string)
	if !ok {
		return "", fmt.Errorf("extracted value is not string : %T", v)
	}
	return ret, nil
}

// ReadInt : evaluate expressions then return int value
func (exps Expressions) ReadInt(val interface{}) (int, error) {
	v, err := exps.Read(val)
	if err != nil {
		return 0, err
	}

	ret, ok := v.(int)
	if !ok {
		return 0, fmt.Errorf("extracted value is not int: %T", v)
	}
	return ret, nil
}

// ReadFloat : evaluate expressions then return int value
func (exps Expressions) ReadFloat(val interface{}) (float64, error) {
	v, err := exps.Read(val)
	if err != nil {
		return 0, err
	}

	ret, ok := v.(float64)
	if !ok {
		return 0, fmt.Errorf("extracted value is not float64: %T", v)
	}
	return ret, nil
}

// ReadBool : evaluate expressions then return bool value
func (exps Expressions) ReadBool(val interface{}) (bool, error) {
	v, err := exps.Read(val)
	if err != nil {
		return false, err
	}

	ret, ok := v.(bool)
	if !ok {
		return false, fmt.Errorf("extracted value is not bool: %T", v)
	}
	return ret, nil
}

func (exp *Expression) read(v *reflect.Value) (*reflect.Value, error) {
	switch exp.Type {
	case Property:
		return exp.readProperty(v)
	case Indexing:
		return exp.readIndexing(v)
	}

	return nil, fmt.Errorf("unknow expression type %v", exp.Type)
}

func (exp *Expression) readIndexing(v *reflect.Value) (*reflect.Value, error) {
	arr, err := exp.readProperty(v)
	if err != nil {
		return nil, err
	}

	switch arr.Kind() {
	case reflect.Array, reflect.Slice:
		if arr.Len() < exp.Index {
			return nil, fmt.Errorf("field %s len %d : index %d is out of range", exp.Name, arr.Len(), exp.Index)
		}
		res := arr.Index(exp.Index)
		return unwrap(&res), nil
	}

	return nil, fmt.Errorf("field %s is not array or slice : %s", exp.Name, arr.Kind())
}

func (exp *Expression) readProperty(v *reflect.Value) (*reflect.Value, error) {
	name := exp.Name
	v = indirecte(v)
	typ := v.Type()
	switch v.Kind() {
	case reflect.Struct:
		tf, ok := v.Type().FieldByName(exp.Name)
		if ok {
			field := v.FieldByIndex(tf.Index)
			if tf.PkgPath != "" { // field is unexported
				return nil, fmt.Errorf("%s is an unexported field of struct type %s", name, typ)
			}
			return unwrap(&field), nil
		}
		return nil, fmt.Errorf("%s is not a field of struct type %s", name, typ)
	case reflect.Map:
		// If it's a map, attempt to use the field name as a key.
		nameVal := reflect.ValueOf(name)
		if nameVal.Type().AssignableTo(v.Type().Key()) {
			result := v.MapIndex(nameVal)
			if !result.IsValid() {
				return nil, fmt.Errorf("map has no entry for key %q", name)
			}
			return unwrap(&result), nil
		}
	}
	return nil, fmt.Errorf("can't evaluate field %s in type %s (%s)", name, typ, v.Kind())
}
