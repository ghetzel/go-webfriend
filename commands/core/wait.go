package core

import (
	"fmt"
	"time"

	"github.com/ghetzel/go-stockutil/log"
	"github.com/ghetzel/go-stockutil/stringutil"
	"github.com/ghetzel/go-stockutil/timeutil"
	defaults "github.com/mcuadros/go-defaults"
)

var WaitForLoadEventName string = `Page.loadEventFired`

// Pauses execution of the current script for the given duration.
func (self *Commands) Wait(delay interface{}) error {
	var duration time.Duration

	if delayD, ok := delay.(time.Duration); ok {
		duration = delayD
	} else if delayMs, err := stringutil.ConvertToInteger(delay); err == nil {
		duration = time.Duration(delayMs) * time.Millisecond
	} else if delayParsed, err := timeutil.ParseDuration(fmt.Sprintf("%v", delay)); err == nil {
		duration = delayParsed
	} else {
		return fmt.Errorf("invalid duration: %v", err)
	}

	log.Infof("Waiting for %v", duration)
	time.Sleep(duration)
	return nil
}

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
