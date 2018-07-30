package core

import (
	"time"

	"github.com/ghetzel/go-webfriend/utils"
	defaults "github.com/mcuadros/go-defaults"
)

var WaitForLoadEventName string = `Page.loadEventFired`

type WaitForArgs struct {
	// The timeout before we stop waiting for the event.
	Timeout time.Duration `json:"timeout" default:"30s"`
}

// Wait for a specific event or events matching the given glob pattern, up to an
// optional Timeout duration.
func (self *Commands) WaitFor(event string, args *WaitForArgs) error {
	if args == nil {
		args = &WaitForArgs{}
	}

	defaults.SetDefaults(args)
	args.Timeout = utils.FudgeDuration(args.Timeout)

	if waiter, err := self.browser.Tab().CreateEventWaiter(event); err == nil {
		defer waiter.Remove()

		// wait for the first event matching the given pattern
		_, err := waiter.Wait(args.Timeout)
		return err
	} else {
		return err
	}
}

// Wait for a page load event.
func (self *Commands) WaitForLoad(args *WaitForArgs) error {
	return self.WaitFor(WaitForLoadEventName, args)
}
