package core

import (
	"github.com/ghetzel/go-webfriend/browser"
	defaults "github.com/mcuadros/go-defaults"
)

type FieldArgs struct {
	// The value to enter into the field.
	Value interface{} `json:"value"`

	// Whether to clear the existing contents of the field before entering new data.
	Autoclear bool `json:"autoclear" default:"true"`

	// Whether to automatically send an "Enter" keystroke after typing in the given value
	Enter bool `json:"enter"`
}

// Locate and enter data into a form input field.
func (self *Commands) Field(selector browser.Selector, args *FieldArgs) (string, error) {
	if args == nil {
		args = &FieldArgs{}
	}

	defaults.SetDefaults(args)

	if elements, err := self.Select(selector, nil); err == nil && len(elements) == 1 {
		field := elements[0]

		if err := field.SetAttribute(`value`, ``); err != nil {
			return ``, err
		}

		if err := field.Focus(); err != nil {
			return ``, err
		}

		_, err := self.Type(args.Value, nil)

		if args.Enter {
			self.Key(`Enter`, &KeyArgs{
				KeyCode: 13,
			})
		}

		return field.Text(), err
	} else if l := len(elements); l > 1 {
		return ``, browser.TooManyMatchesErr(selector, 1, l)
	} else {
		return ``, err
	}
}
