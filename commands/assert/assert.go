// Suite of testing-oriented commands that will trigger errors or failures if they aren't satistifed.
//
// The Assert module is aimed at making it easier to write scripts that perform test validation,
// as is frequently done during automated acceptance testing of web applications.
//
package assert

import (
	"fmt"
	"strings"

	"github.com/ghetzel/friendscript"
	"github.com/ghetzel/friendscript/utils"
	defaults "github.com/ghetzel/go-defaults"
	"github.com/ghetzel/go-stockutil/stringutil"
	"github.com/ghetzel/go-stockutil/timeutil"
	"github.com/ghetzel/go-stockutil/typeutil"
	"github.com/ghetzel/go-webfriend/browser"
)

type Commands struct {
	friendscript.Module
	browser *browser.Browser
}

func New(browser *browser.Browser, scopeable utils.Scopeable) *Commands {
	cmd := &Commands{}

	cmd.browser = browser
	cmd.Module = friendscript.CreateModule(cmd)

	return cmd
}

// Return an error if the given value is null or zero-length.
func (self *Commands) Exists(value interface{}) error {
	if typeutil.V(value).String() != `` {
		return nil
	} else {
		return fmt.Errorf("Expected non-empty value")
	}
}

// Return an error if the given value not empty.
func (self *Commands) Empty(value interface{}) error {
	if typeutil.IsEmpty(value) {
		return nil
	} else {
		return fmt.Errorf("Expected empty value")
	}
}

// Return an error if the given value is not null.
func (self *Commands) Null(value interface{}) error {
	if value == nil {
		return nil
	} else {
		return fmt.Errorf("Expected null value")
	}
}

// Return an error if the given value is null.
func (self *Commands) NotNull(value interface{}) error {
	if value != nil {
		return nil
	} else {
		return fmt.Errorf("Expected non-null value")
	}
}

// Return an error if the given value is not true.
func (self *Commands) True(value interface{}) error {
	if stringutil.IsBooleanTrue(value) {
		return nil
	} else {
		return fmt.Errorf("Expected true value")
	}
}

// Return an error if the given value is not false.
func (self *Commands) False(value interface{}) error {
	if stringutil.IsBooleanFalse(value) {
		return nil
	} else {
		return fmt.Errorf("Expected false value")
	}
}

// Return an error if the given value is not a numeric value.
func (self *Commands) IsNumeric(value interface{}) error {
	if typeutil.IsKindOfInteger(value) || typeutil.IsKindOfFloat(value) {
		return nil
	} else {
		return fmt.Errorf("Expected numeric value")
	}
}

// Return an error if the given value is not a boolean value.
func (self *Commands) IsBoolean(value interface{}) error {
	if typeutil.IsKindOfBool(value) {
		return nil
	} else {
		return fmt.Errorf("Expected boolean value")
	}
}

// Return an error if the given value is not a string.
func (self *Commands) IsString(value interface{}) error {
	if typeutil.IsKindOfString(value) {
		return nil
	} else {
		return fmt.Errorf("Expected string value")
	}
}

// Return an error if the given value is not a scalar value.
func (self *Commands) IsScalar(value interface{}) error {
	if typeutil.IsScalar(value) {
		return nil
	} else {
		return fmt.Errorf("Expected numeric value")
	}
}

// Return an error if the given value is not parsable as a time.
func (self *Commands) IsTime(value interface{}) error {
	if stringutil.IsTime(value) {
		return nil
	} else {
		return fmt.Errorf("Expected time value")
	}
}

// Return an error if the given value is not parsable as a duration.
func (self *Commands) IsDuration(value interface{}) error {
	if _, err := timeutil.ParseDuration(fmt.Sprintf("%v", value)); err == nil {
		return nil
	} else {
		return fmt.Errorf("Expected duration value")
	}
}

// Return an error if the given value is not an object.
func (self *Commands) IsObject(value interface{}) error {
	if typeutil.IsMap(value) {
		return nil
	} else {
		return fmt.Errorf("Expected object")
	}
}

// Return an error if the given value is not an array.
func (self *Commands) IsArray(value interface{}) error {
	if typeutil.IsArray(value) {
		return nil
	} else {
		return fmt.Errorf("Expected array value")
	}
}

type BinaryComparison struct {
	Value interface{} `json:"value"`
	Test  string      `json:"test"`
}

func bc(args *BinaryComparison, op string) *BinaryComparison {
	if args == nil {
		args = &BinaryComparison{}
	}

	defaults.SetDefaults(args)

	if op != `` {
		args.Test = op
	}

	return args
}

// Return an error if the given value is not equal to the other value.
func (self *Commands) Compare(have interface{}, args *BinaryComparison) error {
	if args == nil {
		args = &BinaryComparison{}
	}

	defaults.SetDefaults(args)
	want := args.Value

	switch args.Test {
	case ``, `eq`:
		if typeutil.String(have) != typeutil.String(want) {
			return fmt.Errorf("expected %q == %q", have, want)
		}
	case `ne`:
		if typeutil.String(have) == typeutil.String(want) {
			return fmt.Errorf("expected %q != %q", have, want)
		}
	case `contains`:
		if !strings.Contains(typeutil.String(have), typeutil.String(want)) {
			return fmt.Errorf("expected %q to contain %q", have, want)
		}
	case `gt`:
		if typeutil.Float(have) <= typeutil.Float(want) {
			return fmt.Errorf("expected %v > %v", have, want)
		}
	case `gte`:
		if typeutil.Float(have) < typeutil.Float(want) {
			return fmt.Errorf("expected %v >= %v", have, want)
		}
	case `lt`:
		if typeutil.Float(have) >= typeutil.Float(want) {
			return fmt.Errorf("expected %v < %v", have, want)
		}
	case `lte`:
		if typeutil.Float(have) > typeutil.Float(want) {
			return fmt.Errorf("expected %v <= %v", have, want)
		}
	default:
		return fmt.Errorf("invalid comparison %q", args.Test)
	}

	return nil
}

// Return an error if the given value is not equal to the other value.
func (self *Commands) Equal(have interface{}, args *BinaryComparison) error {
	return self.Compare(have, bc(args, `eq`))
}

// Return an error if the given value is equal to the other value.
func (self *Commands) NotEqual(have interface{}, args *BinaryComparison) error {
	return self.Compare(have, bc(args, `ne`))
}

// Return an error if the given value does not contain another value.
func (self *Commands) Contains(have interface{}, args *BinaryComparison) error {
	return self.Compare(have, bc(args, `contains`))
}

// Return an error if the given value contains another value.
func (self *Commands) NotContains(have interface{}, args *BinaryComparison) error {
	return self.Compare(have, bc(args, `not-contains`))
}

// Return an error if the given value is not numerically greater than the second value.
func (self *Commands) Gt(have interface{}, args *BinaryComparison) error {
	return self.Compare(have, bc(args, `gt`))
}

// Return an error if the given value is not numerically greater than or equal to the second value.
func (self *Commands) Gte(have interface{}, args *BinaryComparison) error {
	return self.Compare(have, bc(args, `gte`))
}

// Return an error if the given value is not numerically less than the second value.
func (self *Commands) Lt(have interface{}, args *BinaryComparison) error {
	return self.Compare(have, bc(args, `lt`))
}

// Return an error if the given value is not numerically less than or equal to the second value.
func (self *Commands) Lte(have interface{}, args *BinaryComparison) error {
	return self.Compare(have, bc(args, `lte`))
}
