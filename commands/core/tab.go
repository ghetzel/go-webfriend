package core

import (
	defaults "github.com/ghetzel/go-defaults"
	"github.com/ghetzel/go-webfriend/browser"
)

type ResizeArgs struct {
	// The width of the screen.
	Width int `json:"width"`

	// The height of the screen.
	Height int `json:"height"`
}

type ResizeResponse struct {
	// The final width of the page after resize.
	Width int `json:"width"`

	// The final height of the page after resize.
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
		args = new(ResizeArgs)
	}

	defaults.SetDefaults(args)

	if pg := self.browser.Page(); pg != nil {
		if err := pg.SetViewportSize(args.Width, args.Height); err == nil {
			return &ResizeResponse{
				Width:  args.Width,
				Height: args.Height,
			}, nil
		} else {
			return nil, err
		}
	} else {
		return nil, browser.NoActivePage
	}
}

// Navigate back in the current tab's history.
func (self *Commands) Back() error {
	if pg := self.browser.Page(); pg != nil {
		var _, err = pg.GoBack()
		return err
	} else {
		return browser.NoActivePage
	}
}

// Navigate forward in the current tab's history.
func (self *Commands) Forward() error {
	if pg := self.browser.Page(); pg != nil {
		var _, err = pg.GoForward()
		return err
	} else {
		return browser.NoActivePage
	}
}

// Reload the current tab.
func (self *Commands) Reload() error {
	if pg := self.browser.Page(); pg != nil {
		var _, err = pg.Reload()
		return err
	} else {
		return browser.NoActivePage
	}
}
