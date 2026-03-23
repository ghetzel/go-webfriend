package core

import (
	"fmt"

	"github.com/ghetzel/go-webfriend/browser"
	"github.com/ghetzel/go-webfriend/dom"
)

// Scroll the viewport to the location of the first element matched by selector.
func (self *Commands) ScrollTo(selector dom.Selector) error {
	if pg := self.browser.Page(); pg != nil {
		if bbox, err := pg.Locator(string(selector)).BoundingBox(); err == nil {
			var _, err = pg.Evaluate(
				fmt.Sprintf("window.scrollTo(%0f, %0f)", bbox.X, bbox.Y),
			)
			return err
		} else {
			return err
		}
	} else {
		return browser.NoActivePage
	}
}

// Scroll the viewport to the given X,Y coordinates relative to the top-left of
// the current page.
func (self *Commands) ScrollToCoords(x int, y int) error {
	if pg := self.browser.Page(); pg != nil {
		var _, err = pg.Evaluate(
			fmt.Sprintf("window.scrollTo(%d, %d)", x, y),
		)
		return err
	} else {
		return browser.NoActivePage
	}
}
