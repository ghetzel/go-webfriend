package scripting

import (
	"github.com/ghetzel/go-stockutil/log"
	"github.com/ghetzel/go-stockutil/stringutil"
	"github.com/ghetzel/go-stockutil/typeutil"
)

type ConditionalExpression struct {
	statement *Statement
	node      *node32
}

func NewConditionalExpression(statement *Statement, node *node32) *ConditionalExpression {
	return &ConditionalExpression{
		statement: statement,
		node:      node,
	}
}

func (self *ConditionalExpression) IsTrue() bool {
	exprNodes := self.node.findN(2, ruleExpression)

	switch len(exprNodes) {
	case 1:
		if value, err := NewExpression(self.statement, exprNodes[0]).Value(); err == nil {
			return isTruthy(value)
		} else {
			log.Fatalf("malformed conditional expression: %v", err)
		}

	case 2:
		if cmp, err := parseComparator(self.node.firstN(2, ruleComparisonOperator)); err == nil {
			return cmp.Evaluate(
				NewExpression(self.statement, exprNodes[0]),
				NewExpression(self.statement, exprNodes[1]),
			)
		} else {
			log.Fatalf("malformed conditional expression: %v", err)
		}
	}

	log.Fatal("malformed conditional expression")
	return false
}

func isTruthy(value interface{}) bool {
	if v, err := exprToValue(value); err == nil {
		value = v
	} else {
		log.Fatal(err)
		return false
	}

	if typeutil.IsEmpty(value) || typeutil.IsZero(value) {
		return false
	} else if stringutil.IsBooleanFalse(value) {
		return false
	}

	return true
}
