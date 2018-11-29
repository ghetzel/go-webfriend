// Suite of testing-oriented commands that will trigger errors or failures if they aren't satistifed.
//
// The Assert module is aimed at making it easier to write scripts that perform test validation,
// as is frequently done during automated acceptance testing of web applications.
//
package assert

import (
	"fmt"

	"github.com/ghetzel/friendscript"
	"github.com/ghetzel/friendscript/utils"
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
