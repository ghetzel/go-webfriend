package core

import (
	"time"

	"github.com/playwright-community/playwright-go"
	defaults "go.gary.cool/go-defaults"
	"go.gary.cool/go-stockutil/typeutil"
	"go.gary.cool/go-webfriend/browser"
	"go.gary.cool/go-webfriend/dom"
	"go.gary.cool/go-webfriend/utils"
)

type ClickArgs struct {
	// Number of clicks to perform.
	Count int `json:"count" default:"1"`

	// Time to waith between mousedown and mouseup events.
	Delay time.Duration `json:"delay" default:"20ms"`

	// Which mouse button to simulate during click. Values are "left", "right", "middle". Default "left".
	Button string `json:"button" default:"left"`

	// If provided, this represents a regular expression that the text value of matching elements must match to be clicked.
	MatchText string `json:"match_text"`
}

// Click on HTML element(s) matches by selector.  If multiple is true, then all
// elements matched by selector will be clicked in the order they are returned.
// Otherwise, an error is returned unless selector matches exactly one element.
//
// #### Examples
//
// ##### Click on the element with id "login"
// ```
// click "#login"
// ```
//
// ##### Click on all `<a>` elements on the page, waiting 150ms between each click.
// ```
//
//	click "a" {
//	  multiple: true,
//	  delay:    "150ms",
//	}
//
// ```
func (self *Commands) Click(selector dom.Selector, args *ClickArgs) error {
	if args == nil {
		args = new(ClickArgs)
	}

	defaults.SetDefaults(args)
	args.Delay = utils.FudgeDuration(args.Delay)

	var btn *playwright.MouseButton

	switch args.Button {
	case `middle`:
		btn = playwright.MouseButtonMiddle
	case `right`:
		btn = playwright.MouseButtonRight
	default:
		btn = playwright.MouseButtonLeft
	}

	if pg := self.browser.Page(); pg != nil {
		return pg.Locator(string(selector), playwright.PageLocatorOptions{
			HasText: args.MatchText,
		}).Click(playwright.LocatorClickOptions{
			ClickCount: playwright.Int(args.Count),
			Delay:      playwright.Float(float64(args.Delay.Milliseconds())),
			Button:     btn,
		})
	} else {
		return browser.NoActivePage
	}
}

type ClickAtArgs struct {
	// The X-coordinate to click at
	X int `json:"x"`

	// The Y-coordinate to click at
	Y int `json:"y"`

	// Number of steps to interpolate when moving.
	Steps int `json:"steps"`
}

// Click the page at the given X, Y coordinates.
func (self *Commands) ClickAt(args *ClickAtArgs) error {
	if pg := self.browser.Page(); pg != nil {
		return pg.Mouse().Move(float64(args.X), float64(args.Y), playwright.MouseMoveOptions{
			Steps: playwright.Int(typeutil.OrNInt(args.Steps, 1)),
		})
	} else {
		return browser.NoActivePage
	}
}
