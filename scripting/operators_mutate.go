package scripting

import (
	"fmt"
	"math"

	"github.com/ghetzel/go-stockutil/stringutil"
)

type operator int

const (
	opNull operator = iota
	opExponentiate
	opMultiply
	opDivide
	opModulus
	opAdd
	opSubtract
	opBitwiseAnd
	opBitwiseOr
	opBitwiseNot
	opBitwiseXor
)

func parseOperator(node *node32) (operator, error) {
	if node == nil {
		return opNull, fmt.Errorf("nil node provided")
	}

	switch node.first().rule() {
	case ruleExponentiate:
		return opExponentiate, nil
	case ruleMultiply:
		return opMultiply, nil
	case ruleDivide:
		return opDivide, nil
	case ruleModulus:
		return opModulus, nil
	case ruleAdd:
		return opAdd, nil
	case ruleSubtract:
		return opSubtract, nil
	case ruleBitwiseAnd:
		return opBitwiseAnd, nil
	case ruleBitwiseOr:
		return opBitwiseOr, nil
	case ruleBitwiseNot:
		return opBitwiseNot, nil
	case ruleBitwiseXor:
		return opBitwiseXor, nil
	default:
		return -1, fmt.Errorf("invalid operator %q", node)
	}
}

func (self operator) evaluate(lhs interface{}, rhs interface{}) (interface{}, error) {
	var lv float64
	var rv float64
	var lverr error
	var rverr error

	if v, err := exprToValue(lhs); err == nil {
		lhs = v
	} else {
		return nil, err
	}

	if v, err := exprToValue(rhs); err == nil {
		rhs = v
	} else {
		return nil, err
	}

	lv, lverr = stringutil.ConvertToFloat(lhs)
	rv, rverr = stringutil.ConvertToFloat(rhs)

	// for operators that can work with non-numeric values, we can proceed; otherwise
	// it is an error to have not been able to convert lhs/rhs into floats
	if self != opAdd {
		if lverr != nil {
			return nil, lverr
		}

		if rverr != nil {
			return nil, rverr
		}
	}

	var output interface{}
	var err error

	switch self {
	case opExponentiate:
		output = math.Pow(lv, rv)
	case opMultiply:
		output = (lv * rv)
	case opDivide:
		if rv == 0 {
			err = fmt.Errorf("cannot divide by zero")
		} else {
			output = (lv / rv)
		}

	case opModulus:
		output = math.Mod(lv, rv)

	case opAdd:
		_, lstr := lhs.(string)
		_, rstr := rhs.(string)

		if lstr || rstr {
			output = fmt.Sprintf("%v%v", lhs, rhs)
		} else {
			output = (lv + rv)
		}

	case opSubtract:
		output = (lv - rv)

	case opBitwiseAnd:
		output = int64(lv) & int64(rv)

	case opBitwiseOr:
		output = int64(lv) | int64(rv)

	case opBitwiseXor:
		output = int64(lv) ^ int64(rv)

	default:
		err = fmt.Errorf("operator '%v' not implemented", self)
	}

	if output == nil {
		output = new(emptyValue)
	}

	if err == nil {
		// log.Debugf("EXPR %v %v %v = %v", lv, self, rv, output)
		return output, nil
	} else {
		// log.Debugf("EXPR %v %v %v !ERR! %v", lv, self, rv, err)
		return nil, err
	}
}

func (self operator) String() string {
	switch self {
	case opExponentiate:
		return `**`
	case opMultiply:
		return `*`
	case opDivide:
		return `/`
	case opModulus:
		return `%`
	case opAdd:
		return `+`
	case opSubtract:
		return `-`
	case opBitwiseAnd:
		return `&`
	case opBitwiseOr:
		return `|`
	case opBitwiseNot:
		return `~`
	case opBitwiseXor:
		return `^`
	default:
		return `INVALID`
	}
}
