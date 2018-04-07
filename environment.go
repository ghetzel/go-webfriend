package webfriend

import (
	"fmt"
	"io"
	"io/ioutil"
	"sort"
	"strings"

	"github.com/ghetzel/go-stockutil/log"
	"github.com/ghetzel/go-stockutil/maputil"
	"github.com/ghetzel/go-stockutil/sliceutil"
	"github.com/ghetzel/go-stockutil/typeutil"
	"github.com/ghetzel/go-webfriend/browser"
	"github.com/ghetzel/go-webfriend/commands"
	"github.com/ghetzel/go-webfriend/commands/core"
	"github.com/ghetzel/go-webfriend/commands/page"
	"github.com/ghetzel/go-webfriend/scripting"
)

type Environment struct {
	Core    *core.Commands
	Page    *page.Commands
	modules map[string]commands.Module
	browser *browser.Browser
	script  *scripting.Friendscript
	stack   []*scripting.Scope
}

func NewEnvironment(browser *browser.Browser) *Environment {
	environment := &Environment{
		Core:    core.New(browser),
		Page:    page.New(browser),
		browser: browser,
		stack:   make([]*scripting.Scope, 0),
	}

	environment.modules = map[string]commands.Module{
		`core`: environment.Core,
		`page`: environment.Page,
	}

	return environment
}

func (self *Environment) EvaluateReader(reader io.Reader, scope ...*scripting.Scope) (*scripting.Scope, error) {
	if data, err := ioutil.ReadAll(reader); err == nil {
		return self.EvaluateString(string(data), scope...)
	} else {
		return nil, err
	}
}

func (self *Environment) EvaluateString(data string, scope ...*scripting.Scope) (*scripting.Scope, error) {
	if script, err := scripting.Parse(data); err == nil {
		return self.Evaluate(script)
	} else {
		return nil, err
	}
}

func (self *Environment) Evaluate(script *scripting.Friendscript, scope ...*scripting.Scope) (*scripting.Scope, error) {
	var rootScope *scripting.Scope

	if len(scope) > 0 && scope[0] != nil {
		rootScope = scope[0]
	} else {
		rootScope = scripting.NewScope(nil)
	}

	self.script = script
	self.pushScope(rootScope)

	for _, block := range script.Blocks() {
		if err := self.evaluateBlock(block); err != nil {
			return self.scope(), err
		}
	}

	return self.scope(), nil
}

func (self *Environment) pushScope(scope *scripting.Scope) {
	if len(self.stack) > 0 {
		log.Debugf("PUSH scope(%d) is masked", self.scope().Level())
	} else {
		log.Debugf("PUSH scope(%d) is ROOT", scope.Level())
	}

	self.stack = append(self.stack, scope)

	if self.script != nil {
		scripting.SetScope(self.scope())
	}

	log.Debugf("PUSH scope(%d) is active", self.scope().Level())
}

func (self *Environment) scope() *scripting.Scope {
	if len(self.stack) > 0 {
		return self.stack[len(self.stack)-1]
	} else {
		log.Fatal("illegal scope retrieval from empty stack")
		return nil
	}
}

func (self *Environment) popScope() *scripting.Scope {
	if len(self.stack) > 1 {
		top := self.stack[len(self.stack)-1]
		self.stack = self.stack[0 : len(self.stack)-1]

		if self.script != nil {
			scripting.SetScope(self.scope())
		}

		log.Debugf("POP  scope(%d) is active", self.scope().Level())

		return top
	} else if len(self.stack) == 1 {
		return self.stack[0]
	} else {
		log.Fatal("attempted pop on an empty scope stack")
		return nil
	}
}

func (self *Environment) evaluateBlock(block *scripting.Block) error {
	log.Debug(strings.Repeat("-", 70))

	switch block.Type() {
	case scripting.StatementBlock:
		for _, statement := range block.Statements() {
			if err := self.evaluateStatement(statement); err != nil {
				return err
			}
		}

	case scripting.EventHandlerBlock:
		return fmt.Errorf("Not Implemented")

	case scripting.FlowControlWord:
		if levels := block.FlowBreak(); levels > 0 {
			return scripting.NewFlowControl(scripting.FlowBreak, levels)
		} else if levels := block.FlowContinue(); levels > 0 {
			return scripting.NewFlowControl(scripting.FlowContinue, levels)
		} else {
			return fmt.Errorf("invalid flow control statement")
		}
	}

	return nil
}

