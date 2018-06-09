package core

import (
	"github.com/ghetzel/go-webfriend/browser"
	defaults "github.com/mcuadros/go-defaults"
)

type KeyArgs struct {
	// The keyboard action to take; either "press" or "release"
	Action  string `json:"action" default:"press"`
	Alt     bool   `json:"alt,omitempty"`
	Control bool   `json:"control,omitempty"`
	Meta    bool   `json:"meta,omitempty"`
	Shift   bool   `json:"shift,omitempty"`

	// The numeric decimal keycode to send.
	KeyCode int `json:"keycode,omitempty"`
}

func (self *Commands) Key(domKeyName string, args *KeyArgs) error {
	if args == nil {
		args = &KeyArgs{}
	}

	defaults.SetDefaults(args)

	var action browser.KeyboardAction

	switch args.Action {
	case `press`:
		action = browser.KeyPressed
	default:
		action = browser.KeyReleased
	}

	return self.browser.Tab().SendKey(domKeyName, &browser.KeyboardActionConfig{
		Action:  action,
		Alt:     args.Alt,
		Control: args.Control,
		Meta:    args.Meta,
		Shift:   args.Shift,
		KeyCode: args.KeyCode,
	})
}
