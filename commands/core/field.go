package core

import (
	"github.com/PerformLine/go-performline-stdlib/log"
	defaults "github.com/ghetzel/go-defaults"
	"github.com/ghetzel/go-webfriend/browser"
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
func (self *Commands) Field(selector browser.Selector, args *FieldArgs) ([]*browser.Element, error) {
	if args == nil {
		args = &FieldArgs{}
	}

	defaults.SetDefaults(args)
	dom := self.browser.Tab().DOM()

	if elements, err := dom.Query(selector, nil); err == nil {
		log.Debugf("els %+v", elements)

		for _, field := range elements {
			log.Notice("NOTICE ME SENPAI")

			if err := field.SetAttribute(`value`, ``); err != nil {
				return nil, err
			}

			if err := field.Focus(); err != nil {
				return nil, err
			}

			if _, err := self.Type(args.Value, nil); err != nil {
				return nil, err
			}

			if args.Enter {
				self.Key(`Enter`, &KeyArgs{
					KeyCode: 13,
				})
			}

		}

		return elements, err
	} else {
		return nil, err
	}
}