func (self *Environment) evaluateStatement(statement *scripting.Statement) error {
	switch statement.Type() {
	case scripting.AssignmentStatement:
		return self.evaluateAssignment(statement.Assignment(), false)

	case scripting.DirectiveStatement:
		return self.evaluateDirective(statement.Directive())

	case scripting.ExpressionStatement:
		fmt.Println("  ExpressionStatement")

	case scripting.ConditionalStatement:
		_, err := self.evaluateConditional(statement.Conditional())
		return err

	case scripting.LoopStatement:
		return self.evaluateLoop(statement.Loop())

	case scripting.FlowControlStatement:
		fmt.Println("  FlowControlStatement")

	case scripting.CommandStatement:
		_, err := self.evaluateCommand(statement.Command(), false)
		return err

	case scripting.NoOpStatement:
		return nil

	default:
		return fmt.Errorf("Unrecognized statement: %v", statement.Type())
	}

	return nil
}

func (self *Environment) evaluateAssignment(assignment *scripting.Assignment, forceDeclare bool) error {
	log.Debugf("ASSN %v", assignment)

	if assignment.Operator.ShouldPreclear() {
		// clear out all the left-hand side variables
		for _, lhs := range assignment.LeftHandSide {
			if forceDeclare {
				self.scope().Declare(lhs)
			} else {
				self.scope().Set(lhs, nil)
			}
		}
	}

	// unpack
	if len(assignment.RightHandSide) == 1 {
		if rhs, err := assignment.RightHandSide[0].Value(); err == nil {
			totalLhsCount := len(assignment.LeftHandSide)

			if totalLhsCount > 1 && typeutil.IsArray(rhs) {
				for i, rhs := range sliceutil.Sliceify(rhs) {
					if i < totalLhsCount {
						if result, err := assignment.Operator.Evaluate(
							self.scope().Get(assignment.LeftHandSide[i]),
							rhs,
						); err == nil {
							self.scope().Set(assignment.LeftHandSide[i], result)
						} else {
							return err
						}
					}
				}

				return nil
			}
		}
	}

	for i, lhs := range assignment.LeftHandSide {
		if i < len(assignment.RightHandSide) {
			if result, err := assignment.Operator.Evaluate(
				self.scope().Get(lhs),
				assignment.RightHandSide[i],
			); err == nil {
				self.scope().Set(lhs, result)
			} else {
				return err
			}
		}
	}

	return nil
}

func (self *Environment) evaluateDirective(directive *scripting.Directive) error {
	switch directive.Type() {
	case scripting.UnsetDirective:
		return fmt.Errorf("'unset' not implemented yet")
	case scripting.IncludeDirective:
		return fmt.Errorf("'include' not implemented yet")
	case scripting.DeclareDirective:
		for _, varname := range directive.VariableNames() {
			self.scope().Declare(varname)
		}
	}

	return nil
}

func (self *Environment) evaluateCommand(command *scripting.Command, forceDeclare bool) (string, error) {
	modname, name := command.Name()
	log.Debugf("EXEC %v::%v", modname, name)

	if first, rest, err := command.Args(); err == nil {
		// locate the module this command belongs to
		if module, ok := self.modules[modname]; ok {
			log.Debugf("CMND called %T(%v), %T(%v)", first, first, rest, rest)

			// tell that module to execute the command, giving it the name and arguments
			if result, err := module.ExecuteCommand(name, first, rest); err == nil {
				// log.Debugf("CMND returned %T(%v)", result, result)

				// if there is an output variable destination, set that in the current scope
				if resultVar := command.OutputName(); resultVar != `` {
					if forceDeclare {
						self.scope().Declare(resultVar)
					}

					self.scope().Set(resultVar, result)
					return resultVar, nil
				}

				return ``, nil
			} else {
				return ``, err
			}
		} else {
			return ``, fmt.Errorf("Cannot locate module %q", modname)
		}
	} else {
		return ``, fmt.Errorf("invalid arguments: %v", err)
	}
}

func (self *Environment) evaluateConditional(conditional *scripting.Conditional) (bool, error) {
	var blocks = make([]*scripting.Block, 0)
	var trueBranch bool
	var conditionScope = scripting.NewScope(self.scope())
	self.pushScope(conditionScope)
	defer self.popScope()

	switch conditional.Type() {
	case scripting.ConditionWithAssignment:
		assignment, condition := conditional.WithAssignment()

		if err := self.evaluateAssignment(assignment, true); err == nil {
			log.Debugf("IFWA evaluate condition")
			result := condition.IsTrue()
			blocks, trueBranch = self.evaluateConditionalGetBranch(conditional, result)
		} else {
			return trueBranch, err
		}

	case scripting.ConditionWithCommand:
		command, condition := conditional.WithCommand()

		log.Debugf("IFWC execute command")

		if _, err := self.evaluateCommand(command, true); err == nil {
			result := condition.IsTrue()
			blocks, trueBranch = self.evaluateConditionalGetBranch(conditional, result)
		} else {
			return trueBranch, err
		}

	case scripting.ConditionWithRegex:
		expression, matchOp, rx := conditional.WithRegex()
		log.Debugf("IFRX parse regex")

		result := matchOp.Evaluate(rx, expression)
		blocks, trueBranch = self.evaluateConditionalGetBranch(conditional, result)

	case scripting.ConditionWithComparator:
		log.Debugf("IFCM do compare")

		lhs, cmp, rhs := conditional.WithComparator()

		result := cmp.Evaluate(lhs, rhs)
		blocks, trueBranch = self.evaluateConditionalGetBranch(conditional, result)

	default:
		return trueBranch, fmt.Errorf("Unrecognized Conditional type")
	}

	for _, block := range blocks {
		if err := self.evaluateBlock(block); err != nil {
			return trueBranch, err
		}
	}

	return trueBranch, nil
}

