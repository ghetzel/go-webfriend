package core

import (
	defaults "github.com/ghetzel/go-defaults"
	"github.com/ghetzel/go-webfriend/browser"
)

type MouseArgs struct {
	// The X-coordinate to perform the mouse action at.
	X float64 `json:"x"`

	// The Y-coordinate to perform the mouse action at.
	Y float64 `json:"y"`

	// The action that should be performed; one of "move", "press", "release", or "scroll".
	Action string `json:"action" default:"move"`

	// Which mouse button to depress when performing the action; one of "left", "middle", or "right".
	Button string `json:"button,omitempty"`

	// Whether the Alt-key should be held down when emitting the event.
	Alt bool `json:"alt,omitempty"`

	// Whether the Control-key should be held down when emitting the event.
	Control bool `json:"control,omitempty"`

	// Whether the Meta/Command-key should be held down when emitting the event.
	Meta bool `json:"meta,omitempty"`

	// Whether the Shift-key should be held down when emitting the event.
	Shift bool `json:"shift,omitempty"`

	// For "scroll" actions, this indicates how much to scroll horizontally (positive for right, negative for left).
	WheelX float64 `json:"wheelX,omitempty"`

	// For "scroll" actions, this indicates how much to scroll vertically (positive for up, negative for down)
	WheelY float64 `json:"wheelY,omitempty"`

	// How many clicks to issue if action is "press"
	Count int `json:"count,omitempty"`
}

func (self *Commands) Mouse(args *MouseArgs) error {
	if args == nil {
		args = &MouseArgs{}
	}

	defaults.SetDefaults(args)

	var action browser.MouseAction

	switch args.Action {
	case `press`:
		action = browser.Pressed
	case `release`:
		action = browser.Released
	case `scroll`:
		action = browser.Scrolled
	default:
		action = browser.Moved
	}

	// log.Noticef("M: %v %v %v (%v,%v)", args.Action, args.X, args.Y, args.WheelX, args.WheelY)

	return self.browser.Tab().MoveMouse(args.X, args.Y, &browser.MouseActionConfig{
		Action:  action,
		Button:  browser.Button(args.Button),
		Alt:     args.Alt,
		Control: args.Control,
		Meta:    args.Meta,
		Shift:   args.Shift,
		WheelX:  args.WheelX,
		WheelY:  args.WheelY,
		Count:   args.Count,
	})
}
