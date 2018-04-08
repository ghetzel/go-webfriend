package core

import (
	"fmt"

	"github.com/ghetzel/go-webfriend/browser"
)

type FieldArgs struct {
	// The value to enter into the field.
	Value interface{} `json:"value"`

	// Whether to clear the existing contents of the field before entering new data.
	Autoclear bool `json:"autoclear" default:"true"`
}

// Locate and enter data into a form input field.
func (self *Commands) Field(selector browser.Selector, args *FieldArgs) (string, error) {
	if elements, err := self.Select(selector, nil); err == nil && len(elements) == 1 {
		field := elements[0]

		if err := field.SetAttribute(`value`, ``); err != nil {
			return ``, err
		}

		if err := field.Focus(); err != nil {
			return ``, err
		}

		_, err := self.Type(args.Value, nil)
		return field.Text(), err
	} else if l := len(elements); l > 1 {
		return ``, fmt.Errorf("Too many elements matched %q; expected 1, got %d", selector, l)
	} else {
		return ``, err
	}
}
