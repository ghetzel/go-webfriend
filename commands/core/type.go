package core

import (
	"math/rand/v2"
	"regexp"
	"time"

	defaults "go.gary.cool/go-defaults"
	"go.gary.cool/go-stockutil/rxutil"
	"go.gary.cool/go-stockutil/stringutil"
	"go.gary.cool/go-stockutil/typeutil"
	"go.gary.cool/go-webfriend/browser"
	"go.gary.cool/go-webfriend/utils"
)

var rxKeyCodes = regexp.MustCompile(`(\[[^\]]*?\]|.)`)

type TypeArgs struct {
	// How long that each individual keystroke will remain down for.
	KeyDownTime time.Duration `json:"key_down_time" default:"30ms"`

	// An amount of time to randomly vary the `key_down_time` duration from within each keystroke.
	KeyDownJitter time.Duration `json:"key_down_jitter"`

	// How long to wait between issuing individual keystrokes.
	Delay time.Duration `json:"delay" default:"30ms"`

	// An amount of time to randomly vary the delay duration from between keystrokes.
	DelayJitter time.Duration `json:"delay_jitter"`
}

// Input the given textual data as keyboard input into the currently focused
// page element.  The input text contains raw unicode characters that will be typed
// literally, as well as key names (in accordance with the DOM pre-defined keynames
// described at https://developer.mozilla.org/en-US/docs/Web/API/KeyboardEvent/key/Key_Values).
// These sequences appear between square brackets "[" "]".
//
// Example: Type in the Konami code
//
//	type "[ArrowUp][ArrowUp][ArrowDown][ArrowDown][ArrowLeft][ArrowRight][ArrowLeft][ArrowRight]BA"
func (self *Commands) Type(input any, args *TypeArgs) (string, error) {
	if typeutil.IsEmpty(input) {
		return ``, nil
	}

	if args == nil {
		args = new(TypeArgs)
	}

	var symbols = rxutil.Match(rxKeyCodes, typeutil.String(input)).AllCaptures()
	var text string

	defaults.SetDefaults(args)

	args.KeyDownTime = utils.FudgeDuration(args.KeyDownTime)
	args.KeyDownJitter = utils.FudgeDuration(args.KeyDownJitter)
	args.Delay = utils.FudgeDuration(args.Delay)
	args.DelayJitter = utils.FudgeDuration(args.DelayJitter)

	if pg := self.browser.Page(); pg != nil {
		for _, symbol := range symbols {
			if stringutil.IsSurroundedBy(symbol, `[`, `]`) {
				symbol = stringutil.Unwrap(symbol, `[`, `]`)
			} else {
				text += string(symbol)
			}

			// simulate the time between key presses
			if args.Delay > 0 {
				time.Sleep(
					args.Delay + (time.Duration(float64(args.DelayJitter)*rand.Float64()) * time.Millisecond),
				)
			}

			var keyboard = pg.Keyboard()

			// send the keyDown event
			if err := keyboard.Down(symbol); err == nil {
				if args.KeyDownTime > 0 {
					time.Sleep(
						args.KeyDownTime + (time.Duration(float64(args.KeyDownJitter)*rand.Float64()) * time.Millisecond),
					)
				}
			} else {
				return ``, err
			}

			// send the keyUp event
			if err := keyboard.Up(symbol); err != nil {
				return ``, err
			}
		}
	} else {
		return ``, browser.NoActivePage
	}

	return text, nil
}
