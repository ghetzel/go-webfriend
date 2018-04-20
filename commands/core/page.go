package core

import (
	"fmt"

	"github.com/ghetzel/go-stockutil/log"
	"github.com/ghetzel/go-stockutil/maputil"
	"github.com/ghetzel/go-stockutil/typeutil"
	"github.com/ghetzel/go-webfriend/browser"
	defaults "github.com/mcuadros/go-defaults"
)

type NewTabArgs struct {
	// The width and height (in pixels) that the tab should be created with.
	Width int `json:"width"`

	// The height and height (in pixels) that the tab should be created with.
	Height int `json:"height"`

	// Whether to automatically switch to the newly-created tab as the active
	// tab for subsequent commands.
	Autoswitch bool `json:"autoswitch"` // true
}

// Open a new tab and navigate to the given URL.
func (self *Commands) NewTab(url string, args *NewTabArgs) (browser.TabID, error) {
	return ``, fmt.Errorf(`NI`)
}

// Close the tab identified by the given ID.
func (self *Commands) CloseTab(id browser.TabID) error {
	return fmt.Errorf(`NI`)
}

// Switches the active tab to a given tab.
func (self *Commands) SwitchTab(id browser.TabID) (*browser.Tab, error) {
	return nil, fmt.Errorf(`NI`)
}

// Reload the currently active tab.
func (self *Commands) Reload() error {
	return self.browser.Tab().AsyncRPC(`Page`, `reload`, nil)
}

// Stop loading the currently active tab.
func (self *Commands) Stop() error {
	return fmt.Errorf(`NI`)
}

type Orientation string

const (
	Portrait           Orientation = `portraitPrimary`
	Landscape                      = `landscapePrimary`
	PortraitSecondary              = `portraitSecondary`
	LandscapeSecondary             = `landscapeSecondary`
)

type ResizeArgs struct {
	// The width of the screen.
	Width int `json:"width"`

	// The height of the screen.
	Height int `json:"height"`

	// The scaling factor of the content.
	Scale float64 `json:"scale"`

	// Whether to emulate a mobile device or not. If a map is provided, mobile
	// emulation will be enabled and configured using the following keys:
	//
	//    width (int, optional), The width of the mobile screen to emulate.;
	//
	//    height (int, optional), The height of the mobile screen to emulate.;
	//
	//    x (int, optional), The horizontal position of the currently viewable
	//                       portion of the mobile screen.;
	//
	//    y (int, optional), The vertical position of the currently viewable
	//                       portion of the mobile screen.;
	//
	Mobile interface{} `json:"mobile"`

	// Whether to fit the viewport contents to the available area or not.
	FitWindow bool `json:"fit_window"`

	// Which screen orientation to emulate, if any.
	Orientation string `json:"orientation" default:"landscapePrimary"`

	// The angle of the screen to emulate (in degrees; 0-360).
	Angle int `json:"angle"`
}

type ResizeResponse struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

// Resizes the active viewport of the current page using the Chrome Device
// Emulation API. This does not resize the window itself, but rather the area
// the current page interprets the window to be.
//
// This is useful for setting the size of the area that will be rendered for
// screenshots and screencasts, or for testing responsive design elements.
func (self *Commands) Resize(args *ResizeArgs) (*ResizeResponse, error) {
	if args == nil {
		args = &ResizeArgs{}
	}

	defaults.SetDefaults(args)

	rpcArgs := map[string]interface{}{
		`width`:             args.Width,
		`height`:            args.Height,
		`deviceScaleFactor`: args.Scale,
		`mobile`:            typeutil.V(args.Mobile).Bool(),
		`screenOrientation`: map[string]interface{}{
			`type`:  string(args.Orientation),
			`angle`: args.Angle,
		},
	}

	if typeutil.IsMap(args.Mobile) {
		mobile := maputil.M(args.Mobile)

		if v := mobile.Int(`width`); v > 0 {
			rpcArgs[`screenWidth`] = int(v)
		}

		if v := mobile.Int(`height`); v > 0 {
			rpcArgs[`screenHeight`] = int(v)
		}

		if v := mobile.Int(`x`); v > 0 {
			rpcArgs[`positionX`] = int(v)
		}

		if v := mobile.Int(`y`); v > 0 {
			rpcArgs[`positionY`] = int(v)
		}
	}

	if _, err := self.browser.Tab().RPC(`Emulation`, `setDeviceMetricsOverride`, rpcArgs); err == nil {
		return &ResizeResponse{
			Width:  args.Width,
			Height: args.Height,
		}, nil
	} else {
		return nil, err
	}
}

// Return all currently open tabs.
func (self *Commands) Tabs() ([]browser.Tab, error) {
	return nil, fmt.Errorf(`NI`)
}

// Navigate back through the current tab's history.
func (self *Commands) Back() error {
	if r, err := self.browser.Tab().RPC(`Page`, `getNavigationHistory`, nil); err == nil {
		results := maputil.M(r)
		current := results.Int(`currentIndex`)
		entries := results.Slice(`entries`)

		log.Debugf("current: %d", current)
		log.Dumpf("entries: %s", entries)

		return nil
	} else {
		return err
	}
}
