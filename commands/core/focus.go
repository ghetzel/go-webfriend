package core

import (
	"time"

	defaults "github.com/ghetzel/go-defaults"
	"github.com/ghetzel/go-webfriend/browser"
	"github.com/ghetzel/go-webfriend/dom"
	"github.com/playwright-community/playwright-go"
)

type FocusArgs struct {
	// The amount of time to wait for the element to focus.
	Timeout time.Duration `json:"timeout" default:"30s"`
}

// Focuses the given HTML element described by selector. One and only one element may match the selector.
func (self *Commands) Focus(selector dom.Selector, args *FocusArgs) error {
	if args == nil {
		args = new(FocusArgs)
	}

	defaults.SetDefaults(args)

	if pg := self.browser.Page(); pg != nil {
		return pg.Locator(string(selector)).Focus(playwright.LocatorFocusOptions{
			Timeout: playwright.Float(float64(args.Timeout.Milliseconds())),
		})
	} else {
		return browser.NoActivePage
	}
}
