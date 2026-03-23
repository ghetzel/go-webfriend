package core

import (
	"time"

	defaults "github.com/ghetzel/go-defaults"
	"github.com/ghetzel/go-webfriend/browser"
	"github.com/playwright-community/playwright-go"
)

type MouseArgs struct {
	// The X-coordinate to perform the mouse action at.
	X float64 `json:"x"`

	// The Y-coordinate to perform the mouse action at.
	Y float64 `json:"y"`

	// The action that should be performed; one of "move", "click", "press", "release", or "scroll".
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

	// How many clicks to issue if an action involves a button press, and how many intermediate points to use when moving the mouse.
	Count int `json:"count,omitempty" default:"1"`

	// Time to waith between mousedown and mouseup events.
	Delay time.Duration `json:"delay" default:"20ms"`
}

func (self *Commands) Mouse(args *MouseArgs) error {
	if args == nil {
		args = &MouseArgs{}
	}

	defaults.SetDefaults(args)

	if pg := self.browser.Page(); pg != nil {
		var mouse = pg.Mouse()
		var btn *playwright.MouseButton

		switch args.Button {
		case `middle`:
			btn = playwright.MouseButtonMiddle
		case `right`:
			btn = playwright.MouseButtonRight
		default:
			btn = playwright.MouseButtonLeft
		}

		switch args.Action {
		case `press`:
			return mouse.Down(playwright.MouseDownOptions{
				Button:     btn,
				ClickCount: playwright.Int(args.Count),
			})
		case `release`:
			return mouse.Up(playwright.MouseUpOptions{
				Button:     btn,
				ClickCount: playwright.Int(args.Count),
			})
		case `click`:
			return mouse.Click(
				float64(args.X),
				float64(args.Y),
				playwright.MouseClickOptions{
					Button:     btn,
					ClickCount: playwright.Int(args.Count),
					Delay:      playwright.Float(float64(args.Delay.Milliseconds())),
				},
			)
		case `scroll`:
			return mouse.Wheel(args.WheelX, args.WheelY)
		default:
			return mouse.Move(
				float64(args.X),
				float64(args.Y),
				playwright.MouseMoveOptions{
					Steps: playwright.Int(args.Count),
				},
			)
		}
	} else {
		return browser.NoActivePage
	}
}
