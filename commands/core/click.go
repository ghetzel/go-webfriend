package core

import (
	"fmt"
	"regexp"
	"time"

	defaults "github.com/ghetzel/go-defaults"
	"github.com/ghetzel/go-webfriend/browser"
	"github.com/ghetzel/go-webfriend/dom"
	"github.com/ghetzel/go-webfriend/utils"
)

type ClickArgs struct {
	// Permit multiple elements to be clicked.
	Multiple bool `json:"value"`

	// If Multiple clicks are permitted, what is the delay between each click.
	Delay time.Duration `json:"delay" default:"20ms"`

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
func (self *Commands) Click(selector dom.Selector, args *ClickArgs) ([]*dom.Element, error) {
	if args == nil {
		args = &ClickArgs{}
	}

	defaults.SetDefaults(args)
	args.Delay = utils.FudgeDuration(args.Delay)

	if elements, err := self.Select(selector, nil); err == nil {
		if mt := args.MatchText; mt != `` {
			if rx, err := regexp.Compile(mt); err == nil {
				var matches = make([]*dom.Element, 0)

				for _, el := range elements {
					if rx.MatchString(el.Text) {
						matches = append(matches, el)
					}
				}

				elements = matches
			} else {
				return nil, fmt.Errorf("match_text: %v", err)
			}
		}

		if len(elements) == 1 || args.Multiple {
			if _, err := self.browser.Tab().Evaluate(fmt.Sprintf(
				"document.querySelectorAll(%q).forEach(function(i){ i.click() })",
				selector,
			)); err != nil {
				return nil, err
			}

			return elements, nil
		} else {
			return nil, dom.TooManyMatchesErr(selector, 1, len(elements))
		}
	} else {
		return nil, err
	}
}

type ClickAtArgs struct {
	// The X-coordinate to click at
	X int `json:"x"`

	// The Y-coordinate to click at
	Y int `json:"y"`
}

// Click the page at the given X, Y coordinates.
func (self *Commands) ClickAt(args *ClickAtArgs) ([]dom.Element, error) {
	return nil, browser.NotImplemented
}
