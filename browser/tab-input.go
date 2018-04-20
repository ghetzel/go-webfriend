package browser

import (
	defaults "github.com/mcuadros/go-defaults"
)

type Button string

const (
	Left   Button = `left`
	Middle        = `middle`
	Right         = `right`
)

func (self Button) String() string {
	return string(self)
}

type MouseAction string

const (
	Pressed  MouseAction = `mousePressed`
	Released             = `mouseReleased`
	Moved                = `mouseMoved`
	Scrolled             = `mouseWheel`
)

func (self MouseAction) String() string {
	return string(self)
}

type MouseActionConfig struct {
	Action  MouseAction `json:"action" default:"mouseMoved"`
	Button  Button      `json:"button,omitempty"`
	Alt     bool        `json:"alt,omitempty"`
	Control bool        `json:"control,omitempty"`
	Meta    bool        `json:"meta,omitempty"`
	Shift   bool        `json:"shift,omitempty"`
	WheelX  float64     `json:"wheelX,omitempty"`
	WheelY  float64     `json:"wheelY,omitempty"`
	Count   int         `json:"count,omitempty"`
}

type KeyboardAction string

const (
	KeyPressed  KeyboardAction = `keyDown`
	KeyReleased                = `keyUp`
	KeyRaw                     = `rawKeyDown`
)

func (self KeyboardAction) String() string {
	return string(self)
}

type KeyboardActionConfig struct {
	Action  KeyboardAction `json:"action" default:"keyDown"`
	Alt     bool           `json:"alt,omitempty"`
	Control bool           `json:"control,omitempty"`
	Meta    bool           `json:"meta,omitempty"`
	Shift   bool           `json:"shift,omitempty"`
	KeyCode int            `json:"keycode,omitempty"`
}

func (self *Tab) MoveMouse(x float64, y float64, config *MouseActionConfig) error {
	if config == nil {
		config = &MouseActionConfig{}
	}

	defaults.SetDefaults(config)
	mods := 0

	if config.Alt {
		mods |= 1
	}

	if config.Control {
		mods |= 2
	}

	if config.Meta {
		mods |= 4
	}

	if config.Shift {
		mods |= 8
	}

	args := map[string]interface{}{
		`type`:       config.Action.String(),
		`x`:          x,
		`y`:          y,
		`modifiers`:  mods,
		`clickCount`: config.Count,
	}

	if v := config.Button; v != `` {
		args[`button`] = v
	}

	if config.Action == Scrolled {
		args[`deltaX`] = config.WheelX
		args[`deltaY`] = config.WheelY
	}

	return self.AsyncRPC(`Input`, `dispatchMouseEvent`, args)
}

func (self *Tab) SendKey(domKeyName string, config *KeyboardActionConfig) error {
	if config == nil {
		config = &KeyboardActionConfig{}
	}

	defaults.SetDefaults(config)
	mods := 0

	if config.Alt {
		mods |= 1
	}

	if config.Control {
		mods |= 2
	}

	if config.Meta {
		mods |= 4
	}

	if config.Shift {
		mods |= 8
	}

	args := map[string]interface{}{
		`type`:      config.Action.String(),
		`modifiers`: mods,
	}

	if len(domKeyName) == 1 {
		args[`text`] = domKeyName
	}

	args[`nativeVirtualKeyCode`] = config.KeyCode
	args[`windowsVirtualKeyCode`] = config.KeyCode

	return self.AsyncRPC(`Input`, `dispatchKeyEvent`, args)
}
