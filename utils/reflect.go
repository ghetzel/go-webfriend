package utils

import (
	"fmt"
	"reflect"

	"github.com/ghetzel/go-stockutil/maputil"
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
		arguments := make([]reflect.Value, fn.Type().NumIn())
		firstV := reflect.ValueOf(first)

		for i := 0; i < len(arguments); i++ {
			argT := fn.Type().In(i)

			if i == 0 && firstV.IsValid() {
				if firstV.Type().AssignableTo(argT) {
					arguments[i] = firstV
				} else if firstV.Type().ConvertibleTo(argT) {
					arguments[i] = firstV.Convert(argT)
				} else {
					return nil, fmt.Errorf("first argument expects %v, got %T", argT, first)
				}
				// } else if argT.Kind() == reflect.Ptr && argT.Elem().Kind() == reflect.Struct {
			} else {
				if argT.Kind() == reflect.Ptr {
					argT = argT.Elem()
				}

				arguments[i] = reflect.New(argT)

				if len(rest) > 0 {
					if err := maputil.TaggedStructFromMap(rest, arguments[i], `json`); err != nil {
						return nil, fmt.Errorf("Cannot populate %v: %v", arguments[i].Type(), err)
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
