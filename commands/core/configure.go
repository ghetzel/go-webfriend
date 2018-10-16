package core

import (
	"fmt"

	defaults "github.com/ghetzel/go-defaults"
	"github.com/ghetzel/go-stockutil/colorutil"
)

type ConfigureArgs struct {
	// Override the default User-Agent header sent with all requests
	UserAgent string `json:"user_agent"`

	// Specify the Geographic latitude to emulate [-90.0, 90.0]
	Latitude float64 `json:"latitude"`

	// Specify the Geographic longitude to emulate [-180.0, 180.0]
	Longitude float64 `json:"longitude"`

	// Specify a Geolocation error margin (accurate to within n meters)
	Accuracy float64 `json:"accuracy" default:"1"`

	// Disable JavaScript execution in the browser.
	DisableScripts bool `json:"disable_scripts"`

	// Emulate a touch-capable device.
	EmulateTouch bool `json:"emulate_touch"`

	// Set the default background color of the underlying window in the following formats: `#RRGGBB`, `#RRGGBBAA`, `rgb()`, `rgba()`, `hsv()`, `hsva()`, `hsl()`, `hsla()`.
	BackgroundColor string `json:"background_color"`

	// Set whether scrollbars should be hidden all the time.
	HideScrollbars bool `json:"hide_scrollbars"`
}

// Configures various features of the Remote Debugging protocol and provides
// environment setup.
func (self *Commands) Configure(args *ConfigureArgs) error {
	if args == nil {
		args = &ConfigureArgs{}
	}

	defaults.SetDefaults(args)

	// Geolocation
	// ---------------------------------------------------------------------------------------------
	lat := args.Latitude
	lon := args.Longitude

	if lat != 0 && lon != 0 {
		if _, err := self.browser.Tab().RPC(`Emulation`, `setGeolocationOverride`, map[string]interface{}{
			`latitude`:  lat,
			`longitude`: lon,
			`accuracy`:  args.Accuracy,
		}); err != nil {
			return err
		}
	} else {
		self.browser.Tab().RPC(`Emulation`, `clearGeolocationOverride`, nil)
	}

	// User Agent
	// ---------------------------------------------------------------------------------------------
	if ua := args.UserAgent; ua != `` {
		if _, err := self.browser.Tab().RPC(`Emulation`, `setUserAgentOverride`, map[string]interface{}{
			`userAgent`: ua,
		}); err != nil {
			return err
		}
	}

	// Emulation Flags & Features
	// ---------------------------------------------------------------------------------------------
	self.browser.Tab().AsyncRPC(`Emulation`, `setScriptExecutionDisabled`, map[string]interface{}{
		`value`: args.DisableScripts,
	})

	self.browser.Tab().AsyncRPC(`Emulation`, `setTouchEmulationEnabled`, map[string]interface{}{
		`value`: args.EmulateTouch,
	})

	self.browser.Tab().AsyncRPC(`Emulation`, `setScrollbarsHidden`, map[string]interface{}{
		`hidden`: args.HideScrollbars,
	})

	if bgcolor := args.BackgroundColor; bgcolor != `` {
		if col, err := colorutil.Parse(bgcolor); err == nil {
			r, g, b, a := col.RGBA255()

			self.browser.Tab().AsyncRPC(`Emulation`, `setDefaultBackgroundColorOverride`, map[string]interface{}{
				`r`: r,
				`g`: g,
				`b`: b,
				`a`: a,
			})
		} else {
			return fmt.Errorf("invalid background color: %v", err)
		}
	} else {
		self.browser.Tab().AsyncRPC(`Emulation`, `setDefaultBackgroundColorOverride`, nil)
	}

	return nil
}
