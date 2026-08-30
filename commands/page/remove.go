package page

import (
	defaults "go.gary.cool/go-defaults"
	"go.gary.cool/go-webfriend/browser"
	"go.gary.cool/go-webfriend/dom"
)

type RemoveArgs struct {
	Parent dom.Selector `json:"parent"`
}

// Remove all occurrences of the element(s) matching the given selector.
func (self *Commands) Remove(selector dom.Selector, args *RemoveArgs) (int, error) {
	if args == nil {
		args = new(RemoveArgs)
	}

	defaults.SetDefaults(args)

	if !selector.IsNone() {
		// query for the elements to remove from the found parent, or throughout the whole
		// document if no parent was given.
		return 0, browser.NotImplemented
	} else {
		return 0, nil
	}
}
