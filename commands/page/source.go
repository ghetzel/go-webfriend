package page

import (
	"github.com/ghetzel/go-stockutil/typeutil"
	"github.com/ghetzel/go-webfriend/browser"
)

// Return the source for the given element, or for the whole page.
func (self *Commands) Source(selector browser.Selector) (string, error) {
	if source, err := self.browser.Tab().DOM().EvaluateOn(
		selector,
		`return this.outerHTML`,
	); err == nil {
		return typeutil.V(source).String(), nil
	} else {
		return ``, err
	}
}

// Return the text content for the given element, or for the whole page.
func (self *Commands) Text(selector browser.Selector) (string, error) {
	if source, err := self.browser.Tab().DOM().EvaluateOn(
		selector,
		`return this.outerText`,
	); err == nil {
		return typeutil.V(source).String(), nil
	} else {
		return ``, err
	}
}
