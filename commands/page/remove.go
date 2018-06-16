package page

import (
	"fmt"

	"github.com/ghetzel/go-webfriend/browser"
	defaults "github.com/mcuadros/go-defaults"
)

type RemoveArgs struct {
	Parent browser.Selector `json:"parent"`
}

// Remove all occurrences of the element(s) matching the given selector.
func (self *Commands) Remove(selector browser.Selector, args *RemoveArgs) (int, error) {
	if args == nil {
		args = &RemoveArgs{}
	}

	defaults.SetDefaults(args)

	if !selector.IsNone() {
		docroot := self.browser.Tab().DOM()

		var parent *browser.Element

		// if a parent selector was specified, find that element
		if !args.Parent.IsNone() {
			if elements, err := docroot.Query(args.Parent, nil); err == nil {
				if len(elements) == 1 {
					parent = elements[0]
				} else {
					return 0, fmt.Errorf("Ambiguous parent selector: got %d results:", len(elements))
				}
			} else {
				return 0, err
			}
		}

		// query for the elements to remove from the found parent, or throughout the whole
		// document if no parent was given.
		if elements, err := docroot.Query(selector, parent); err == nil {
			n := 0

			for _, element := range elements {
				if err := element.Remove(); err == nil {
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
