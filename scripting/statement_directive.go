package scripting

import (
	"fmt"
)

type Directive struct {
	statement *Statement
}

type emptyValue int
type DirectiveType int

const (
	UnknownDirective DirectiveType = iota
	UnsetDirective
	IncludeDirective
	DeclareDirective
)

func (self DirectiveType) String() string {
	switch self {
	case UnsetDirective:
		return `UnsetDirective`
	case IncludeDirective:
		return `IncludeDirective`
	case DeclareDirective:
		return `DeclareDirective`
	default:
		return `UnknownDirective`
	}
}

func (self *Directive) Type() DirectiveType {
	if self.statement.node.first(ruleDirectiveUnset) != nil {
		return UnsetDirective
	}

	if self.statement.node.first(ruleDirectiveInclude) != nil {
		return IncludeDirective
	}

	if self.statement.node.first(ruleDirectiveDeclare) != nil {
		return DeclareDirective
	}

	return UnknownDirective
}

func (self *Directive) VariableNames() []string {
	names := make([]string, 0)

	for _, node := range self.statement.node.find(ruleVariable) {
		names = append(names, self.statement.raw(node))
	}

	return names
}

func (self *Directive) String() string {
	return fmt.Sprintf("%v", self.Type())
}
