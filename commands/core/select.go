package core

import (
	"fmt"
	"time"

	defaults "github.com/ghetzel/go-defaults"
	"github.com/ghetzel/go-webfriend/dom"
	"github.com/ghetzel/go-webfriend/utils"
)

type SelectArgs struct {
	// The timeout before we stop waiting for the element to appear.
	Timeout time.Duration `json:"timeout" default:"5s"`

	// The minimum number of matches necessary to be considered a successful match.
	MinMatches int `json:"min_matches" default:"1"`

	// The polling interval between element re-checks.
	Interval time.Duration `json:"interval" default:"125ms"`
}

// Polls the DOM for an element that matches the given selector. Either the
// element will be found and returned within the given timeout, or a
// TimeoutError will be returned.
func (self *Commands) Select(selector dom.Selector, args *SelectArgs) ([]*dom.Element, error) {
	if args == nil {
		args = &SelectArgs{}
	}

	defaults.SetDefaults(args)
	args.Timeout = utils.FudgeDuration(args.Timeout)
	args.Interval = utils.FudgeDuration(args.Interval)

	started := time.Now()

	for time.Since(started) <= args.Timeout {
		if elements, err := self.browser.Tab().ElementQuery(selector, nil); err == nil {
			if len(elements) >= args.MinMatches {
				return elements, nil
			}
		}

		time.Sleep(args.Interval)
	}

	// If we timed out but are allowed to have zero hits, then it's not an error.
	if args.MinMatches == 0 {
		return make([]*dom.Element, 0), nil
	} else {
		return nil, fmt.Errorf("Timed out waiting for '%v'", selector)
	}
}
