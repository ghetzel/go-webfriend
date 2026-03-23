package core

import (
	"time"

	defaults "github.com/ghetzel/go-defaults"
	"github.com/ghetzel/go-stockutil/log"
	"github.com/ghetzel/go-webfriend/browser"
	"github.com/ghetzel/go-webfriend/dom"
	"github.com/playwright-community/playwright-go"
)

type SelectArgs struct {
	// The timeout before we stop waiting for the element to appear.
	Timeout time.Duration `json:"timeout" default:"30s"`

	// Waits for matching elements to be in a particular state. Values are "visible", "hidden", "attached", "detached". Default is "visible".
	State string `json:"state" default:"visible"`
}

// Polls the DOM for an element that matches the given selector. Either the
// element will be found and returned within the given timeout, or a
// TimeoutError will be returned.
func (self *Commands) Select(selector dom.Selector, args *SelectArgs) ([]*dom.Element, error) {
	if args == nil {
		args = new(SelectArgs)
	}

	defaults.SetDefaults(args)

	if pg := self.browser.Page(); pg != nil {
		var state *playwright.WaitForSelectorState

		switch args.State {
		case `hidden`:
			state = playwright.WaitForSelectorStateHidden
		case `attached`:
			state = playwright.WaitForSelectorStateAttached
		case `detached`:
			state = playwright.WaitForSelectorStateDetached
		default:
			state = playwright.WaitForSelectorStateVisible
		}

		var query = pg.Locator(string(selector))

		query.WaitFor(playwright.LocatorWaitForOptions{
			State:   state,
			Timeout: playwright.Float(float64(args.Timeout.Milliseconds())),
		})

		log.Debugf("%v (%d len)", selector, len(dom.FromPlaywright(query)))

		return dom.FromPlaywright(query), nil
	} else {
		return nil, browser.NoActivePage
	}
}
