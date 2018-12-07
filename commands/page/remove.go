package page

import (
	defaults "github.com/ghetzel/go-defaults"
	"github.com/ghetzel/go-webfriend/dom"
)

type RemoveArgs struct {
	Parent dom.Selector `json:"parent"`
}

// Remove all occurrences of the element(s) matching the given selector.
func (self *Commands) Remove(selector dom.Selector, args *RemoveArgs) (int, error) {
	if args == nil {
		args = &RemoveArgs{}
	}

	defaults.SetDefaults(args)

	if !selector.IsNone() {
		// query for the elements to remove from the found parent, or throughout the whole
		// document if no parent was given.
		if elements, err := self.browser.Tab().ElementQuery(selector, &args.Parent); err == nil {
			n := 0

			for _, element := range elements {
				if _, err := self.browser.Tab().EvaluateOn(element, `this.remove()`); err == nil {
					n += 1
				}
			}

			return n, nil
		} else {
			return 0, err
		}
	} else {
		return 0, nil
	}
}
