package core

import (
	"math/rand"
	"time"

	defaults "github.com/ghetzel/go-defaults"
	"github.com/ghetzel/go-stockutil/typeutil"
	"github.com/ghetzel/go-webfriend/utils"
)

type TypeArgs struct {
	Alt     bool `json:"alt"`
	Control bool `json:"control"`
	Shift   bool `json:"shift"`
	Meta    bool `json:"meta"`

	// Whether the text being input is issued via the numeric keypad or not.
	IsKeypad bool `json:"is_keypad"`

	// How long that each individual keystroke will remain down for.
	KeyDownTime time.Duration `json:"key_down_time" default:"30ms"`

	// An amount of time to randomly vary the key_down_time duration from within each keystroke.
	KeyDownJitter time.Duration `json:"key_down_jitter"`

	// How long to wait between issuing individual keystrokes.
	Delay time.Duration `json:"delay" default:"30ms"`

	// An amount of time to randomly vary the delay duration from between keystrokes.
	DelayJitter time.Duration `json:"delay_jitter"`
}

// Input the given textual data as keyboard input into the currently focused
// page element.
func (self *Commands) Type(input interface{}, args *TypeArgs) (string, error) {
	if args == nil {
		args = &TypeArgs{}
	}

	text := typeutil.V(input).String()
	defaults.SetDefaults(args)

	args.KeyDownTime = utils.FudgeDuration(args.KeyDownTime)
	args.KeyDownJitter = utils.FudgeDuration(args.KeyDownJitter)
	args.Delay = utils.FudgeDuration(args.Delay)
	args.DelayJitter = utils.FudgeDuration(args.DelayJitter)

	for _, char := range text {
		// send the keyDown event
		if _, err := self.browser.Tab().RPC(`Input`, `dispatchKeyEvent`, map[string]interface{}{
			`type`:     `keyDown`,
			`text`:     string(char),
			`isKeypad`: args.IsKeypad,
		}); err == nil {
			if args.KeyDownTime > 0 {
				time.Sleep(
					args.KeyDownTime + (time.Duration(float64(args.KeyDownJitter)*rand.Float64()) * time.Millisecond),
				)
			}
		} else {
			return ``, err
		}

		// send the keyUp event
		if _, err := self.browser.Tab().RPC(`Input`, `dispatchKeyEvent`, map[string]interface{}{
			`type`: `keyUp`,
		}); err != nil {
			return ``, err
		}

		// simulate the time between key presses
		if args.Delay > 0 {
			time.Sleep(
				args.Delay + (time.Duration(float64(args.DelayJitter)*rand.Float64()) * time.Millisecond),
			)
		}
	}

	return text, nil
}
