/*
 * Copyright 2022 ByteDance Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package mockey

import (
	"reflect"
	"unsafe"

	"github.com/bytedance/mockey/internal/tool"
)

// GetMethod resolve a certain public method from an instance.
func GetMethod(instance interface{}, methodName string) interface{} {
	if typ := reflect.TypeOf(instance); typ != nil {
		if m, ok := getNestedMethod(reflect.ValueOf(instance), methodName); ok {
			return m.Func.Interface()
		}
		if m, ok := typ.MethodByName(methodName); ok {
			return m.Func.Interface()
		}
		if m, ok := getFieldMethod(instance, methodName); ok {
			return m
		}
	}
	tool.Assert(false, "can't reflect instance method :%v", methodName)
	return nil
}

// getFieldMethod gets a functional field's value as an instance
// The return instance is not original field but a new function object points to
// the same function.
// for example:
//
//	  type Fn func()
//	  type Foo struct {
//			privateField Fn
//	  }
//	  func NewFoo() Foo { return Foo{ privateField: func() { /*do nothing*/ } }}
//
// getFieldMethod(NewFoo(),"privateField") will return a function object which
// points to the anonymous function in NewFoo
func getFieldMethod(instance interface{}, fieldName string) (interface{}, bool) {
	v := reflect.Indirect(reflect.ValueOf(instance))
	if v.Kind() != reflect.Struct {
		return nil, false
	}

	field := v.FieldByName(fieldName)
	if !field.IsValid() || field.Kind() != reflect.Func {
		return nil, false
	}

	carrier := reflect.MakeFunc(field.Type(), nil)
	type function struct {
		_      uintptr
		fnAddr *uintptr
	}
	*(*function)(unsafe.Pointer(&carrier)).fnAddr = field.Pointer()
	return carrier.Interface(), true
}

// GetPrivateMethod ...
// Deprecated, use GetMethod instead.
func GetPrivateMethod(instance interface{}, methodName string) interface{} {
	m, ok := reflect.TypeOf(instance).MethodByName(methodName)
	if ok {
		return m.Func.Interface()
	}
	tool.Assert(false, "can't reflect instance method :%v", methodName)
	return nil
}

// GetNestedMethod resolves a certain public method in anonymous structs, it will
// look for the specific method in every anonymous struct field recursively.
// Deprecated, use GetMethod instead.
func GetNestedMethod(instance interface{}, methodName string) interface{} {
	if typ := reflect.TypeOf(instance); typ != nil {
		if m, ok := getNestedMethod(reflect.ValueOf(instance), methodName); ok {
			return m.Func.Interface()
		}
	}
	tool.Assert(false, "can't reflect instance method :%v", methodName)
	return nil
}

func getNestedMethod(val reflect.Value, methodName string) (reflect.Method, bool) {
	typ := val.Type()
	kind := typ.Kind()
	if kind == reflect.Ptr || kind == reflect.Interface {
		val = val.Elem()
	}
	if !val.IsValid() {
		return reflect.Method{}, false
	}

	typ = val.Type()
	kind = typ.Kind()
	if kind == reflect.Struct {
		for i := 0; i < typ.NumField(); i++ {
			if !typ.Field(i).Anonymous {
				// there is no need to acquire non-anonymous method
				continue
			}
			if m, ok := getNestedMethod(val.Field(i), methodName); ok {
				return m, true
			}
		}
	}
	// a struct receiver is prior to the corresponding pointer receiver
	if m, ok := typ.MethodByName(methodName); ok {
		return m, true
	}
	return reflect.PtrTo(typ).MethodByName(methodName)
}

// GetGoroutineId ...
// Deprecated
func GetGoroutineId() int64 {
	return tool.GetGoroutineID()
}
