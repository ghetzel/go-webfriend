package core

import (
	"fmt"
	"time"

	"github.com/ghetzel/go-webfriend/browser"
)

type TypeArgs struct {
	Alt     bool `json:"alt"`
	Control bool `json:"control"`
	Shift   bool `json:"shift"`
	Meta    bool `json:"meta"`

	// Whether the text being input is issued via the numeric keypad or not.
	IsKeypad bool `json:"is_keypad"`

	// How long that each individual keystroke will remain down for.
	KeyDownTime time.Duration `json:"key_down_time"`

	// An amount of time to randomly vary the key_down_time duration from within each keystroke.
	KeyDownJitter time.Duration `json:"key_down_jitter"`

	// How long to wait between issuing individual keystrokes.
	Delay time.Duration `json:"delay"`

	// An amount of time to randomly vary the delay duration from between keystrokes.
	DelayJitter time.Duration `json:"delay_jitter"`
}

// Input the given textual data as keyboard input into the currently focused
// page element.
func (self *Commands) Type(input interface{}, args *TypeArgs) (*browser.Element, error) {
	return nil, fmt.Errorf(`NI`)
}
