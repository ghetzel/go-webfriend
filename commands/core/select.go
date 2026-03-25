package core

import (
	"fmt"

	defaults "github.com/ghetzel/go-defaults"
	"github.com/ghetzel/go-stockutil/stringutil"
	"github.com/ghetzel/go-webfriend/browser"
	"github.com/ghetzel/go-webfriend/dom"
	"github.com/playwright-community/playwright-go"
)

type SelectArgs struct {
	// The timeout before we stop waiting for the element to appear (in milliseconds.)
	Timeout int `json:"timeout" default:"1000"`

	// Waits for matching elements to be in a particular state. Values are "visible", "hidden", "attached", "detached". Default is "visible".
	State string `json:"state"`
}

// Retrieve the first element matching the given selector.
func (self *Commands) Select(selector dom.Selector, args *SelectArgs) (*dom.Element, error) {
	if matches, err := self.SelectAll(selector, args); err == nil {
		if len(matches) > 0 {
			return matches[0], nil
		} else {
			return nil, nil
		}
	} else {
		return nil, err
	}
}

// Retrieve all elements matching the given selector.
func (self *Commands) SelectAll(selector dom.Selector, args *SelectArgs) ([]*dom.Element, error) {
	if args == nil {
		args = new(SelectArgs)
	}

	defaults.SetDefaults(args)

	if pg := self.browser.Page(); pg != nil {
		var query playwright.Locator
		var syntax, expr = stringutil.SplitPairTrailing(string(selector), `=`)

		switch syntax {
		case `alt`:
			query = pg.GetByAltText(expr)
		case `label`:
			query = pg.GetByLabel(expr)
		case `placeholder`:
			query = pg.GetByPlaceholder(expr)
		case `role`:
			query = pg.GetByRole(playwright.AriaRole(expr))
		case `test`:
			query = pg.GetByTestId(expr)
		case `text`:
			query = pg.GetByText(expr)
		case `title`:
			query = pg.GetByTitle(expr)
		default:
			query = pg.Locator(string(selector))
		}

		if args.State != `` {
			var state *playwright.WaitForSelectorState

			switch args.State {
			case `hidden`:
				state = playwright.WaitForSelectorStateHidden
			case `attached`:
				state = playwright.WaitForSelectorStateAttached
			case `detached`:
				state = playwright.WaitForSelectorStateDetached
			case `visible`:
				state = playwright.WaitForSelectorStateVisible
			default:
				return nil, fmt.Errorf("invalid state %v", state)
			}

			if err := query.WaitFor(playwright.LocatorWaitForOptions{
				State:   state,
				Timeout: playwright.Float(float64(args.Timeout)),
			}); err != nil {
				return nil, nil
			}
		}

		return dom.FromPlaywright(query), nil
	} else {
		return nil, browser.NoActivePage
	}
}
