package page

import (
	"github.com/ghetzel/go-stockutil/typeutil"
	"github.com/ghetzel/go-webfriend/browser"
)

// Return the source for the given element, or for the whole page.
func (self *Commands) Source(selector browser.Selector) (string, error) {
	var source interface{}
	var err error

	if selector.IsNone() {
		source, err = self.browser.Tab().DOM().Evaluate(`return document.documentElement.outerHTML`)
	} else {
		source, err = self.browser.Tab().DOM().EvaluateOn(selector, `return this.outerHTML`)
	}

	if err == nil {
		return typeutil.V(source).String(), nil
	} else {
		return ``, err
	}
}

// Return the text content for the given element, or for the whole page.
func (self *Commands) Text(selector browser.Selector) (string, error) {
	var source interface{}
	var err error

	if selector.IsNone() {
		source, err = self.browser.Tab().DOM().Evaluate(`return document.documentElement.outerText`)
	} else {
		source, err = self.browser.Tab().DOM().EvaluateOn(selector, `return this.outerText`)
	}

	if err == nil {
		return typeutil.V(source).String(), nil
	} else {
		return ``, err
	}
}
