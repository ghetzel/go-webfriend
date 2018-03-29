package scripting

import "fmt"

type Assignment struct {
	LeftHandSide  []string
	Operator      AssignmentOperator
	RightHandSide []*Expression
	statement     *Statement
}

func (self *Assignment) String() string {
	return fmt.Sprintf("%v %v (%d expressions)", self.LeftHandSide, self.Operator, len(self.RightHandSide))
}
