package scripting

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/ghetzel/go-stockutil/stringutil"
)

type Statement struct {
	node  *node32
	block *Block
}

type StatementType int

const (
	UnknownStatement StatementType = iota
	AssignmentStatement
	DirectiveStatement
	ExpressionStatement
	CommandStatement
	ConditionalStatement
	LoopStatement
	FlowControlStatement
	NoOpStatement
)

func (self StatementType) String() string {
	switch self {
	case AssignmentStatement:
		return `AssignmentStatement`
	case DirectiveStatement:
		return `DirectiveStatement`
	case ExpressionStatement:
		return `ExpressionStatement`
	case CommandStatement:
		return `CommandStatement`
	case ConditionalStatement:
		return `ConditionalStatement`
	case LoopStatement:
		return `LoopStatement`
	case FlowControlStatement:
		return `FlowControlStatement`
	case NoOpStatement:
		return `NoOpStatement`
	default:
		return `UnknownStatement`
	}
}

func (self *Statement) Script() *Friendscript {
	return self.block.Script()
}

func (self *Statement) raw(node *node32) string {
	return self.Script().s(node)
}

func (self *Statement) s(node *node32) string {
	if node != nil {
		raw := self.raw(node)

		if child := node.firstChild(); child != nil {
			switch child.rule() {
			case ruleStringLiteral:
				raw = strings.TrimPrefix(raw, `'`)
				raw = strings.TrimSuffix(raw, `'`)
				return raw

			case ruleStringInterpolated:
				raw = strings.TrimPrefix(raw, `"`)
				raw = strings.TrimSuffix(raw, `"`)
				return globalScope.Interpolate(raw)

			case ruleHeredoc:
				raw = self.raw(child.firstChild(ruleHeredocBody))
				lcp := self.Script().lcp()

				if lcp != `` {
					lines := strings.Split(raw, "\n")

					for i, line := range lines {
						lines[i] = strings.TrimPrefix(line, lcp)
					}

					raw = strings.Join(lines, "\n")
				}

				return globalScope.Interpolate(raw)
			default:
				return raw
			}
		}
	}

	return ``
}

func (self *Statement) Type() StatementType {
	if self.block.Type() == StatementBlock {
		switch self.node.rule() {
		case ruleSEMI:
			return NoOpStatement
		case ruleAssignment:
			return AssignmentStatement
		case ruleDirective:
			return DirectiveStatement
		case ruleExpression:
			return ExpressionStatement
		case ruleLoop:
			return LoopStatement
		case ruleCommand:
			return CommandStatement
		case ruleConditional:
			return ConditionalStatement
		}
	}

	return UnknownStatement
}

func (self *Statement) Assignment() *Assignment {
	if self.Type() == AssignmentStatement {
		return self.makeAssignment(self.node)
	}

	return nil
}

func (self *Statement) Directive() *Directive {
	if self.Type() == DirectiveStatement {
		return &Directive{
			statement: self,
		}
	}

	return nil
}

func (self *Statement) Command() *Command {
	if self.Type() == CommandStatement {
		return &Command{
			statement: self,
			node:      self.node,
		}
	}

	return nil
}

func (self *Statement) Conditional() *Conditional {
	if self.Type() == ConditionalStatement {
		return &Conditional{
			statement: self,
		}
	}

	return nil
}

func (self *Statement) Loop() *Loop {
	if self.Type() == LoopStatement {
		return &Loop{
			statement:  self,
			iterations: -1,
		}
	}

	return nil
}

func (self *Statement) parseObject(node *node32) (map[string]interface{}, error) {
	output := make(map[string]interface{})

	if node != nil {
		if pairs := node.children(ruleKeyValuePair); len(pairs) > 0 {
			for _, pair := range pairs {
				key := self.s(pair.first(ruleKey))

				if value, err := self.parseValue(pair.first(ruleKValue)); err == nil {
					output[key] = value
				} else {
					return nil, err
				}
			}
		}
	}

	return output, nil
}

func (self *Statement) parseRegex(node *node32) (*regexp.Regexp, error) {
	if node.rule() == ruleRegularExpression {
		rx := self.raw(node)

		if strings.HasPrefix(rx, `/`) {
			rx = strings.TrimPrefix(rx, `/`)
			flags := ``

			if i := strings.LastIndex(rx, `/`); i > 0 {
				flags = strings.TrimPrefix(rx[i:], `/`)
				rx = rx[:i]
			}

			for _, flag := range flags {
				rx = `(?` + string(flag) + `)` + rx
			}

			return regexp.Compile(rx)
		} else {
			return nil, fmt.Errorf("malformed regex")
		}
	} else {
		return nil, fmt.Errorf("not a regex node")
	}
}

