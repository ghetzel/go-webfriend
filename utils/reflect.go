package utils

import (
	"fmt"
	"reflect"

	"github.com/ghetzel/go-stockutil/maputil"
	"github.com/ghetzel/go-stockutil/typeutil"
)

var errorInterface = reflect.TypeOf((*error)(nil)).Elem()

func GetFunctionByName(from interface{}, name string) (reflect.Value, error) {
	var fromV reflect.Value

	if fV, ok := from.(reflect.Value); ok {
		fromV = fV
	} else {
		fromV = reflect.ValueOf(from)
	}

	if methodV := fromV.MethodByName(name); methodV.IsValid() && methodV.Kind() == reflect.Func {
		return methodV, nil
	} else {
		return reflect.Value{}, fmt.Errorf("could not locate method %v in %T (%v)", name, from, fromV)
	}
}

func CallCommandFunction(from interface{}, name string, first interface{}, rest map[string]interface{}) (interface{}, error) {
	if fn, err := GetFunctionByName(from, name); err == nil {
		inputs := []interface{}{first, rest}
		arguments := make([]reflect.Value, fn.Type().NumIn())

		for i := 0; i < len(arguments); i++ {
			if i < len(inputs) {
				argT := fn.Type().In(i)

				if inV := reflect.ValueOf(inputs[i]); inV.IsValid() {
					if inV.Type().AssignableTo(argT) {
						// attempt direct assignment
						arguments[i] = inV
						continue
					} else if inV.Type().ConvertibleTo(argT) {
						// attempt type conversion
						arguments[i] = inV.Convert(argT)
						continue
					}

					// dereference pointers
					if argT.Kind() == reflect.Ptr {
						argT = argT.Elem()
					}

					// instantiate new arg type
					arguments[i] = reflect.New(argT)

					// map arguments are used to populate newly instantiated structs
					if typeutil.IsMap(inputs[i]) {
						if argT.Kind() == reflect.Struct {
							inputM := maputil.DeepCopy(inputs[i])

							if err := maputil.TaggedStructFromMap(inputM, arguments[i], `json`); err != nil {
								return nil, fmt.Errorf("Cannot populate %v: %v", arguments[i].Type(), err)
							}
						} else {
							return nil, fmt.Errorf("Map arguments can only be used to populate structs")
						}
					}
				}
			}
		}

		returns := fn.Call(arguments)

		switch len(returns) {
		case 2:
			if lastT := returns[1].Type(); lastT.Implements(errorInterface) {
				value := returns[0].Interface()

				if v2 := returns[1].Interface(); v2 == nil {
					err = nil
				} else {
					err = v2.(error)
				}

				return value, err
			} else {
				return nil, fmt.Errorf("last return value must be an error, got %v", lastT)
			}

		case 1:
			if lastT := returns[0].Type(); lastT.Implements(errorInterface) {
				if v1 := returns[0].Interface(); v1 == nil {
					return nil, nil
				} else {
					return nil, v1.(error)
				}
			} else {
				return nil, fmt.Errorf("functions returning a single value must return an error, got %v", lastT)
			}
		}

		return nil, nil
	} else {
		return nil, err
	}
}
