package core

import (
	defaults "go.gary.cool/go-defaults"
	"go.gary.cool/go-webfriend/browser"
)

type KeyArgs struct {
	// The keyboard action to take; either "press" or "release"
	Action  string `json:"action" default:"press"`
	Alt     bool   `json:"alt,omitempty"`
	Control bool   `json:"control,omitempty"`
	Meta    bool   `json:"meta,omitempty"`
	Shift   bool   `json:"shift,omitempty"`
}

func (self *Commands) Key(domKeyName string, args *KeyArgs) error {
	if args == nil {
		args = new(KeyArgs)
	}

	defaults.SetDefaults(args)

	if pg := self.browser.Page(); pg != nil {
		var keyboard = pg.Keyboard()

		domKeyName = normalizeKeyCodes(domKeyName, args.Meta, args.Control, args.Alt, args.Shift)

		switch args.Action {
		case `press`:
			return keyboard.Down(domKeyName)
		default:
			return keyboard.Up(domKeyName)
		}
	} else {
		return browser.NoActivePage
	}
}
