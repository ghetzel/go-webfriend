package core

import (
	"fmt"

	"github.com/ghetzel/go-webfriend/browser"
)

// Click on HTML element(s) matches by selector.  If multiple is true, then all
// elements matched by selector will be clicked in the order they are returned.
// Otherwise, an error is returned unless selector matches exactly one element.
func (self *Commands) Click(selector browser.Selector, multiple bool) ([]browser.Element, error) {
	return nil, fmt.Errorf(`NI`)
}

// Click the page at the given X, Y coordinates.
func (self *Commands) ClickAt(x int, y int) ([]browser.Element, error) {
	return nil, fmt.Errorf(`NI`)
}
