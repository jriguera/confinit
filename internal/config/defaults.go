package config

// Source: https://github.com/creasty/defaults
// Copyright (c) 2017-present Yuki Iwanaga
//
// MIT License
//
// Permission is hereby granted, free of charge, to any person obtaining
// a copy of this software and associated documentation files (the
// "Software"), to deal in the Software without restriction, including
// without limitation the rights to use, copy, modify, merge, publish,
// distribute, sublicense, and/or sell copies of the Software, and to
// permit persons to whom the Software is furnished to do so, subject to
// the following conditions:
//
// The above copyright notice and this permission notice shall be
// included in all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
// EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
// MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
// NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE
// LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION
// OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION
// WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

import (
	"encoding/json"
	"errors"
	"reflect"
	"strconv"
	"time"

	"confinit/pkg/log"
)

const (
	configDefaultFieldTag = "default"
)

// Setter is an interface for setting default values
type Setter interface {
	SetDefaults()
}

// SetDefaultConfig initializes members in a struct referenced by a pointer.
// Maps and slices are initialized by `make` and other primitive types are set
// with default values.
func (c *Config) SetDefaultConfig() error {
	return c.SetDefault(c)
}

// SetDefault initializes members in a struct referenced by a pointer.
// Maps and slices are initialized by `make` and other primitive types are set
// with default values. `ptr` should be a struct pointer
func (c *Config) SetDefault(ptr interface{}) error {
	if reflect.TypeOf(ptr).Kind() != reflect.Ptr {
		err := errors.New("Not a struct pointer")
		log.Fatal(err.Error())
		return err
	}
	v := reflect.ValueOf(ptr).Elem()
	t := v.Type()
	if t.Kind() != reflect.Struct {
		err := errors.New("Not a struct")
		log.Fatal(err.Error())
		return err
	}
	for i := 0; i < t.NumField(); i++ {
		if defaultVal := t.Field(i).Tag.Get(configDefaultFieldTag); defaultVal != "-" {
			//fmt.Printf("Item: %v: %s\n", t.Field(i).Name, defaultVal)
			if err := c.setField(v.Field(i), defaultVal, false); err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *Config) setBaseTypeField(field reflect.Value, defaultVal string) bool {
	switch field.Kind() {
	case reflect.Bool:
		if val, err := strconv.ParseBool(defaultVal); err == nil {
			field.Set(reflect.ValueOf(val).Convert(field.Type()))
		}
		return true
	case reflect.Int:
		if val, err := strconv.ParseInt(defaultVal, 10, 64); err == nil {
			field.Set(reflect.ValueOf(int(val)).Convert(field.Type()))
		}
		return true
	case reflect.Int8:
		if val, err := strconv.ParseInt(defaultVal, 10, 8); err == nil {
			field.Set(reflect.ValueOf(int8(val)).Convert(field.Type()))
		}
		return true
	case reflect.Int16:
		if val, err := strconv.ParseInt(defaultVal, 10, 16); err == nil {
			field.Set(reflect.ValueOf(int16(val)).Convert(field.Type()))
		}
		return true
	case reflect.Int32:
		if val, err := strconv.ParseInt(defaultVal, 10, 32); err == nil {
			field.Set(reflect.ValueOf(int32(val)).Convert(field.Type()))
		}
		return true
	case reflect.Int64:
		if val, err := time.ParseDuration(defaultVal); err == nil {
			field.Set(reflect.ValueOf(val).Convert(field.Type()))
		} else if val, err := strconv.ParseInt(defaultVal, 10, 64); err == nil {
			field.Set(reflect.ValueOf(val).Convert(field.Type()))
		}
		return true
	case reflect.Uint:
		if val, err := strconv.ParseUint(defaultVal, 10, 64); err == nil {
			field.Set(reflect.ValueOf(uint(val)).Convert(field.Type()))
		}
		return true
	case reflect.Uint8:
		if val, err := strconv.ParseUint(defaultVal, 10, 8); err == nil {
			field.Set(reflect.ValueOf(uint8(val)).Convert(field.Type()))
		}
		return true
	case reflect.Uint16:
		if val, err := strconv.ParseUint(defaultVal, 10, 16); err == nil {
			field.Set(reflect.ValueOf(uint16(val)).Convert(field.Type()))
		}
		return true
	case reflect.Uint32:
		if val, err := strconv.ParseUint(defaultVal, 10, 32); err == nil {
			field.Set(reflect.ValueOf(uint32(val)).Convert(field.Type()))
		}
		return true
	case reflect.Uint64:
		if val, err := strconv.ParseUint(defaultVal, 10, 64); err == nil {
			field.Set(reflect.ValueOf(val).Convert(field.Type()))
		}
		return true
	case reflect.Uintptr:
		if val, err := strconv.ParseUint(defaultVal, 10, 64); err == nil {
			field.Set(reflect.ValueOf(uintptr(val)).Convert(field.Type()))
		}
		return true
	case reflect.Float32:
		if val, err := strconv.ParseFloat(defaultVal, 32); err == nil {
			field.Set(reflect.ValueOf(float32(val)).Convert(field.Type()))
		}
		return true
	case reflect.Float64:
		if val, err := strconv.ParseFloat(defaultVal, 64); err == nil {
			field.Set(reflect.ValueOf(val).Convert(field.Type()))
		}
		return true
	case reflect.String:
		field.Set(reflect.ValueOf(defaultVal).Convert(field.Type()))
		return true
	}
	return false
}

func (c *Config) setField(field reflect.Value, defaultVal string, ptr bool) error {
	if !field.CanSet() {
		return nil
	}
	switch field.Kind() {
	case reflect.Slice, reflect.Map:
		if field.Len() == 0 && defaultVal == "" {
			return nil
		}
	}
	if c.isInitialValue(field) {
		if !ptr {
			c.setBaseTypeField(field, defaultVal)
		}
		switch field.Kind() {
		case reflect.Slice:
			ref := reflect.New(field.Type())
			ref.Elem().Set(reflect.MakeSlice(field.Type(), 0, 0))
			if defaultVal != "" && defaultVal != "[]" {
				if err := json.Unmarshal([]byte(defaultVal), ref.Interface()); err != nil {
					return err
				}
			}
			field.Set(ref.Elem().Convert(field.Type()))
		case reflect.Map:
			ref := reflect.New(field.Type())
			ref.Elem().Set(reflect.MakeMap(field.Type()))
			if defaultVal != "" && defaultVal != "{}" {
				if err := json.Unmarshal([]byte(defaultVal), ref.Interface()); err != nil {
					return err
				}
			}
			field.Set(ref.Elem().Convert(field.Type()))
		case reflect.Struct:
			ref := reflect.New(field.Type())
			if defaultVal != "" && defaultVal != "{}" {
				if err := json.Unmarshal([]byte(defaultVal), ref.Interface()); err != nil {
					return err
				}
			}
			field.Set(ref.Elem())
		case reflect.Ptr:
			field.Set(reflect.New(field.Type().Elem()))
			if done := c.setBaseTypeField(field.Elem(), defaultVal); done {
				return nil
			}
		}
	}
	switch field.Kind() {
	case reflect.Ptr:
		c.setField(field.Elem(), defaultVal, true)
		c.callSetter(field.Interface())
	case reflect.Struct:
		ref := reflect.New(field.Type())
		ref.Elem().Set(field)
		if err := c.SetDefault(ref.Interface()); err != nil {
			return err
		}
		c.callSetter(ref.Interface())
		field.Set(ref.Elem())
	case reflect.Slice:
		for j := 0; j < field.Len(); j++ {
			if err := c.setField(field.Index(j), defaultVal, false); err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *Config) isInitialValue(field reflect.Value) bool {
	return reflect.DeepEqual(reflect.Zero(field.Type()).Interface(), field.Interface())
}

func (c *Config) callSetter(v interface{}) {
	if ds, ok := v.(Setter); ok {
		ds.SetDefaults()
	}
}
