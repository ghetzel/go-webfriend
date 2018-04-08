package scripting

import (
	"fmt"
	"strings"

	"github.com/ghetzel/go-stockutil/log"
	"github.com/ghetzel/go-stockutil/maputil"
	"github.com/ghetzel/go-stockutil/sliceutil"
	"github.com/ghetzel/go-stockutil/stringutil"
	"github.com/ghetzel/go-stockutil/typeutil"
)

type Comparator int

const (
	cmpEquality Comparator = iota
	cmpNonEquality
	cmpGreaterThan
	cmpGreaterEqual
	cmpLessEqual
	cmpLessThan
	cmpMembership
	cmpNonMembership
)

func parseComparator(node *node32) (Comparator, error) {
	if node == nil {
		return -1, fmt.Errorf("nil node provided")
	}

	if node.rule() == ruleComparisonOperator {
		node = node.firstChild()
	}

	switch node.rule() {
	case ruleEquality:
		return cmpEquality, nil
	case ruleNonEquality:
		return cmpNonEquality, nil
	case ruleGreaterThan:
		return cmpGreaterThan, nil
	case ruleGreaterEqual:
		return cmpGreaterEqual, nil
	case ruleLessEqual:
		return cmpLessEqual, nil
	case ruleLessThan:
		return cmpLessThan, nil
	case ruleMembership:
		return cmpMembership, nil
	case ruleNonMembership:
		return cmpNonMembership, nil
	default:
		return -1, fmt.Errorf("invalid operator %q", node)
	}
}

func (self Comparator) Evaluate(lhs *Expression, rhs *Expression) bool {
	var lvv, rvv interface{}
	var lv, rv float64
	var lverr error
	var rverr error

	if lhs == nil {
		log.Fatal("malformed expression: missing left-hand side")
	} else if v, err := lhs.Value(); err == nil {
		lvv = v
	} else {
		log.Panicf("invalid expression result: %v", err)
	}

	if rhs == nil {
		return isTruthy(lhs)
	} else if v, err := rhs.Value(); err == nil {
		rvv = v
	} else {
		log.Panicf("invalid expression result: %v", err)
	}

	lv, lverr = stringutil.ConvertToFloat(lvv)
	rv, rverr = stringutil.ConvertToFloat(rvv)

	switch self {
	case cmpEquality:
		if res, err := stringutil.RelaxedEqual(lvv, rvv); err == nil {
			return res
		} else {
			log.Panicf("incomparable types %T, %T: %v", lvv, rvv, err)
			return false
		}
	case cmpNonEquality:
		if res, err := stringutil.RelaxedEqual(lvv, rvv); err == nil {
			return !res
		} else {
			log.Panicf("incomparable types %T, %T: %v", lvv, rvv, err)
			return false
		}

	case cmpGreaterThan:
		if lverr == nil && rverr == nil {
			return (lv > rv)
		} else {
			log.Panicf("incomparable types %T, %T", lvv, rvv)
			return false
		}

	case cmpGreaterEqual:
		if lverr == nil && rverr == nil {
			return (lv >= rv)
		} else {
			log.Panicf("incomparable types %T, %T", lvv, rvv)
			return false
		}

	case cmpLessEqual:
		if lverr == nil && rverr == nil {
			return (lv <= rv)
		} else {
			log.Panicf("incomparable types %T, %T", lvv, rvv)
			return false
		}

	case cmpLessThan:
		if lverr == nil && rverr == nil {
			return (lv < rv)
		} else {
			log.Panicf("incomparable types %T, %T", lvv, rvv)
			return false
		}

	case cmpMembership:
		return isMemberOf(lvv, rvv)

	case cmpNonMembership:
		return !isMemberOf(lvv, rvv)

	default:
		return false
	}
}

func membershipTest(i int, first interface{}, second interface{}) bool {
	return fmt.Sprintf("%v", first) == fmt.Sprintf("%v", second)
}

func isMemberOf(lhs interface{}, rhs interface{}) bool {
	if typeutil.IsArray(rhs) {
		log.Debugf("is %v in %v", lhs, rhs)
		return sliceutil.Contains(rhs, lhs, membershipTest)
	} else if typeutil.IsMap(rhs) {
		return sliceutil.Contains(maputil.Keys(rhs), lhs, membershipTest)
	} else {
		return strings.Contains(fmt.Sprintf("%v", rhs), fmt.Sprintf("%v", lhs))
	}
}
