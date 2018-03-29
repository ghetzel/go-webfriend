package core

import (
	"fmt"

	"github.com/ghetzel/go-webfriend/browser"
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
	return fmt.Errorf(`NI`)
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
	// The scaling factor of the content.
	Scale float64 `json:"scale"` // 0

	// Whether to emulate a mobile device or not. If a map is provided, mobile
	// emulation will be enabled and configured using the following keys:
	//
	//    width (int, optional): The width of the mobile screen to emulate.
	//
	//    height (int, optional): The height of the mobile screen to emulate.
	//
	//    x (int, optional): The horizontal position of the currently viewable
	//                       portion of the mobile screen.
	//
	//    y (int, optional): The vertical position of the currently viewable
	//                       portion of the mobile screen.
	//
	Mobile interface{} `json:"mobile"` // false

	// Whether to fit the viewport contents to the available area or not.
	FitWindow bool `json:"fit_window"` // false

	// Which screen orientation to emulate, if any.
	Orientation Orientation `json:"orientation"` // null

	// The angle of the screen to emulate (in degrees; 0-360).
	Angle int `json:"angle"` // 0
}

// Resizes the active viewport of the current page using the Chrome Device
// Emulation API. This does not resize the window itself, but rather the area
// the current page interprets the window to be.
//
// This is useful for setting the size of the area that will be rendered for
// screenshots and screencasts, or for testing responsive design elements.
func (self *Commands) Resize(width int, height int, args *ResizeArgs) (int, int, error) {
	return -1, -1, fmt.Errorf(`NI`)
}

// Return all currently open tabs.
func (self *Commands) Tabs() ([]browser.Tab, error) {
	return nil, fmt.Errorf(`NI`)
}
