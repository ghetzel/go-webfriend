package core

import (
	"fmt"
	"time"

	"github.com/ghetzel/go-stockutil/log"
	"github.com/ghetzel/go-webfriend/browser"
	"github.com/ghetzel/go-webfriend/utils"
	defaults "github.com/mcuadros/go-defaults"
)

type SelectArgs struct {
	// The maximum amount of time to wait for at least MinMatches elements to match.
	Timeout time.Duration `json:"timeout" default:"5s"`

	// How often to poll the page looking for matching elements.
	Interval time.Duration `json:"interval" default:"125ms"`

	// The minimum number of elements that need to match to consider the selection a success.
	MinMatches int `json:"min_matches" default:"1"`

	// Whether no matches returns an error or not.
	CanBeEmpty bool `json:"can_be_empty" default:"false"`
}

// Polls the DOM for an element that matches the given selector. Either the
// element will be found and returned within the given timeout, or a
// TimeoutError will be returned.
func (self *Commands) Select(selector browser.Selector, args *SelectArgs) ([]*browser.Element, error) {
	if args == nil {
		args = &SelectArgs{}
	}

	defaults.SetDefaults(args)
	args.Timeout = utils.FudgeDuration(args.Timeout)
	args.Interval = utils.FudgeDuration(args.Interval)

	docroot := self.browser.Tab().DOM()
	start := time.Now()
	deadline := start.Add(args.Timeout)
	i := 0

	// repeatedly perform the query, until at least MinMatches elements are returned
	// or until the deadline is exceeded
	for t := start; t.Before(deadline); t = time.Now() {
		if i > 0 {
			log.Debugf("Polling for elements matching: %v", selector)
		}

		if elements, err := docroot.Query(selector, nil); err == nil || browser.IsElementNotFoundErr(err) {
			if len(elements) >= args.MinMatches {
				return elements, err
			} else {
				i++
				time.Sleep(args.Interval)
			}
		} else {
			return nil, err
		}
	}

	if args.CanBeEmpty {
		return nil, nil
	} else {
		return nil, fmt.Errorf("No elements matched the given selector")
	}
}
