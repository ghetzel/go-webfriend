package scripting

import (
	"fmt"
	"regexp"
)

type ConditionalType int

const (
	ConditionWithAssignment ConditionalType = iota
	ConditionWithCommand
	ConditionWithRegex
	ConditionWithComparator
)

func (self ConditionalType) String() string {
	switch self {
	case ConditionWithAssignment:
		return `ConditionWithAssignment`
	case ConditionWithCommand:
		return `ConditionWithCommand`
	case ConditionWithRegex:
		return `ConditionWithRegex`
	case ConditionWithComparator:
		return `ConditionWithComparator`
	default:
		return `UNKNOWN`
	}
}

type Conditional struct {
	statement *Statement
	n         *node32
}

func (self *Conditional) String() string {
	return fmt.Sprintf("if<%v>", self.Type())
}

func (self *Conditional) Type() ConditionalType {
	if condType := self.testStatementNode(); condType != nil {
		switch condType.rule() {
		case ruleConditionWithAssignment:
			return ConditionWithAssignment
		case ruleConditionWithCommand:
			return ConditionWithCommand
		case ruleConditionWithRegex:
			return ConditionWithRegex
		case ruleConditionWithComparator:
			return ConditionWithComparator
		}
	}

	return -1
}

// Return the objects necessary to perform assignment then evaluate an expression
func (self *Conditional) WithAssignment() (*Assignment, *ConditionalExpression) {
	if condType := self.testStatementNode(); condType != nil {
		assignment := self.statement.makeAssignment(condType.firstChild(ruleAssignment))
		condition := NewConditionalExpression(self.statement, condType.firstChild(ruleConditionalExpression))

		return assignment, condition
	}

	log.Fatal("malformed conditional statement")
	return nil, nil
}

// Return the objects necessary to execute a command then evaluate an expression
func (self *Conditional) WithCommand() (*Command, *ConditionalExpression) {
	if condType := self.testStatementNode(); condType != nil {
		command := &Command{
			statement: self.statement,
			node:      condType.firstChild(ruleCommand),
		}

		condition := NewConditionalExpression(self.statement, condType.firstChild(ruleConditionalExpression))

		return command, condition
	}

	log.Fatal("malformed conditional statement")
	return nil, nil
}

// Return the expression, operator, and regular expression in a regex if-test
func (self *Conditional) WithRegex() (*Expression, MatchOperator, *regexp.Regexp) {
	if condType := self.testStatementNode(); condType != nil {
		exprNode := condType.firstChild(ruleExpression)
		matchNode := condType.firstChild(ruleMatchOperator)
		regxNode := condType.firstChild(ruleRegularExpression)

		expr := NewExpression(self.statement, exprNode)

		if rx, err := self.statement.parseRegex(regxNode); err == nil {
			if cmp, err := parseMatchComparator(matchNode); err == nil {
				return expr, cmp, rx
			} else {
				log.Fatalf("malformed match operator: %v", err)
			}
		} else {
			log.Fatalf("malformed regular expression: %v", err)
		}
	}

	log.Fatal("malformed conditional statement")
	return nil, -1, nil
}

// Return the the left- and right-hand sides of an if-test, joined by the comparator.
func (self *Conditional) WithComparator() (*Expression, Comparator, *Expression) {
	if condType := self.testStatementNode(); condType != nil {
		var lhsNode, rhsNode *node32

		lhsNode = condType.first(ruleConditionWithComparatorLHS)
		rhsNode = condType.first(ruleConditionWithComparatorRHS)

		if lhsNode == nil {
			log.Fatal("malformed conditional statement: missing left-hand side expression")
		}

		if rhsNode == nil {
			return NewExpression(self.statement, lhsNode.firstChild(ruleExpression)), -1, nil
		} else {
			if node := rhsNode.firstChild(ruleComparisonOperator); node != nil {
				if cmp, err := parseComparator(node); err == nil {
					lhs := NewExpression(self.statement, lhsNode.firstChild(ruleExpression))
					rhs := NewExpression(self.statement, rhsNode.firstChild(ruleExpression))

					return lhs, cmp, rhs
				}
			} else {
				log.Fatal("malformed conditional statement: missing comparator")
			}
		}
	}

	log.Fatal("malformed conditional statement")
	return nil, -1, nil
}

func (self *Conditional) node() *node32 {
	if self.n != nil {
		return self.n
	} else {
		return self.statement.node
	}
}

func (self *Conditional) IsNegated() bool {
	if condEx := self.node().first(ruleConditionalExpression); condEx != nil {
		if condType := condEx.firstChild(ruleNOT); condType != nil {
			return true
		}
	}

	return false
}

func (self *Conditional) testStatementNode() *node32 {
	if condEx := self.node().first(ruleConditionalExpression); condEx != nil {
		if condType := condEx.firstChild(
			ruleConditionWithAssignment,
			ruleConditionWithCommand,
			ruleConditionWithRegex,
			ruleConditionWithComparator,
		); condType != nil {
			return condType
		}
	}

	return nil
}

func (self *Conditional) ifNode() *node32 {
	return self.node().firstChild(ruleIfStanza)
}

func (self *Conditional) elseNode() *node32 {
	return self.node().firstChild(ruleElseStanza)
}

func (self *Conditional) blocksFor(branch *node32) []*Block {
	blocks := make([]*Block, 0)

	if branch != nil {
		for _, node := range branch.findUntil(0, ruleCLOSE, ruleBlock) {
			blocks = append(blocks, &Block{
				friendscript: self.statement.Script(),
				node:         node.first(),
				parent:       self.statement,
			})
		}
	}

	return blocks
}

func (self *Conditional) IfBlocks() []*Block {
	return self.blocksFor(self.ifNode())
}

func (self *Conditional) ElseIfConditions() []*Conditional {
	branches := make([]*Conditional, 0)

	for _, branch := range self.node().children(ruleElseIfStanza) {
		branches = append(branches, &Conditional{
			statement: self.statement,
			n:         branch,
		})
	}

	return branches
}

func (self *Conditional) ElseBlocks() []*Block {
	return self.blocksFor(self.elseNode())
}
