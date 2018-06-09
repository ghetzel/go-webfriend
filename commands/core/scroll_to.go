package core

import (
	"fmt"

	"github.com/ghetzel/go-webfriend/browser"
)

// Scroll the viewport to the location of the first element matched by selector.
func (self *Commands) ScrollTo(selector browser.Selector) error {
	if _, err := self.browser.Tab().DOM().Root(); err == nil {
		// _, err := root.Evaluate(fmt.Sprintf("window.scrollTo(%d, %d)")
		return fmt.Errorf(`Not Implemented`)
	} else {
		return err
	}
}

// Scroll the viewport to the given X,Y coordinates relative to the top-left of
// the current page.
func (self *Commands) ScrollToCoords(x int, y int) error {
	return fmt.Errorf(`Not Implemented`)
}
