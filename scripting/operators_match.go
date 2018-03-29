package scripting

import (
	"fmt"
	"regexp"
)

type MatchOperator int

const (
	matchOp MatchOperator = iota
	unmatchOp
)

func parseMatchComparator(node *node32) (MatchOperator, error) {
	if node == nil {
		return -1, fmt.Errorf("nil node provided")
	}

	if node.rule() == ruleMatchOperator {
		node = node.firstChild()
	}

	switch node.rule() {
	case ruleMatch:
		return matchOp, nil
	case ruleUnmatch:
		return unmatchOp, nil
	default:
		return -1, fmt.Errorf("invalid operator %q", node)
	}
}

func (self MatchOperator) Evaluate(pattern *regexp.Regexp, want interface{}) bool {
	if v, err := exprToValue(want); err == nil {
		want = v
	} else {
		panic(fmt.Errorf("malformed expression: %v", err))
	}

	log.Debugf("RXMO(%v) %v match %v -> %v", self, pattern, want, pattern.MatchString(fmt.Sprintf("%v", want)))

	switch self {
	case matchOp:
		return pattern.MatchString(fmt.Sprintf("%v", want))
	default:
		return !pattern.MatchString(fmt.Sprintf("%v", want))
	}
}
