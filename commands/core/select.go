package core

import (
	"time"

	"github.com/ghetzel/go-webfriend/browser"
)

type SelectArgs struct {
	// The maximum amount of time to wait for at least one element to match.
	Timeout time.Duration `json:"timeout"`

	// How often to poll the page looking for matching elements.
	Interval time.Duration `json:"interval"`
}

// Polls the DOM for an element that matches the given selector. Either the
// element will be found and returned within the given timeout, or a
// TimeoutError will be returned.
func (self *Commands) Select(selector browser.Selector, args *SelectArgs) ([]browser.Element, error) {
	dom := self.browser.Tab().DOM()
	dom.PrintTree()
	return dom.Query(selector, nil)
}
