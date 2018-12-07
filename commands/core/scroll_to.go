package core

import (
	"github.com/ghetzel/go-webfriend/browser"
	"github.com/ghetzel/go-webfriend/dom"
)

// Scroll the viewport to the location of the first element matched by selector.
func (self *Commands) ScrollTo(selector dom.Selector) error {
	return browser.NotImplemented
	// if _, err := self.browser.Tab().DOM().Root(); err == nil {
	// 	// _, err := root.Evaluate(fmt.Sprintf("window.scrollTo(%d, %d)")
	// } else {
	// 	return err
	// }
}

// Scroll the viewport to the given X,Y coordinates relative to the top-left of
// the current page.
func (self *Commands) ScrollToCoords(x int, y int) error {
	return browser.NotImplemented
}
