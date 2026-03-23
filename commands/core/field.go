package core

import (
	defaults "github.com/ghetzel/go-defaults"
	"github.com/ghetzel/go-stockutil/typeutil"
	"github.com/ghetzel/go-webfriend/browser"
	"github.com/ghetzel/go-webfriend/dom"
)

type FieldArgs struct {
	// The value to enter into the field.
	Value any `json:"value"`

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
//
//	field '#username' {
//	  value: 'myuser',
//	}
//
//	field '#password' {
//	  value: 'p@ssw0rd!',
//	  enter: true,
//	}
//
// ```
func (self *Commands) Field(selector dom.Selector, args *FieldArgs) error {
	if args == nil {
		args = &FieldArgs{}
	}

	defaults.SetDefaults(args)

	if pg := self.browser.Page(); pg != nil {
		var query = pg.Locator(string(selector))

		if args.Autoclear {
			if err := query.Clear(); err != nil {
				return err
			}
		}

		if err := query.Fill(typeutil.String(args.Value)); err != nil {
			return err
		}

		if args.Enter {
			return query.Press(`Enter`)
		} else {
			return nil
		}
	} else {
		return browser.NoActivePage
	}
}

type CheckboxArgs struct {
	// Whether the field should be made active or inactive.
	Active any `json:"active" default:"true"`
}

// Make a checkbox or radio button active or inactive.
//
// #### Examples
//
// ##### Type in a username and password, then hit Enter to submit.
// ```
//
//	field '#username' {
//	  value: 'myuser',
//	}
//
//	field '#password' {
//	  value: 'p@ssw0rd!',
//	  enter: true,
//	}
//
// ```
func (self *Commands) Toggle(selector dom.Selector, args *CheckboxArgs) error {
	if args == nil {
		args = new(CheckboxArgs)
	}

	defaults.SetDefaults(args)

	if pg := self.browser.Page(); pg != nil {
		var query = pg.Locator(string(selector))

		if typeutil.Bool(args.Active) {
			return query.Check()
		} else {
			return query.Uncheck()
		}
	} else {
		return browser.NoActivePage
	}
}
