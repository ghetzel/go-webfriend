// Commands for manipulating the current Friendscript variable scope.
package vars

import (
	"fmt"

	"github.com/ghetzel/go-stockutil/sliceutil"
	"github.com/ghetzel/go-stockutil/stringutil"
	"github.com/ghetzel/go-stockutil/typeutil"
	"github.com/ghetzel/go-webfriend/browser"
	"github.com/ghetzel/go-webfriend/utils"
	defaults "github.com/mcuadros/go-defaults"
)

type Commands struct {
	browser   *browser.Browser
	scopeable utils.Scopeable
}

func New(browser *browser.Browser, scopeable utils.Scopeable) *Commands {
	return &Commands{
		browser:   browser,
		scopeable: scopeable,
	}
}

func (self *Commands) ExecuteCommand(name string, arg interface{}, objargs map[string]interface{}) (interface{}, error) {
	return utils.CallCommandFunction(self, stringutil.Camelize(name), arg, objargs)
}

func (self *Commands) Clear(key string) {
	self.scopeable.Scope().Set(key, nil)
}

type GetArgs struct {
	Fallback interface{} `json:"fallback"`
}

// Return the value of a specific variable defined in a scope.
func (self *Commands) Get(key string, args *GetArgs) (interface{}, error) {
	if args == nil {
		args = &GetArgs{}
	}

	defaults.SetDefaults(args)

	return self.scopeable.Scope().Get(key, args.Fallback), nil
}

type InterpolateArgs struct {
	Values interface{} `json:"values"`
}

// Return a value interpolated with values from a scope or ones that are explicitly provided.
func (self *Commands) Interpolate(format string, args *InterpolateArgs) (string, error) {
	return self.scopeable.Scope().Interpolate(format), nil
}

type SetArgs struct {
	Value       interface{} `json:"value"`
	Interpolate bool        `json:"interpolate" default:"true"`
}

// Set the named variable to the given value, optionally interpolating variables from the current
// scope into the variable.
func (self *Commands) Set(key string, args *SetArgs) (interface{}, error) {
	if args == nil {
		args = &SetArgs{}
	}

	defaults.SetDefaults(args)

	if args.Interpolate {
		if v, err := self.Interpolate(typeutil.V(args.Value).String(), nil); err == nil {
			args.Value = v
		} else {
			return ``, err
		}
	}

	self.scopeable.Scope().Set(key, args.Value)
	return args.Value, nil
}

type EnsureArgs struct {
	Message string `json:"message"`
}

// Emit an error if the given key does not exist, optionally with a user-specified message.
func (self *Commands) Ensure(key string, args *EnsureArgs) error {
	if args == nil {
		args = &EnsureArgs{}
	}

	defaults.SetDefaults(args)

	if self.scopeable.Scope().Get(key) != nil {
		return nil
	} else {
		if args.Message != `` {
			return fmt.Errorf(args.Message, key)
		} else {
			return fmt.Errorf("Variable '%s' must be specified", key)
		}
	}
}

type PushArgs struct {
	Value interface{} `json:"value"`
}

// Push the given value onto the array at the specified key, creating the array if not
// present, and converting the existing value into an array already set to non-array value.
func (self *Commands) Push(key string, args *PushArgs) error {
	if args == nil {
		args = &PushArgs{}
	}

	defaults.SetDefaults(args)

	var newValue interface{}

	if existing := self.scopeable.Scope().Get(key); existing != nil {
		newValue = append(sliceutil.Sliceify(existing), args.Value)
	} else {
		newValue = sliceutil.Sliceify(args.Value)
	}

	self.scopeable.Scope().Set(key, newValue)
	return nil
}

// Take the last value from the array at key.  If key is an array, the last value of
// that array will be returned and the remainder will be left at key.  Empty arrays will
// return nil and be unset.  Non-array values will be returned and the key will be unset.
func (self *Commands) Pop(key string) (interface{}, error) {
	var value interface{}

	if existing := self.scopeable.Scope().Get(key); existing != nil {
		values := sliceutil.Sliceify(existing)

		switch len(values) {
		case 0:
			// clear key, return nil
			self.scopeable.Scope().Set(key, nil)
		case 1:
			// clear key, return only value
			value = values[0]
			self.scopeable.Scope().Set(key, nil)
		default:
			// set existing array to all but last item, return last item
			value = values[0]
			self.scopeable.Scope().Set(key, values[1:])
		}
	}

	return value, nil
}
