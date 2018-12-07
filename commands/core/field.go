package core

import (
	"fmt"
	"time"

	defaults "github.com/ghetzel/go-defaults"
	"github.com/ghetzel/go-webfriend/dom"
)

type FieldArgs struct {
	// The value to enter into the field.
	Value interface{} `json:"value"`

	// Whether to clear the existing contents of the field before entering new data.
	Autoclear bool `json:"autoclear" default:"true"`

	// Whether to automatically send an "Enter" keystroke after typing in the given value
	Enter bool `json:"enter"`

	// An element to click after the field value is changed.
	Click dom.Selector `json:"click"`
}

// Locate and enter data into a form input field.
//
// #### Examples
//
// ##### Type in a username and password, then hit Enter to submit.
// ```
// field '#username' {
//   value: 'myuser',
// }
//
// field '#password' {
//   value: 'p@ssw0rd!',
//   enter: true,
// }
// ```
//
func (self *Commands) Field(selector dom.Selector, args *FieldArgs) ([]*dom.Element, error) {
	if args == nil {
		args = &FieldArgs{}
	}

	defaults.SetDefaults(args)

	if elements, err := self.Select(selector, nil); err == nil {
		for _, field := range elements {
			if args.Autoclear {
				if _, err := self.browser.Tab().EvaluateOn(field, `this.value = ''`); err != nil {
					return nil, fmt.Errorf("autoclear: %v", err)
				}
			}

			if _, err := self.browser.Tab().EvaluateOn(field, `this.focus()`); err != nil {
				return nil, fmt.Errorf("focus: %v", err)
			}

			if _, err := self.Type(args.Value, nil); err != nil {
				return nil, fmt.Errorf("type: %v", err)
			}

			if args.Enter {
				if err := self.Key(`Enter`, &KeyArgs{
					KeyCode: 13,
				}); err != nil {
					return nil, fmt.Errorf("keypress: %v", err)
				}
			}

			if !args.Click.IsNone() {
				if _, err := self.Click(args.Click, &ClickArgs{
					Multiple: true,
					Delay:    50 * time.Millisecond,
				}); err != nil {
					return nil, fmt.Errorf("click: %v", err)
				}
			}
		}

		return elements, nil
	} else {
		return nil, err
	}
}
