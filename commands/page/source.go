package page

import (
	"github.com/ghetzel/go-stockutil/typeutil"
	"github.com/ghetzel/go-webfriend/dom"
)

// Return the source for the given element, or for the whole page.
func (self *Commands) Source(selector dom.Selector) (string, error) {
	var source interface{}
	var err error

	if selector.IsNone() {
		source, err = self.browser.Tab().Evaluate(`return document.documentElement.outerHTML`)
	} else if elements, err := self.browser.Tab().ElementQuery(selector, nil); err == nil && len(elements) > 0 {
		src := ``

		for _, element := range elements {
			if s, err := self.browser.Tab().EvaluateOn(element, `return this.outerHTML`); err == nil {
				src += typeutil.String(s)
			} else {
				return ``, err
			}
		}

		source = src
	}

	if err == nil && source != nil {
		return typeutil.V(source).String(), nil
	} else {
		return ``, err
	}
}

// Return the text content for the given element, or for the whole page.
func (self *Commands) Text(selector dom.Selector) (string, error) {
	var source interface{}
	var err error

	if selector.IsNone() {
		source, err = self.browser.Tab().Evaluate(`return document.documentElement.innerText`)

	} else if elements, err := self.browser.Tab().ElementQuery(selector, nil); err == nil && len(elements) > 0 {
		txt := ``

		for _, element := range elements {
			if s, err := self.browser.Tab().EvaluateOn(element, `return this.innerText`); err == nil {
				txt += typeutil.String(s)
			} else {
				return ``, err
			}
		}

		source = txt
	}

	if err == nil {
		return typeutil.V(source).String(), nil
	} else {
		return ``, err
	}
}
