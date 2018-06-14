package core

import (
	"fmt"
	"time"

	"github.com/ghetzel/go-webfriend/browser"
	defaults "github.com/mcuadros/go-defaults"
)

type ClickArgs struct {
	// The value to enter into the field.
	Multiple bool `json:"value"`

	// If Multiple clicks are permitted, what is the delay between each click.
	Delay time.Duration `json:"delay" default:"20ms"`
}

// Click on HTML element(s) matches by selector.  If multiple is true, then all
// elements matched by selector will be clicked in the order they are returned.
// Otherwise, an error is returned unless selector matches exactly one element.
func (self *Commands) Click(selector browser.Selector, args *ClickArgs) ([]*browser.Element, error) {
	if args == nil {
		args = &ClickArgs{}
	}

	defaults.SetDefaults(args)

	if elements, err := self.Select(selector, nil); err == nil {
		if len(elements) == 1 || args.Multiple {
			for i, element := range elements {
				if i > 0 && args.Delay > 0 {
					time.Sleep(args.Delay)
				}

				if err := element.Click(); err != nil {
					return elements[0:i], err
				}
			}

			return elements, nil
		} else {
			return nil, browser.TooManyMatchesErr(selector, 1, len(elements))
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
func (self *Commands) ClickAt(args *ClickAtArgs) ([]browser.Element, error) {
	return nil, fmt.Errorf(`Not Implemented`)
}
