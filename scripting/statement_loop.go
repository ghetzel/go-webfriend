package scripting

import (
	"fmt"

	"github.com/ghetzel/go-stockutil/stringutil"
)

type LoopType int

const (
	InfiniteLoop LoopType = iota
	FixedLengthLoop
	IteratorLoop
	ConditionBoundedLoop
	WhileLoop
)

type FlowControlType int

const (
	FlowBreak FlowControlType = iota
	FlowContinue
)

type FlowControlErr struct {
	Type  FlowControlType
	Level int
}

func NewFlowControl(flowType FlowControlType, levels int) *FlowControlErr {
	return &FlowControlErr{
		Type:  flowType,
		Level: levels,
	}
}

func (self FlowControlErr) Error() string {
	msg := ``

	switch self.Type {
	case FlowBreak:
		msg = `break`
	case FlowContinue:
		msg = `continue`
	default:
		self.Level = 0
		return `invalid flow control statement`
	}

	return fmt.Sprintf("%v %d", msg, self.Level)
}

func (self LoopType) String() string {
	switch self {
	case InfiniteLoop:
		return `InfiniteLoop`
	case FixedLengthLoop:
		return `FixedLengthLoop`
	case IteratorLoop:
		return `IteratorLoop`
	case ConditionBoundedLoop:
		return `ConditionBoundedLoop`
	case WhileLoop:
		return `WhileLoop`
	default:
		return ``
	}
}

type Loop struct {
	statement  *Statement
	iterations int
}

func (self *Loop) String() string {
	loopType := self.Type()

	switch loopType {
	case FixedLengthLoop:
		return fmt.Sprintf("%v (%d iterations)", loopType, self.UpperBound())
	default:
		return fmt.Sprintf("%v", loopType)
	}
}

func (self *Loop) Type() LoopType {
	subnode := self.statement.node.first(
		ruleLoopConditionFixedLength,
		ruleLoopConditionIterable,
		ruleLoopConditionBounded,
		ruleLoopConditionTruthy,
	)

	if subnode != nil {
		switch subnode.rule() {
		case ruleLoopConditionFixedLength:
			return FixedLengthLoop
		case ruleLoopConditionIterable:
			return IteratorLoop
		case ruleLoopConditionBounded:
			return ConditionBoundedLoop
		case ruleLoopConditionTruthy:
			return WhileLoop
		}
	}

	return InfiniteLoop
}

func (self *Loop) UpperBound() int {
	if self.Type() == FixedLengthLoop {
		if lenNode := self.statement.node.firstChild(ruleLoopConditionFixedLength); lenNode != nil {
			var nI interface{}

			if arg := lenNode.first(ruleInteger); arg != nil {
				nI = self.statement.raw(arg)
			} else if arg := lenNode.first(ruleVariable); arg != nil {
				varname := self.statement.raw(arg)

				if v, err := self.statement.resolveVariable(arg); err == nil {
					nI = v
				} else {
					log.Fatalf("error resolving variable '%v': %v", varname, err)
				}
			} else {
				log.Fatalf(
					"invalid loop syntax: '%v'",
					self.statement.raw(lenNode),
				)
			}

			if nI != nil {
				if n, err := stringutil.ConvertToInteger(nI); err == nil {
					return int(n)
				} else {
					log.Fatalf("invalid loop argument: '%v'", nI)
				}
			} else {
				log.Fatal("missing loop argument")
				return -1
			}
		}
	}

	return -1
}

func (self *Loop) CurrentIndex() int {
	return self.iterations
}

func (self *Loop) Reset() {
	self.iterations = -1
}

func (self *Loop) ShouldContinue() bool {
	self.iterations += 1

	switch self.Type() {
	case InfiniteLoop:
		return true

	case FixedLengthLoop:
		if self.iterations < self.UpperBound() {
			log.Debugf("LOOP iteration %d < %d\n", self.iterations, self.UpperBound())
			return true
		}

	case IteratorLoop:
		return true

	default:
		log.Fatalf("NI type=%v", self.Type())
	}

	return false
}

func (self *Loop) Blocks() []*Block {
	blocks := make([]*Block, 0)

	for _, node := range self.statement.node.children(ruleBlock) {
		blocks = append(blocks, &Block{
			friendscript: self.statement.block.friendscript,
			node:         node.first(),
			parent:       self.statement,
		})
	}

	return blocks
}

func (self *Loop) IteratableParts() ([]string, interface{}) {
	if self.Type() == IteratorLoop {
		if node := self.statement.node.firstChild(ruleLoopConditionIterable); node != nil {
			lhs := node.first(ruleLoopIterableLHS)
			rhs := node.first(ruleLoopIterableRHS)

			if lhs != nil && rhs != nil {
				names := make([]string, 0)
				var rightHand interface{}

				for _, varNode := range lhs.first().children(ruleVariable) {
					if key, err := self.statement.resolveVariableKey(varNode); err == nil {
						names = append(names, key)
					} else {
						log.Fatalf("unable to resolve variable name: %v", err)
					}
				}

				if rhsNode := rhs.first(0, ruleCommand, ruleVariable); rhsNode != nil {
					if rhsNode.rule() == ruleCommand {
						rightHand = &Command{
							statement: self.statement,
							node:      rhsNode,
						}
					} else if key, err := self.statement.resolveVariableKey(rhsNode); err == nil {
						rightHand = key
					} else {
						log.Fatalf("unable to resolve variable name: %v", err)
					}
				}

				return names, rightHand
			}
		}
	}

	log.Fatal("cannot build iterable expression from given node")
	return nil, nil
}
