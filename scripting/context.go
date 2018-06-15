package scripting

import (
	"fmt"
	"time"
)

type ContextType string

const (
	BlockContext     ContextType = `block`
	StatementContext             = `statement`
	CommandContext               = `command`
)

type Context struct {
	Type                ContextType
	Label               string
	Script              *Friendscript
	Parent              *Context
	AbsoluteStartOffset int
	Length              int
	Error               error
	StartedAt           time.Time
	Took                time.Duration
}

func (self *Context) String() string {
	return fmt.Sprintf("[%v] %v %d + %d", self.Type, self.Label, self.AbsoluteStartOffset, self.Length)
}
