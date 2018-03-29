package core

import (
	"fmt"
	"time"

	"github.com/ghetzel/go-stockutil/stringutil"
	"github.com/ghetzel/go-stockutil/timeutil"
)

// Pauses execution of the current script for the given duration.
func (self *Commands) Wait(delay interface{}) error {
	var duration time.Duration

	if delayD, ok := delay.(time.Duration); ok {
		duration = delayD
	} else if delayMs, err := stringutil.ConvertToInteger(delay); err == nil {
		duration = time.Duration(delayMs) * time.Millisecond
	} else if delayParsed, err := timeutil.ParseDuration(fmt.Sprintf("%v", delay)); err == nil {
		duration = delayParsed
	} else {
		return fmt.Errorf("invalid duration: %v", err)
	}

	time.Sleep(duration)
	return nil
}
