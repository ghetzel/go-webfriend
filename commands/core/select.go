package core

import (
	"fmt"

	defaults "github.com/ghetzel/go-defaults"
	"github.com/ghetzel/go-webfriend/browser"
)

type SelectArgs struct {
	// Whether no matches returns an error or not.
	CanBeEmpty bool `json:"can_be_empty" default:"false"`
}

// Polls the DOM for an element that matches the given selector. Either the
// element will be found and returned within the given timeout, or a
// TimeoutError will be returned.
func (self *Commands) Select(selector browser.Selector, args *SelectArgs) ([]*browser.Element, error) {
	if args == nil {
		args = &SelectArgs{}
	}

	defaults.SetDefaults(args)
	dom := self.browser.Tab().DOM()

	if elements, err := dom.Query(selector, nil); err == nil || browser.IsElementNotFoundErr(err) {
		if len(elements) > 0 {
			return elements, nil
		} else if args.CanBeEmpty {
			return nil, nil
		} else {
			return nil, fmt.Errorf("No elements matched the given selector")
		}
	} else {
		return nil, err
	}
}
