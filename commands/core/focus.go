package core

import (
	"fmt"

	"github.com/ghetzel/go-webfriend/dom"
)

// Focuses the given HTML element described by selector. One and only one element may match the selector.
func (self *Commands) Focus(selector dom.Selector) (*dom.Element, error) {
	if elements, err := self.Select(selector, nil); err == nil && len(elements) == 1 {
		if _, err := self.browser.Tab().Evaluate(fmt.Sprintf(
			"document.querySelector(%q).focus()",
			selector,
		)); err == nil {
			return elements[0], nil
		} else {
			return nil, err
		}
	} else if l := len(elements); l > 1 {
		return nil, dom.TooManyMatchesErr(selector, 1, l)
	} else {
		return nil, err
	}
}