func (self *Environment) evaluateConditionalGetBranch(conditional *scripting.Conditional, result bool) ([]*scripting.Block, bool) {
	var blocks = make([]*scripting.Block, 0)
	var trueBranch bool

	if conditional.IsNegated() {
		result = !result
	}

	if result {
		log.Debugf("IF branch")
		blocks = conditional.IfBlocks()
		trueBranch = true
	} else {
		var tookElifBranch bool

		for ei, elif := range conditional.ElseIfConditions() {
			if t, err := self.evaluateConditional(elif); err == nil {
				if t {
					log.Debugf("ELSE-IF %d branch", ei)
					tookElifBranch = true
					blocks = elif.IfBlocks()
					break
				}
			} else {
				log.Fatal(err)
				return nil, false
			}
		}

		if !tookElifBranch {
			log.Debugf("ELSE branch")
			blocks = conditional.ElseBlocks()
		}
	}

	return blocks, trueBranch
}

func (self *Environment) evaluateLoop(loop *scripting.Loop) error {
	var i int
	var sourceVar string
	var destVars []string
	var loopScope = scripting.NewScope(self.scope())

	loopScope.Declare(`index`)

	self.pushScope(loopScope)
	defer self.popScope()

	// if we have an iterator, we have to initialize the values
	if loop.Type() == scripting.IteratorLoop {
		if s, d, err := self.evaluateLoopIterationStart(loop, loopScope); err == nil {
			sourceVar = s
			destVars = d

			log.Debugf("Iterator initialized: %v -> %v", sourceVar, destVars)
		} else {
			return err
		}
	}

	log.Debugf("LOOP BEGIN")

LoopEval:
	for loop.ShouldContinue() {
		if loop.Type() == scripting.IteratorLoop {
			iterVector := loopScope.Get(sourceVar)

			if typeutil.IsMap(iterVector) {
				remap := make([][]interface{}, 0)
				keys := maputil.StringKeys(iterVector)
				sort.Strings(keys)

				for _, key := range keys {
					remap = append(remap, []interface{}{
						key,
						maputil.Get(iterVector, key),
					})
				}

				iterVector = remap
			}

			if iterLen := sliceutil.Len(iterVector); i < iterLen {
				if iterItem, ok := sliceutil.At(iterVector, i); ok {
					var didSet bool

					if totalLhsCount := len(destVars); totalLhsCount > 1 {
						if typeutil.IsArray(iterItem) {
							for j, rhs := range sliceutil.Sliceify(iterItem) {
								if j < totalLhsCount {
									loopScope.Set(destVars[j], rhs)
									didSet = true
								}
							}
						}
					}

					if !didSet {
						loopScope.Set(destVars[0], iterItem)
					}
				} else {
					return fmt.Errorf("Failed to retrieve iterator item %d", i)
				}
			} else {
				break
			}
		}

		loopScope.Set(`index`, loop.CurrentIndex())

		for _, block := range loop.Blocks() {
			if err := self.evaluateBlock(block); err != nil {
				if fc, ok := err.(*scripting.FlowControlErr); ok {
					if fc.Level <= 0 {
						return fc
					} else if fc.Level == 1 {
						if fc.Type == scripting.FlowContinue {
							continue LoopEval
						} else {
							break LoopEval
						}
					} else {
						fc.Level = fc.Level - 1
						return fc
					}
				} else {
					return err
				}
			}
		}

		i += 1
	}

	log.Debugf("LOOP END")
	return nil
}

func (self *Environment) evaluateLoopIterationStart(loop *scripting.Loop, scope *scripting.Scope) (string, []string, error) {
	destVars, source := loop.IteratableParts()
	var sourceVar string

	if cmd, ok := source.(*scripting.Command); ok {
		if resultVar, err := self.evaluateCommand(cmd, true); err == nil {
			sourceVar = resultVar
		} else {
			return ``, nil, err
		}
	} else if srcvar, ok := source.(string); ok {
		sourceVar = srcvar
	}

	for _, v := range destVars {
		scope.Declare(v)
	}

	return sourceVar, destVars, nil
}
