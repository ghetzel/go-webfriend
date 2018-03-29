package scripting

import (
	"fmt"
	"strings"
)

type Expression struct {
	statement *Statement
	node      *node32
}

func NewExpression(statement *Statement, node *node32) *Expression {
	if node == nil || node.first(ruleValueYielding) == nil {
		panic("expression node must have a ValueYielding child")
	}

	return &Expression{
		statement: statement,
		node:      node,
	}
}

func (self *Expression) Script() *Friendscript {
	return self.statement.Script()
}

func (self *Expression) String() string {
	str := strings.TrimSpace(self.statement.raw(self.node))
	return str
}

func (self *Expression) Value() (interface{}, error) {
	if lhs := self.node.first(ruleExpressionLHS); lhs != nil {
		if value, err := self.resolveValue(lhs.firstChild(ruleValueYielding)); err == nil {
			if rhs := self.node.first(ruleExpressionRHS); rhs != nil {
				if op, err := parseOperator(rhs.firstChild(ruleOperator)); err == nil {
					if exprNode := rhs.firstChild(ruleExpression); exprNode != nil {
						return op.evaluate(value, NewExpression(self.statement, exprNode))
					}
				} else if op != opNull {
					return new(emptyValue), err
				}
			}

			// log.Debugf("EXPR %T(%v)", value, value)
			return value, nil
		} else {
			return nil, fmt.Errorf("invalid value: %v", err)
		}
	} else {
		return nil, fmt.Errorf("left-hand side of expression did not yield a value")
	}

	return new(emptyValue), nil
}

func (self *Expression) resolveValue(node *node32) (interface{}, error) {
	// expand variables
	if varNode := node.firstN(1, ruleVariable); varNode != nil {
		return self.statement.resolveVariable(varNode)

	} else if typeNode := node.firstN(1, ruleType); typeNode != nil {
		return self.statement.parseValue(typeNode)
	} else {
		return nil, fmt.Errorf("invalid value argument '%v'", self.statement.raw(node))
	}
}

func exprToValue(in interface{}) (interface{}, error) {
	if in == nil {
		return nil, nil
	} else if expr, ok := in.(*Expression); ok && expr != nil {
		if v, err := expr.Value(); err == nil {
			in = v
		} else {
			return nil, err
		}
	} else if expr, ok := in.(Expression); ok {
		if v, err := expr.Value(); err == nil {
			in = v
		} else {
			return nil, err
		}
	}

	return in, nil
}
