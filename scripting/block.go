package scripting

import (
	"fmt"
	"strings"

	"github.com/ghetzel/go-stockutil/stringutil"
)

type Block struct {
	friendscript *Friendscript
	node         *node32
	parent       *Statement
}

type BlockType int

const (
	UnknownBlock BlockType = iota
	StatementBlock
	EventHandlerBlock
	FlowControlWord
)

func (self BlockType) String() string {
	switch self {
	case StatementBlock:
		return `StatementBlock`
	case EventHandlerBlock:
		return `EventHandlerBlock`
	case FlowControlWord:
		return `FlowControlWord`
	default:
		return `UnknownBlock`
	}
}

func (self *Block) Script() *Friendscript {
	return self.friendscript
}

func (self *Block) String() string {
	return fmt.Sprintf("%v (%d children)", self.Type(), len(self.node.children()))
}

func (self *Block) Type() BlockType {
	switch self.node.rule() {
	case ruleStatementBlock:
		return StatementBlock
	// case ruleEventHandlerBlock:
	case ruleFlowControlWord:
		return FlowControlWord
	default:
		return UnknownBlock
	}
}

func (self *Block) FlowBreak() int {
	return self.flowControl(ruleFlowControlBreak)
}

func (self *Block) FlowContinue() int {
	return self.flowControl(ruleFlowControlContinue)
}

func (self *Block) flowControl(rule pegRule) int {
	if self.Type() == FlowControlWord {
		if n := self.node.firstChild(rule); n != nil {
			if levels := n.firstChild(rulePositiveInteger); levels != nil {
				l := int(stringutil.MustInteger(
					strings.TrimSpace(self.Script().s(levels)),
				))

				if l < 1 {
					return 1
				} else {
					return l
				}
			} else {
				return 1
			}
		}
	}

	return 0
}

func (self *Block) Statements() []*Statement {
	statements := make([]*Statement, 0)

	for _, node := range self.node.children() {
		statement := &Statement{
			node:  node,
			block: self,
		}

		log.Debugf("STMT %v", strings.TrimSpace(statement.raw(node)))
		statements = append(statements, statement)
	}

	return statements
}
