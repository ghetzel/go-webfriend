package core

import (
	"math/rand"
	"regexp"
	"time"

	defaults "github.com/ghetzel/go-defaults"
	"github.com/ghetzel/go-stockutil/rxutil"
	"github.com/ghetzel/go-stockutil/stringutil"
	"github.com/ghetzel/go-stockutil/typeutil"
	"github.com/ghetzel/go-webfriend/utils"
)

var rxKeyCodes = regexp.MustCompile(`(\[[^\]]*?\]|.)`)

type TypeArgs struct {
	Alt     bool `json:"alt"`
	Control bool `json:"control"`
	Shift   bool `json:"shift"`
	Meta    bool `json:"meta"`

	// Whether the text being input is issued via the numeric keypad or not.
	IsKeypad bool `json:"is_keypad"`

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
//   type "[ArrowUp][ArrowUp][ArrowDown][ArrowDown][ArrowLeft][ArrowRight][ArrowLeft][ArrowRight]BA"
//
func (self *Commands) Type(input interface{}, args *TypeArgs) (string, error) {
	if typeutil.IsEmpty(input) {
		return ``, nil
	}

	if args == nil {
		args = &TypeArgs{}
	}

	var modifiers int
	var symbols = rxutil.Match(rxKeyCodes, typeutil.String(input)).AllCaptures()
	var text string

	defaults.SetDefaults(args)

	args.KeyDownTime = utils.FudgeDuration(args.KeyDownTime)
	args.KeyDownJitter = utils.FudgeDuration(args.KeyDownJitter)
	args.Delay = utils.FudgeDuration(args.Delay)
	args.DelayJitter = utils.FudgeDuration(args.DelayJitter)

	// modifiers: Bit field representing pressed modifier keys. Alt=1, Ctrl=2, Meta/Command=4, Shift=8 (default: 0)
	if args.Alt {
		modifiers |= 1
	}
	if args.Control {
		modifiers |= 2
	}
	if args.Meta {
		modifiers |= 4
	}
	if args.Shift {
		modifiers |= 8
	}

	for _, symbol := range symbols {
		var keyEvent = map[string]interface{}{
			`type`:      `keyDown`,
			`isKeypad`:  args.IsKeypad,
			`modifiers`: modifiers,
		}

		if stringutil.IsSurroundedBy(symbol, `[`, `]`) {
			keyEvent[`key`] = stringutil.Unwrap(symbol, `[`, `]`)
		} else {
			text += string(symbol)
			keyEvent[`text`] = string(symbol)
		}

		// send the keyDown event
		if _, err := self.browser.Tab().RPC(`Input`, `dispatchKeyEvent`, keyEvent); err == nil {
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