func (self *Statement) parseArray(node *node32) ([]interface{}, error) {
	output := make([]interface{}, 0)

	if node != nil {
		if seq := node.first(ruleExpressionSequence); seq != nil {
			for i, exprNode := range seq.children(ruleExpression) {
				if value, err := NewExpression(self, exprNode).Value(); err == nil {
					output = append(output, value)
				} else {
					return nil, fmt.Errorf("index %d: %v", i, err)
				}
			}
		}
	}

	return output, nil
}

func (self *Statement) parseValue(node *node32) (interface{}, error) {
	value := node.first(
		ruleArray,
		ruleObject,
		ruleExpression,
		ruleScalarType,
	)

	if value == nil {
		return new(emptyValue), nil
	}

	switch value.rule() {
	case ruleExpression:
		return NewExpression(self, value).Value()

	case ruleScalarType:
		value = value.first(
			ruleNullValue,
			ruleInteger,
			ruleFloat,
			ruleBoolean,
			ruleString,
		)

		if value == nil {
			return nil, fmt.Errorf("Type value does not have a valid value")
		}
	}

	switch value.rule() {
	case ruleBoolean:
		if self.raw(value) == `true` {
			return true, nil
		} else {
			return false, nil
		}

	case ruleNullValue:
		return new(emptyValue), nil

	case ruleInteger:
		return stringutil.MustInteger(self.raw(value)), nil

	case ruleFloat:
		return stringutil.MustFloat(self.raw(value)), nil

	case ruleString:
		return self.s(value), nil

	case ruleArray:
		return self.parseArray(value)

	case ruleObject:
		return self.parseObject(value)

	default:
		return nil, fmt.Errorf("unhandled rule %v", rul3s[value.rule()])
	}

	return nil, fmt.Errorf("Unrecognized value type")
}

func (self *Statement) resolveVariableKey(node *node32) (string, error) {
	if node.rule() == ruleVariable {
		child := node.firstChild()
		keyparts := make([]string, 0)

		switch child.rule() {
		case ruleVariableNameSequence:
			nameNodes := child.children(ruleVariableName)

			for _, varpart := range nameNodes {
				identNode := varpart.firstChild(ruleIdentifier)
				ident := self.raw(identNode)
				keyparts = append(keyparts, ident)

				if index := identNode.firstChild(ruleVariableIndex); index != nil {
					if indexNode := index.firstChild(ruleExpression); indexNode != nil {
						if indexValue, err := NewExpression(self, indexNode).Value(); err == nil {
							keyparts = append(keyparts, fmt.Sprintf("%v", indexValue))
						} else {
							return ``, err
						}
					} else {
						return ``, fmt.Errorf("expected expression for index key")
					}
				}
			}

			return strings.Join(keyparts, `.`), nil

		case ruleSKIPVAR:
			return ``, nil

		default:
			return ``, fmt.Errorf("invalid variable usage '%v'", self.raw(node))
		}
	} else {
		return ``, fmt.Errorf("expected variable, got %v", node)
	}
}

func (self *Statement) resolveVariable(node *node32) (interface{}, error) {
	if key, err := self.resolveVariableKey(node); err == nil {
		if key == `` {
			return nil, nil
		} else {
			return globalScope.Get(key), nil
		}
	} else {
		return nil, fmt.Errorf("expected variable, got %v", node)
	}
}

func (self *Statement) makeAssignment(node *node32) *Assignment {
	lhs := node.first(ruleAssignmentLHS)
	op := node.first(ruleAssignmentOperator)
	rhs := node.first(ruleAssignmentRHS)

	if lhs != nil && rhs != nil {
		if aop, err := parseAssignmentOperator(op); err == nil {
			names := make([]string, 0)
			expressions := make([]*Expression, 0)

			for _, varNode := range lhs.first().children(ruleVariable) {
				if key, err := self.resolveVariableKey(varNode); err == nil {
					names = append(names, key)
				} else {
					panic(fmt.Errorf("unable to resolve variable name: %v", err))
				}
			}

			for _, exprNode := range rhs.first().children(ruleExpression) {
				expressions = append(expressions, NewExpression(self, exprNode))
			}

			return &Assignment{
				LeftHandSide:  names,
				Operator:      aop,
				RightHandSide: expressions,
				statement:     self,
			}
		} else {
			panic(fmt.Errorf("invalid assignment operator: %v", err))
		}
	}

	panic("cannot build Assignment from given node")
}
