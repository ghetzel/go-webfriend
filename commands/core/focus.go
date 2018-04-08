package core

import (
	"github.com/ghetzel/go-webfriend/browser"
)

// Focuses the given HTML element described by selector. One and only one element may match the selector.
func (self *Commands) Focus(selector browser.Selector) (*browser.Element, error) {
	if elements, err := self.Select(selector, nil); err == nil && len(elements) == 1 {
		if err := elements[0].Focus(); err == nil {
			return elements[0], err
		} else {
			return nil, err
		}
	} else if l := len(elements); l > 1 {
		return nil, browser.TooManyMatchesErr(selector, 1, l)
	} else {
		return nil, err
	}
}
