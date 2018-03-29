package scripting

import (
	"fmt"

	"github.com/ghetzel/go-stockutil/sliceutil"
	"github.com/ghetzel/go-stockutil/stringutil"
)

type AssignmentOperator int

const (
	assnAssignEq AssignmentOperator = iota
	assnStarEq
	assnDivEq
	assnPlusEq
	assnMinusEq
	assnAndEq
	assnOrEq
	assnAppend
)

func (self AssignmentOperator) String() string {
	switch self {
	case assnStarEq:
		return `*=`
	case assnDivEq:
		return `/=`
	case assnPlusEq:
		return `+=`
	case assnMinusEq:
		return `-=`
	case assnAndEq:
		return `&=`
	case assnOrEq:
		return `|=`
	case assnAppend:
		return `<<`
	default:
		return `=`
	}
}

func parseAssignmentOperator(node *node32) (AssignmentOperator, error) {
	if node == nil {
		return -1, fmt.Errorf("nil node provided")
	}

	if node.rule() == ruleAssignmentOperator {
		node = node.firstChild()
	}

	switch node.rule() {
	case ruleAssignEq:
		return assnAssignEq, nil
	case ruleStarEq:
		return assnStarEq, nil
	case ruleDivEq:
		return assnDivEq, nil
	case rulePlusEq:
		return assnPlusEq, nil
	case ruleMinusEq:
		return assnMinusEq, nil
	case ruleAndEq:
		return assnAndEq, nil
	case ruleOrEq:
		return assnOrEq, nil
	case ruleAppend:
		return assnAppend, nil
	default:
		return -1, fmt.Errorf("invalid operator %q", node)
	}
}

func (self AssignmentOperator) ShouldPreclear() bool {
	switch self {
	case assnAssignEq:
		return true
	default:
		return false
	}
}

func (self AssignmentOperator) Evaluate(lhs interface{}, rhs interface{}) (interface{}, error) {
	if v, err := exprToValue(lhs); err == nil {
		lhs = v
	} else {
		panic(fmt.Errorf("malformed expression: %v", err))
	}

	if v, err := exprToValue(rhs); err == nil {
		rhs = v
	} else {
		panic(fmt.Errorf("malformed expression: %v", err))
	}

	var lv float64
	var rv float64
	var lverr error
	var rverr error

	log.Debugf("lhs=%T(%v) %v rhs=%T(%v)", lhs, lhs, self, rhs, rhs)

	lv, lverr = stringutil.ConvertToFloat(lhs)
	rv, rverr = stringutil.ConvertToFloat(rhs)

	switch self {
	case assnAssignEq, assnAppend:
		break
	default:
		if lverr != nil {
			lv = 0
		}

		if rverr != nil {
			return nil, rverr
		}
	}

	switch self {
	case assnAssignEq:
		return rhs, nil

	case assnStarEq:
		return (lv * rv), nil

	case assnDivEq:
		if rv == 0 {
			return 0, fmt.Errorf("divide by zero")
		}

		return (lv / rv), nil

	case assnPlusEq:
		return (lv + rv), nil

	case assnMinusEq:
		return (lv - rv), nil
	case assnAndEq:
		return int64(lv) & int64(rv), nil

	case assnOrEq:
		return int64(lv) | int64(rv), nil
	case assnAppend:
		return append(sliceutil.Sliceify(lhs), rhs), nil
	}

	return 0, fmt.Errorf("unsupported assignment operator %v", self)
}
