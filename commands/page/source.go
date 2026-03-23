package page

import (
	"github.com/ghetzel/go-webfriend/browser"
	"github.com/ghetzel/go-webfriend/dom"
)

// Return the source for the given element, or for the whole page.
func (self *Commands) Source(selector dom.Selector) (string, error) {
	if pg := self.browser.Page(); pg != nil {
		if selector == `` {
			return pg.Content()
		} else {
			return pg.Locator(string(selector)).InnerHTML()
		}
	} else {
		return ``, browser.NoActivePage
	}
}

// Return the text content for the given element, or for the whole page.
func (self *Commands) Text(selector dom.Selector) (string, error) {
	return ``, browser.NotImplemented
}
