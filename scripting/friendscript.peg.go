package scripting

//go:generate peg -inline friendscript.peg

import (
	"fmt"
	"math"
	"sort"
	"strconv"
)

const endSymbol rune = 1114112

/* The rule types inferred from the grammar are below. */
type pegRule uint8

const (
	ruleUnknown pegRule = iota
	ruleFriendscript
	rule_
	rule__
	ruleASSIGN
	ruleBEGIN
	ruleBREAK
	ruleCLOSE
	ruleCOLON
	ruleCOMMA
	ruleComment
	ruleCONT
	ruleCOUNT
	ruleDECLARE
	ruleDOT
	ruleELSE
	ruleEND
	ruleIF
	ruleIN
	ruleINCLUDE
	ruleLOOP
	ruleNOOP
	ruleNOT
	ruleOPEN
	ruleSCOPE
	ruleSEMI
	ruleSHEBANG
	ruleSKIPVAR
	ruleUNSET
	ruleOperator
	ruleExponentiate
	ruleMultiply
	ruleDivide
	ruleModulus
	ruleAdd
	ruleSubtract
	ruleBitwiseAnd
	ruleBitwiseOr
	ruleBitwiseNot
	ruleBitwiseXor
	ruleMatchOperator
	ruleUnmatch
	ruleMatch
	ruleAssignmentOperator
	ruleAssignEq
	ruleStarEq
	ruleDivEq
	rulePlusEq
	ruleMinusEq
	ruleAndEq
	ruleOrEq
	ruleAppend
	ruleComparisonOperator
	ruleEquality
	ruleNonEquality
	ruleGreaterThan
	ruleGreaterEqual
	ruleLessEqual
	ruleLessThan
	ruleMembership
	ruleNonMembership
	ruleVariable
	ruleVariableNameSequence
	ruleVariableName
	ruleVariableIndex
	ruleBlock
	ruleFlowControlWord
	ruleFlowControlBreak
	ruleFlowControlContinue
	ruleStatementBlock
	ruleAssignment
	ruleAssignmentLHS
	ruleAssignmentRHS
	ruleVariableSequence
	ruleExpressionSequence
	ruleExpression
	ruleExpressionLHS
	ruleExpressionRHS
	ruleValueYielding
	ruleDirective
	ruleDirectiveUnset
	ruleDirectiveInclude
	ruleDirectiveDeclare
	ruleCommand
	ruleCommandName
	ruleCommandFirstArg
	ruleCommandSecondArg
	ruleCommandResultAssignment
	ruleConditional
	ruleIfStanza
	ruleElseIfStanza
	ruleElseStanza
	ruleLoop
	ruleLoopConditionFixedLength
	ruleLoopConditionIterable
	ruleLoopIterableLHS
	ruleLoopIterableRHS
	ruleLoopConditionBounded
	ruleLoopConditionTruthy
	ruleConditionalExpression
	ruleConditionWithAssignment
	ruleConditionWithCommand
	ruleConditionWithRegex
	ruleConditionWithComparator
	ruleConditionWithComparatorLHS
	ruleConditionWithComparatorRHS
	ruleScalarType
	ruleIdentifier
	ruleFloat
	ruleBoolean
	ruleInteger
	rulePositiveInteger
	ruleString
	ruleStringLiteral
	ruleStringInterpolated
	ruleHeredoc
	ruleHeredocBody
	ruleNullValue
	ruleObject
	ruleArray
	ruleRegularExpression
	ruleKeyValuePair
	ruleKey
	ruleKValue
	ruleType
)

var rul3s = [...]string{
	"Unknown",
	"Friendscript",
	"_",
	"__",
	"ASSIGN",
	"BEGIN",
	"BREAK",
	"CLOSE",
	"COLON",
	"COMMA",
	"Comment",
	"CONT",
	"COUNT",
	"DECLARE",
	"DOT",
	"ELSE",
	"END",
	"IF",
	"IN",
	"INCLUDE",
	"LOOP",
	"NOOP",
	"NOT",
	"OPEN",
	"SCOPE",
	"SEMI",
	"SHEBANG",
	"SKIPVAR",
	"UNSET",
	"Operator",
	"Exponentiate",
	"Multiply",
	"Divide",
	"Modulus",
	"Add",
	"Subtract",
	"BitwiseAnd",
	"BitwiseOr",
	"BitwiseNot",
	"BitwiseXor",
	"MatchOperator",
	"Unmatch",
	"Match",
	"AssignmentOperator",
	"AssignEq",
	"StarEq",
	"DivEq",
	"PlusEq",
	"MinusEq",
	"AndEq",
	"OrEq",
	"Append",
	"ComparisonOperator",
	"Equality",
	"NonEquality",
	"GreaterThan",
	"GreaterEqual",
	"LessEqual",
	"LessThan",
	"Membership",
	"NonMembership",
	"Variable",
	"VariableNameSequence",
	"VariableName",
	"VariableIndex",
	"Block",
	"FlowControlWord",
	"FlowControlBreak",
	"FlowControlContinue",
	"StatementBlock",
	"Assignment",
	"AssignmentLHS",
	"AssignmentRHS",
	"VariableSequence",
	"ExpressionSequence",
	"Expression",
	"ExpressionLHS",
	"ExpressionRHS",
	"ValueYielding",
	"Directive",
	"DirectiveUnset",
	"DirectiveInclude",
	"DirectiveDeclare",
	"Command",
	"CommandName",
	"CommandFirstArg",
	"CommandSecondArg",
	"CommandResultAssignment",
	"Conditional",
	"IfStanza",
	"ElseIfStanza",
	"ElseStanza",
	"Loop",
	"LoopConditionFixedLength",
	"LoopConditionIterable",
	"LoopIterableLHS",
	"LoopIterableRHS",
	"LoopConditionBounded",
	"LoopConditionTruthy",
	"ConditionalExpression",
	"ConditionWithAssignment",
	"ConditionWithCommand",
	"ConditionWithRegex",
	"ConditionWithComparator",
	"ConditionWithComparatorLHS",
	"ConditionWithComparatorRHS",
	"ScalarType",
	"Identifier",
	"Float",
	"Boolean",
	"Integer",
	"PositiveInteger",
	"String",
	"StringLiteral",
	"StringInterpolated",
	"Heredoc",
	"HeredocBody",
	"NullValue",
	"Object",
	"Array",
	"RegularExpression",
	"KeyValuePair",
	"Key",
	"KValue",
	"Type",
}

type token32 struct {
	pegRule
	begin, end uint32
}

func (t *token32) String() string {
	return fmt.Sprintf("\x1B[34m%v\x1B[m %v %v", rul3s[t.pegRule], t.begin, t.end)
}

type node32 struct {
	token32
	up, next *node32
}

func (node *node32) print(pretty bool, buffer string) {
	var print func(node *node32, depth int)
	print = func(node *node32, depth int) {
		for node != nil {
			for c := 0; c < depth; c++ {
				fmt.Printf(" ")
			}
			rule := rul3s[node.pegRule]
			quote := strconv.Quote(string(([]rune(buffer)[node.begin:node.end])))
			if !pretty {
				fmt.Printf("%v %v\n", rule, quote)
			} else {
				fmt.Printf("\x1B[34m%v\x1B[m %v\n", rule, quote)
			}
			if node.up != nil {
				print(node.up, depth+1)
			}
			node = node.next
		}
	}
	print(node, 0)
}

func (node *node32) Print(buffer string) {
	node.print(false, buffer)
}

func (node *node32) PrettyPrint(buffer string) {
	node.print(true, buffer)
}

type tokens32 struct {
	tree []token32
}

func (t *tokens32) Trim(length uint32) {
	t.tree = t.tree[:length]
}

func (t *tokens32) Print() {
	for _, token := range t.tree {
		fmt.Println(token.String())
	}
}

func (t *tokens32) AST() *node32 {
	type element struct {
		node *node32
		down *element
	}
	tokens := t.Tokens()
	var stack *element
	for _, token := range tokens {
		if token.begin == token.end {
			continue
		}
		node := &node32{token32: token}
		for stack != nil && stack.node.begin >= token.begin && stack.node.end <= token.end {
			stack.node.next = node.up
			node.up = stack.node
			stack = stack.down
		}
		stack = &element{node: node, down: stack}
	}
	if stack != nil {
		return stack.node
	}
	return nil
}

func (t *tokens32) PrintSyntaxTree(buffer string) {
	t.AST().Print(buffer)
}

func (t *tokens32) PrettyPrintSyntaxTree(buffer string) {
	t.AST().PrettyPrint(buffer)
}

func (t *tokens32) Add(rule pegRule, begin, end, index uint32) {
	if tree := t.tree; int(index) >= len(tree) {
		expanded := make([]token32, 2*len(tree))
		copy(expanded, tree)
		t.tree = expanded
	}
	t.tree[index] = token32{
		pegRule: rule,
		begin:   begin,
		end:     end,
	}
}

func (t *tokens32) Tokens() []token32 {
	return t.tree
}

type Friendscript struct {
	runtime

	Buffer string
	buffer []rune
	rules  [125]func() bool
	parse  func(rule ...int) error
	reset  func()
	Pretty bool
	tokens32
}

func (p *Friendscript) Parse(rule ...int) error {
	return p.parse(rule...)
}

func (p *Friendscript) Reset() {
	p.reset()
}

type textPosition struct {
	line, symbol int
}

type textPositionMap map[int]textPosition

func translatePositions(buffer []rune, positions []int) textPositionMap {
	length, translations, j, line, symbol := len(positions), make(textPositionMap, len(positions)), 0, 1, 0
	sort.Ints(positions)

search:
	for i, c := range buffer {
		if c == '\n' {
			line, symbol = line+1, 0
		} else {
			symbol++
		}
		if i == positions[j] {
			translations[positions[j]] = textPosition{line, symbol}
			for j++; j < length; j++ {
				if i != positions[j] {
					continue search
				}
			}
			break search
		}
	}

	return translations
}

type parseError struct {
	p   *Friendscript
	max token32
}

func (e *parseError) Error() string {
	tokens, error := []token32{e.max}, "\n"
	positions, p := make([]int, 2*len(tokens)), 0
	for _, token := range tokens {
		positions[p], p = int(token.begin), p+1
		positions[p], p = int(token.end), p+1
	}
	translations := translatePositions(e.p.buffer, positions)
	format := "parse error near %v (line %v symbol %v - line %v symbol %v):\n%v\n"
	if e.p.Pretty {
		format = "parse error near \x1B[34m%v\x1B[m (line %v symbol %v - line %v symbol %v):\n%v\n"
	}
	for _, token := range tokens {
		begin, end := int(token.begin), int(token.end)
		error += fmt.Sprintf(format,
			rul3s[token.pegRule],
			translations[begin].line, translations[begin].symbol,
			translations[end].line, translations[end].symbol,
			strconv.Quote(string(e.p.buffer[begin:end])))
	}

	return error
}

func (p *Friendscript) PrintSyntaxTree() {
	if p.Pretty {
		p.tokens32.PrettyPrintSyntaxTree(p.Buffer)
	} else {
		p.tokens32.PrintSyntaxTree(p.Buffer)
	}
}

func (p *Friendscript) Init() {
	var (
		max                  token32
		position, tokenIndex uint32
		buffer               []rune
	)
	p.reset = func() {
		max = token32{}
		position, tokenIndex = 0, 0

		p.buffer = []rune(p.Buffer)
		if len(p.buffer) == 0 || p.buffer[len(p.buffer)-1] != endSymbol {
			p.buffer = append(p.buffer, endSymbol)
		}
		buffer = p.buffer
	}
	p.reset()

	_rules := p.rules
	tree := tokens32{tree: make([]token32, math.MaxInt16)}
	p.parse = func(rule ...int) error {
		r := 1
		if len(rule) > 0 {
			r = rule[0]
		}
		matches := p.rules[r]()
		p.tokens32 = tree
		if matches {
			p.Trim(tokenIndex)
			return nil
		}
		return &parseError{p, max}
	}

	add := func(rule pegRule, begin uint32) {
		tree.Add(rule, begin, position, tokenIndex)
		tokenIndex++
		if begin != position && position > max.end {
			max = token32{rule, begin, position}
		}
	}

	matchDot := func() bool {
		if buffer[position] != endSymbol {
			position++
			return true
		}
		return false
	}

	/*matchChar := func(c byte) bool {
		if buffer[position] == c {
			position++
			return true
		}
		return false
	}*/

	/*matchRange := func(lower byte, upper byte) bool {
		if c := buffer[position]; c >= lower && c <= upper {
			position++
			return true
		}
		return false
	}*/

	_rules = [...]func() bool{
		nil,
		/* 0 Friendscript <- <(_ SHEBANG? _ Block* !.)> */
		func() bool {
			position0, tokenIndex0 := position, tokenIndex
			{
				position1 := position
				if !_rules[rule_]() {
					goto l0
				}
				{
					position2, tokenIndex2 := position, tokenIndex
					{
						position4 := position
						if buffer[position] != rune('#') {
							goto l2
						}
						position++
						if buffer[position] != rune('!') {
							goto l2
						}
						position++
						{
							position7, tokenIndex7 := position, tokenIndex
							if buffer[position] != rune('\n') {
								goto l7
							}
							position++
							goto l2
						l7:
							position, tokenIndex = position7, tokenIndex7
						}
						if !matchDot() {
							goto l2
						}
					l5:
						{
							position6, tokenIndex6 := position, tokenIndex
							{
								position8, tokenIndex8 := position, tokenIndex
								if buffer[position] != rune('\n') {
									goto l8
								}
								position++
								goto l6
							l8:
								position, tokenIndex = position8, tokenIndex8
							}
							if !matchDot() {
								goto l6
							}
							goto l5
						l6:
							position, tokenIndex = position6, tokenIndex6
						}
						if buffer[position] != rune('\n') {
							goto l2
						}
						position++
						add(ruleSHEBANG, position4)
					}
					goto l3
				l2:
					position, tokenIndex = position2, tokenIndex2
				}
			l3:
				if !_rules[rule_]() {
					goto l0
				}
			l9:
				{
					position10, tokenIndex10 := position, tokenIndex
					if !_rules[ruleBlock]() {
						goto l10
					}
					goto l9
				l10:
					position, tokenIndex = position10, tokenIndex10
				}
				{
					position11, tokenIndex11 := position, tokenIndex
					if !matchDot() {
						goto l11
					}
					goto l0
				l11:
					position, tokenIndex = position11, tokenIndex11
				}
				add(ruleFriendscript, position1)
			}
			return true
		l0:
			position, tokenIndex = position0, tokenIndex0
			return false
		},
		/* 1 _ <- <(' ' / '\t' / '\r' / '\n')*> */
		func() bool {
			{
				position13 := position
			l14:
				{
					position15, tokenIndex15 := position, tokenIndex
					{
						position16, tokenIndex16 := position, tokenIndex
						if buffer[position] != rune(' ') {
							goto l17
						}
						position++
						goto l16
					l17:
						position, tokenIndex = position16, tokenIndex16
						if buffer[position] != rune('\t') {
							goto l18
						}
						position++
						goto l16
					l18:
						position, tokenIndex = position16, tokenIndex16
						if buffer[position] != rune('\r') {
							goto l19
						}
						position++
						goto l16
					l19:
						position, tokenIndex = position16, tokenIndex16
						if buffer[position] != rune('\n') {
							goto l15
						}
						position++
					}
				l16:
					goto l14
				l15:
					position, tokenIndex = position15, tokenIndex15
				}
				add(rule_, position13)
			}
			return true
		},
		/* 2 __ <- <(' ' / '\t' / '\r' / '\n')+> */
		func() bool {
			position20, tokenIndex20 := position, tokenIndex
			{
				position21 := position
				{
					position24, tokenIndex24 := position, tokenIndex
					if buffer[position] != rune(' ') {
						goto l25
					}
					position++
					goto l24
				l25:
					position, tokenIndex = position24, tokenIndex24
					if buffer[position] != rune('\t') {
						goto l26
					}
					position++
					goto l24
				l26:
					position, tokenIndex = position24, tokenIndex24
					if buffer[position] != rune('\r') {
						goto l27
					}
					position++
					goto l24
				l27:
					position, tokenIndex = position24, tokenIndex24
					if buffer[position] != rune('\n') {
						goto l20
					}
					position++
				}
			l24:
			l22:
				{
					position23, tokenIndex23 := position, tokenIndex
					{
						position28, tokenIndex28 := position, tokenIndex
						if buffer[position] != rune(' ') {
							goto l29
						}
						position++
						goto l28
					l29:
						position, tokenIndex = position28, tokenIndex28
						if buffer[position] != rune('\t') {
							goto l30
						}
						position++
						goto l28
					l30:
						position, tokenIndex = position28, tokenIndex28
						if buffer[position] != rune('\r') {
							goto l31
						}
						position++
						goto l28
					l31:
						position, tokenIndex = position28, tokenIndex28
						if buffer[position] != rune('\n') {
							goto l23
						}
						position++
					}
				l28:
					goto l22
				l23:
					position, tokenIndex = position23, tokenIndex23
				}
				add(rule__, position21)
			}
			return true
		l20:
			position, tokenIndex = position20, tokenIndex20
			return false
		},
		/* 3 ASSIGN <- <(_ ('-' '>') _)> */
		nil,
		/* 4 BEGIN <- <(_ ('b' 'e' 'g' 'i' 'n'))> */
		nil,
		/* 5 BREAK <- <(_ ('b' 'r' 'e' 'a' 'k') _)> */
		nil,
		/* 6 CLOSE <- <(_ '}' _)> */
		func() bool {
			position35, tokenIndex35 := position, tokenIndex
			{
				position36 := position
				if !_rules[rule_]() {
					goto l35
				}
				if buffer[position] != rune('}') {
					goto l35
				}
				position++
				if !_rules[rule_]() {
					goto l35
				}
				add(ruleCLOSE, position36)
			}
			return true
		l35:
			position, tokenIndex = position35, tokenIndex35
			return false
		},
		/* 7 COLON <- <(_ ':' _)> */
		nil,
		/* 8 COMMA <- <(_ ',' _)> */
		func() bool {
			position38, tokenIndex38 := position, tokenIndex
			{
				position39 := position
				if !_rules[rule_]() {
					goto l38
				}
				if buffer[position] != rune(',') {
					goto l38
				}
				position++
				if !_rules[rule_]() {
					goto l38
				}
				add(ruleCOMMA, position39)
			}
			return true
		l38:
			position, tokenIndex = position38, tokenIndex38
			return false
		},
		/* 9 Comment <- <(_ '#' (!'\n' .)*)> */
		nil,
		/* 10 CONT <- <(_ ('c' 'o' 'n' 't' 'i' 'n' 'u' 'e') _)> */
		nil,
		/* 11 COUNT <- <(_ ('c' 'o' 'u' 'n' 't') _)> */
		nil,
		/* 12 DECLARE <- <(_ ('d' 'e' 'c' 'l' 'a' 'r' 'e') __)> */
		nil,
		/* 13 DOT <- <'.'> */
		nil,
		/* 14 ELSE <- <(_ ('e' 'l' 's' 'e') _)> */
		func() bool {
			position45, tokenIndex45 := position, tokenIndex
			{
				position46 := position
				if !_rules[rule_]() {
					goto l45
				}
				if buffer[position] != rune('e') {
					goto l45
				}
				position++
				if buffer[position] != rune('l') {
					goto l45
				}
				position++
				if buffer[position] != rune('s') {
					goto l45
				}
				position++
				if buffer[position] != rune('e') {
					goto l45
				}
				position++
				if !_rules[rule_]() {
					goto l45
				}
				add(ruleELSE, position46)
			}
			return true
		l45:
			position, tokenIndex = position45, tokenIndex45
			return false
		},
		/* 15 END <- <(_ ('e' 'n' 'd') _)> */
		func() bool {
			position47, tokenIndex47 := position, tokenIndex
			{
				position48 := position
				if !_rules[rule_]() {
					goto l47
				}
				if buffer[position] != rune('e') {
					goto l47
				}
				position++
				if buffer[position] != rune('n') {
					goto l47
				}
				position++
				if buffer[position] != rune('d') {
					goto l47
				}
				position++
				if !_rules[rule_]() {
					goto l47
				}
				add(ruleEND, position48)
			}
			return true
		l47:
			position, tokenIndex = position47, tokenIndex47
			return false
		},
		/* 16 IF <- <(_ ('i' 'f') _)> */
		nil,
		/* 17 IN <- <(__ ('i' 'n') __)> */
		nil,
		/* 18 INCLUDE <- <(_ ('i' 'n' 'c' 'l' 'u' 'd' 'e') __)> */
		nil,
		/* 19 LOOP <- <(_ ('l' 'o' 'o' 'p') _)> */
		nil,
		/* 20 NOOP <- <SEMI> */
		nil,
		/* 21 NOT <- <(_ ('n' 'o' 't') __)> */
		nil,
		/* 22 OPEN <- <(_ '{' _)> */
		func() bool {
			position55, tokenIndex55 := position, tokenIndex
			{
				position56 := position
				if !_rules[rule_]() {
					goto l55
				}
				if buffer[position] != rune('{') {
					goto l55
				}
				position++
				if !_rules[rule_]() {
					goto l55
				}
				add(ruleOPEN, position56)
			}
			return true
		l55:
			position, tokenIndex = position55, tokenIndex55
			return false
		},
		/* 23 SCOPE <- <(':' ':')> */
		nil,
		/* 24 SEMI <- <(_ ';' _)> */
		func() bool {
			position58, tokenIndex58 := position, tokenIndex
			{
				position59 := position
				if !_rules[rule_]() {
					goto l58
				}
				if buffer[position] != rune(';') {
					goto l58
				}
				position++
				if !_rules[rule_]() {
					goto l58
				}
				add(ruleSEMI, position59)
			}
			return true
		l58:
			position, tokenIndex = position58, tokenIndex58
			return false
		},
		/* 25 SHEBANG <- <('#' '!' (!'\n' .)+ '\n')> */
		nil,
		/* 26 SKIPVAR <- <(_ '_' _)> */
		nil,
		/* 27 UNSET <- <(_ ('u' 'n' 's' 'e' 't') __)> */
		nil,
		/* 28 Operator <- <(_ (Exponentiate / Multiply / Divide / Modulus / Add / Subtract / BitwiseAnd / BitwiseOr / BitwiseNot / BitwiseXor) _)> */
		nil,
		/* 29 Exponentiate <- <(_ ('*' '*') _)> */
		nil,
		/* 30 Multiply <- <(_ '*' _)> */
		nil,
		/* 31 Divide <- <(_ '/' _)> */
		nil,
		/* 32 Modulus <- <(_ '%' _)> */
		nil,
		/* 33 Add <- <(_ '+' _)> */
		nil,
		/* 34 Subtract <- <(_ '-' _)> */
		nil,
		/* 35 BitwiseAnd <- <(_ '&' _)> */
		nil,
		/* 36 BitwiseOr <- <(_ '|' _)> */
		nil,
		/* 37 BitwiseNot <- <(_ '~' _)> */
		nil,
		/* 38 BitwiseXor <- <(_ '^' _)> */
		nil,
		/* 39 MatchOperator <- <(Match / Unmatch)> */
		nil,
		/* 40 Unmatch <- <(_ ('!' '~') _)> */
		nil,
		/* 41 Match <- <(_ ('=' '~') _)> */
		nil,
		/* 42 AssignmentOperator <- <(_ (AssignEq / StarEq / DivEq / PlusEq / MinusEq / AndEq / OrEq / Append) _)> */
		nil,
		/* 43 AssignEq <- <(_ '=' _)> */
		nil,
		/* 44 StarEq <- <(_ ('*' '=') _)> */
		nil,
		/* 45 DivEq <- <(_ ('/' '=') _)> */
		nil,
		/* 46 PlusEq <- <(_ ('+' '=') _)> */
		nil,
		/* 47 MinusEq <- <(_ ('-' '=') _)> */
		nil,
		/* 48 AndEq <- <(_ ('&' '=') _)> */
		nil,
		/* 49 OrEq <- <(_ ('|' '=') _)> */
		nil,
		/* 50 Append <- <(_ ('<' '<') _)> */
		nil,
		/* 51 ComparisonOperator <- <(_ (Equality / NonEquality / GreaterEqual / LessEqual / GreaterThan / LessThan / Membership / NonMembership) _)> */
		nil,
		/* 52 Equality <- <(_ ('=' '=') _)> */
		nil,
		/* 53 NonEquality <- <(_ ('!' '=') _)> */
		nil,
		/* 54 GreaterThan <- <(_ '>' _)> */
		nil,
		/* 55 GreaterEqual <- <(_ ('>' '=') _)> */
		nil,
		/* 56 LessEqual <- <(_ ('<' '=') _)> */
		nil,
		/* 57 LessThan <- <(_ '<' _)> */
		nil,
		/* 58 Membership <- <(_ ('i' 'n') _)> */
		nil,
		/* 59 NonMembership <- <(_ ('n' 'o' 't') __ ('i' 'n') _)> */
		nil,
		/* 60 Variable <- <(('$' VariableNameSequence) / SKIPVAR)> */
		func() bool {
			position95, tokenIndex95 := position, tokenIndex
			{
				position96 := position
				{
					position97, tokenIndex97 := position, tokenIndex
					if buffer[position] != rune('$') {
						goto l98
					}
					position++
					{
						position99 := position
					l100:
						{
							position101, tokenIndex101 := position, tokenIndex
							if !_rules[ruleVariableName]() {
								goto l101
							}
							{
								position102 := position
								if buffer[position] != rune('.') {
									goto l101
								}
								position++
								add(ruleDOT, position102)
							}
							goto l100
						l101:
							position, tokenIndex = position101, tokenIndex101
						}
						if !_rules[ruleVariableName]() {
							goto l98
						}
						add(ruleVariableNameSequence, position99)
					}
					goto l97
				l98:
					position, tokenIndex = position97, tokenIndex97
					{
						position103 := position
						if !_rules[rule_]() {
							goto l95
						}
						if buffer[position] != rune('_') {
							goto l95
						}
						position++
						if !_rules[rule_]() {
							goto l95
						}
						add(ruleSKIPVAR, position103)
					}
				}
			l97:
				add(ruleVariable, position96)
			}
			return true
		l95:
			position, tokenIndex = position95, tokenIndex95
			return false
		},
		/* 61 VariableNameSequence <- <((VariableName DOT)* VariableName)> */
		nil,
		/* 62 VariableName <- <(Identifier ('[' _ VariableIndex _ ']')?)> */
		func() bool {
			position105, tokenIndex105 := position, tokenIndex
			{
				position106 := position
				if !_rules[ruleIdentifier]() {
					goto l105
				}
				{
					position107, tokenIndex107 := position, tokenIndex
					if buffer[position] != rune('[') {
						goto l107
					}
					position++
					if !_rules[rule_]() {
						goto l107
					}
					{
						position109 := position
						if !_rules[ruleExpression]() {
							goto l107
						}
						add(ruleVariableIndex, position109)
					}
					if !_rules[rule_]() {
						goto l107
					}
					if buffer[position] != rune(']') {
						goto l107
					}
					position++
					goto l108
				l107:
					position, tokenIndex = position107, tokenIndex107
				}
			l108:
				add(ruleVariableName, position106)
			}
			return true
		l105:
			position, tokenIndex = position105, tokenIndex105
			return false
		},
		/* 63 VariableIndex <- <Expression> */
		nil,
		/* 64 Block <- <(_ (Comment / FlowControlWord / StatementBlock) SEMI? _)> */
		func() bool {
			position111, tokenIndex111 := position, tokenIndex
			{
				position112 := position
				if !_rules[rule_]() {
					goto l111
				}
				{
					position113, tokenIndex113 := position, tokenIndex
					{
						position115 := position
						if !_rules[rule_]() {
							goto l114
						}
						if buffer[position] != rune('#') {
							goto l114
						}
						position++
					l116:
						{
							position117, tokenIndex117 := position, tokenIndex
							{
								position118, tokenIndex118 := position, tokenIndex
								if buffer[position] != rune('\n') {
									goto l118
								}
								position++
								goto l117
							l118:
								position, tokenIndex = position118, tokenIndex118
							}
							if !matchDot() {
								goto l117
							}
							goto l116
						l117:
							position, tokenIndex = position117, tokenIndex117
						}
						add(ruleComment, position115)
					}
					goto l113
				l114:
					position, tokenIndex = position113, tokenIndex113
					{
						position120 := position
						{
							position121, tokenIndex121 := position, tokenIndex
							{
								position123 := position
								{
									position124 := position
									if !_rules[rule_]() {
										goto l122
									}
									if buffer[position] != rune('b') {
										goto l122
									}
									position++
									if buffer[position] != rune('r') {
										goto l122
									}
									position++
									if buffer[position] != rune('e') {
										goto l122
									}
									position++
									if buffer[position] != rune('a') {
										goto l122
									}
									position++
									if buffer[position] != rune('k') {
										goto l122
									}
									position++
									if !_rules[rule_]() {
										goto l122
									}
									add(ruleBREAK, position124)
								}
								{
									position125, tokenIndex125 := position, tokenIndex
									if !_rules[rulePositiveInteger]() {
										goto l125
									}
									goto l126
								l125:
									position, tokenIndex = position125, tokenIndex125
								}
							l126:
								add(ruleFlowControlBreak, position123)
							}
							goto l121
						l122:
							position, tokenIndex = position121, tokenIndex121
							{
								position127 := position
								{
									position128 := position
									if !_rules[rule_]() {
										goto l119
									}
									if buffer[position] != rune('c') {
										goto l119
									}
									position++
									if buffer[position] != rune('o') {
										goto l119
									}
									position++
									if buffer[position] != rune('n') {
										goto l119
									}
									position++
									if buffer[position] != rune('t') {
										goto l119
									}
									position++
									if buffer[position] != rune('i') {
										goto l119
									}
									position++
									if buffer[position] != rune('n') {
										goto l119
									}
									position++
									if buffer[position] != rune('u') {
										goto l119
									}
									position++
									if buffer[position] != rune('e') {
										goto l119
									}
									position++
									if !_rules[rule_]() {
										goto l119
									}
									add(ruleCONT, position128)
								}
								{
									position129, tokenIndex129 := position, tokenIndex
									if !_rules[rulePositiveInteger]() {
										goto l129
									}
									goto l130
								l129:
									position, tokenIndex = position129, tokenIndex129
								}
							l130:
								add(ruleFlowControlContinue, position127)
							}
						}
					l121:
						add(ruleFlowControlWord, position120)
					}
					goto l113
				l119:
					position, tokenIndex = position113, tokenIndex113
					{
						position131 := position
						{
							position132, tokenIndex132 := position, tokenIndex
							{
								position134 := position
								if !_rules[ruleSEMI]() {
									goto l133
								}
								add(ruleNOOP, position134)
							}
							goto l132
						l133:
							position, tokenIndex = position132, tokenIndex132
							if !_rules[ruleAssignment]() {
								goto l135
							}
							goto l132
						l135:
							position, tokenIndex = position132, tokenIndex132
							{
								position137 := position
								{
									position138, tokenIndex138 := position, tokenIndex
									{
										position140 := position
										{
											position141 := position
											if !_rules[rule_]() {
												goto l139
											}
											if buffer[position] != rune('u') {
												goto l139
											}
											position++
											if buffer[position] != rune('n') {
												goto l139
											}
											position++
											if buffer[position] != rune('s') {
												goto l139
											}
											position++
											if buffer[position] != rune('e') {
												goto l139
											}
											position++
											if buffer[position] != rune('t') {
												goto l139
											}
											position++
											if !_rules[rule__]() {
												goto l139
											}
											add(ruleUNSET, position141)
										}
										if !_rules[ruleVariableSequence]() {
											goto l139
										}
										add(ruleDirectiveUnset, position140)
									}
									goto l138
								l139:
									position, tokenIndex = position138, tokenIndex138
									{
										position143 := position
										{
											position144 := position
											if !_rules[rule_]() {
												goto l142
											}
											if buffer[position] != rune('i') {
												goto l142
											}
											position++
											if buffer[position] != rune('n') {
												goto l142
											}
											position++
											if buffer[position] != rune('c') {
												goto l142
											}
											position++
											if buffer[position] != rune('l') {
												goto l142
											}
											position++
											if buffer[position] != rune('u') {
												goto l142
											}
											position++
											if buffer[position] != rune('d') {
												goto l142
											}
											position++
											if buffer[position] != rune('e') {
												goto l142
											}
											position++
											if !_rules[rule__]() {
												goto l142
											}
											add(ruleINCLUDE, position144)
										}
										if !_rules[ruleString]() {
											goto l142
										}
										add(ruleDirectiveInclude, position143)
									}
									goto l138
								l142:
									position, tokenIndex = position138, tokenIndex138
									{
										position145 := position
										{
											position146 := position
											if !_rules[rule_]() {
												goto l136
											}
											if buffer[position] != rune('d') {
												goto l136
											}
											position++
											if buffer[position] != rune('e') {
												goto l136
											}
											position++
											if buffer[position] != rune('c') {
												goto l136
											}
											position++
											if buffer[position] != rune('l') {
												goto l136
											}
											position++
											if buffer[position] != rune('a') {
												goto l136
											}
											position++
											if buffer[position] != rune('r') {
												goto l136
											}
											position++
											if buffer[position] != rune('e') {
												goto l136
											}
											position++
											if !_rules[rule__]() {
												goto l136
											}
											add(ruleDECLARE, position146)
										}
										if !_rules[ruleVariableSequence]() {
											goto l136
										}
										add(ruleDirectiveDeclare, position145)
									}
								}
							l138:
								add(ruleDirective, position137)
							}
							goto l132
						l136:
							position, tokenIndex = position132, tokenIndex132
							{
								position148 := position
								if !_rules[ruleIfStanza]() {
									goto l147
								}
							l149:
								{
									position150, tokenIndex150 := position, tokenIndex
									{
										position151 := position
										if !_rules[ruleELSE]() {
											goto l150
										}
										if !_rules[ruleIfStanza]() {
											goto l150
										}
										add(ruleElseIfStanza, position151)
									}
									goto l149
								l150:
									position, tokenIndex = position150, tokenIndex150
								}
								{
									position152, tokenIndex152 := position, tokenIndex
									{
										position154 := position
										if !_rules[ruleELSE]() {
											goto l152
										}
										if !_rules[ruleOPEN]() {
											goto l152
										}
									l155:
										{
											position156, tokenIndex156 := position, tokenIndex
											if !_rules[ruleBlock]() {
												goto l156
											}
											goto l155
										l156:
											position, tokenIndex = position156, tokenIndex156
										}
										if !_rules[ruleCLOSE]() {
											goto l152
										}
										add(ruleElseStanza, position154)
									}
									goto l153
								l152:
									position, tokenIndex = position152, tokenIndex152
								}
							l153:
								add(ruleConditional, position148)
							}
							goto l132
						l147:
							position, tokenIndex = position132, tokenIndex132
							{
								position158 := position
								{
									position159 := position
									if !_rules[rule_]() {
										goto l157
									}
									if buffer[position] != rune('l') {
										goto l157
									}
									position++
									if buffer[position] != rune('o') {
										goto l157
									}
									position++
									if buffer[position] != rune('o') {
										goto l157
									}
									position++
									if buffer[position] != rune('p') {
										goto l157
									}
									position++
									if !_rules[rule_]() {
										goto l157
									}
									add(ruleLOOP, position159)
								}
								{
									position160, tokenIndex160 := position, tokenIndex
									if !_rules[ruleOPEN]() {
										goto l161
									}
								l162:
									{
										position163, tokenIndex163 := position, tokenIndex
										if !_rules[ruleBlock]() {
											goto l163
										}
										goto l162
									l163:
										position, tokenIndex = position163, tokenIndex163
									}
									if !_rules[ruleCLOSE]() {
										goto l161
									}
									goto l160
								l161:
									position, tokenIndex = position160, tokenIndex160
									{
										position165 := position
										{
											position166 := position
											if !_rules[rule_]() {
												goto l164
											}
											if buffer[position] != rune('c') {
												goto l164
											}
											position++
											if buffer[position] != rune('o') {
												goto l164
											}
											position++
											if buffer[position] != rune('u') {
												goto l164
											}
											position++
											if buffer[position] != rune('n') {
												goto l164
											}
											position++
											if buffer[position] != rune('t') {
												goto l164
											}
											position++
											if !_rules[rule_]() {
												goto l164
											}
											add(ruleCOUNT, position166)
										}
										{
											position167, tokenIndex167 := position, tokenIndex
											if !_rules[ruleInteger]() {
												goto l168
											}
											goto l167
										l168:
											position, tokenIndex = position167, tokenIndex167
											if !_rules[ruleVariable]() {
												goto l164
											}
										}
									l167:
										add(ruleLoopConditionFixedLength, position165)
									}
									if !_rules[ruleOPEN]() {
										goto l164
									}
								l169:
									{
										position170, tokenIndex170 := position, tokenIndex
										if !_rules[ruleBlock]() {
											goto l170
										}
										goto l169
									l170:
										position, tokenIndex = position170, tokenIndex170
									}
									if !_rules[ruleCLOSE]() {
										goto l164
									}
									goto l160
								l164:
									position, tokenIndex = position160, tokenIndex160
									{
										position172 := position
										{
											position173 := position
											if !_rules[ruleVariableSequence]() {
												goto l171
											}
											add(ruleLoopIterableLHS, position173)
										}
										{
											position174 := position
											if !_rules[rule__]() {
												goto l171
											}
											if buffer[position] != rune('i') {
												goto l171
											}
											position++
											if buffer[position] != rune('n') {
												goto l171
											}
											position++
											if !_rules[rule__]() {
												goto l171
											}
											add(ruleIN, position174)
										}
										{
											position175 := position
											{
												position176, tokenIndex176 := position, tokenIndex
												if !_rules[ruleCommand]() {
													goto l177
												}
												goto l176
											l177:
												position, tokenIndex = position176, tokenIndex176
												if !_rules[ruleVariable]() {
													goto l171
												}
											}
										l176:
											add(ruleLoopIterableRHS, position175)
										}
										add(ruleLoopConditionIterable, position172)
									}
									if !_rules[ruleOPEN]() {
										goto l171
									}
								l178:
									{
										position179, tokenIndex179 := position, tokenIndex
										if !_rules[ruleBlock]() {
											goto l179
										}
										goto l178
									l179:
										position, tokenIndex = position179, tokenIndex179
									}
									if !_rules[ruleCLOSE]() {
										goto l171
									}
									goto l160
								l171:
									position, tokenIndex = position160, tokenIndex160
									{
										position181 := position
										if !_rules[ruleCommand]() {
											goto l180
										}
										if !_rules[ruleSEMI]() {
											goto l180
										}
										if !_rules[ruleConditionalExpression]() {
											goto l180
										}
										if !_rules[ruleSEMI]() {
											goto l180
										}
										if !_rules[ruleCommand]() {
											goto l180
										}
										add(ruleLoopConditionBounded, position181)
									}
									if !_rules[ruleOPEN]() {
										goto l180
									}
								l182:
									{
										position183, tokenIndex183 := position, tokenIndex
										if !_rules[ruleBlock]() {
											goto l183
										}
										goto l182
									l183:
										position, tokenIndex = position183, tokenIndex183
									}
									if !_rules[ruleCLOSE]() {
										goto l180
									}
									goto l160
								l180:
									position, tokenIndex = position160, tokenIndex160
									{
										position184 := position
										if !_rules[ruleConditionalExpression]() {
											goto l157
										}
										add(ruleLoopConditionTruthy, position184)
									}
									if !_rules[ruleOPEN]() {
										goto l157
									}
								l185:
									{
										position186, tokenIndex186 := position, tokenIndex
										if !_rules[ruleBlock]() {
											goto l186
										}
										goto l185
									l186:
										position, tokenIndex = position186, tokenIndex186
									}
									if !_rules[ruleCLOSE]() {
										goto l157
									}
								}
							l160:
								add(ruleLoop, position158)
							}
							goto l132
						l157:
							position, tokenIndex = position132, tokenIndex132
							if !_rules[ruleCommand]() {
								goto l111
							}
						}
					l132:
						add(ruleStatementBlock, position131)
					}
				}
			l113:
				{
					position187, tokenIndex187 := position, tokenIndex
					if !_rules[ruleSEMI]() {
						goto l187
					}
					goto l188
				l187:
					position, tokenIndex = position187, tokenIndex187
				}
			l188:
				if !_rules[rule_]() {
					goto l111
				}
				add(ruleBlock, position112)
			}
			return true
		l111:
			position, tokenIndex = position111, tokenIndex111
			return false
		},
		/* 65 FlowControlWord <- <(FlowControlBreak / FlowControlContinue)> */
		nil,
		/* 66 FlowControlBreak <- <(BREAK PositiveInteger?)> */
		nil,
		/* 67 FlowControlContinue <- <(CONT PositiveInteger?)> */
		nil,
		/* 68 StatementBlock <- <(NOOP / Assignment / Directive / Conditional / Loop / Command)> */
		nil,
		/* 69 Assignment <- <(AssignmentLHS AssignmentOperator AssignmentRHS)> */
		func() bool {
			position193, tokenIndex193 := position, tokenIndex
			{
				position194 := position
				{
					position195 := position
					if !_rules[ruleVariableSequence]() {
						goto l193
					}
					add(ruleAssignmentLHS, position195)
				}
				{
					position196 := position
					if !_rules[rule_]() {
						goto l193
					}
					{
						position197, tokenIndex197 := position, tokenIndex
						{
							position199 := position
							if !_rules[rule_]() {
								goto l198
							}
							if buffer[position] != rune('=') {
								goto l198
							}
							position++
							if !_rules[rule_]() {
								goto l198
							}
							add(ruleAssignEq, position199)
						}
						goto l197
					l198:
						position, tokenIndex = position197, tokenIndex197
						{
							position201 := position
							if !_rules[rule_]() {
								goto l200
							}
							if buffer[position] != rune('*') {
								goto l200
							}
							position++
							if buffer[position] != rune('=') {
								goto l200
							}
							position++
							if !_rules[rule_]() {
								goto l200
							}
							add(ruleStarEq, position201)
						}
						goto l197
					l200:
						position, tokenIndex = position197, tokenIndex197
						{
							position203 := position
							if !_rules[rule_]() {
								goto l202
							}
							if buffer[position] != rune('/') {
								goto l202
							}
							position++
							if buffer[position] != rune('=') {
								goto l202
							}
							position++
							if !_rules[rule_]() {
								goto l202
							}
							add(ruleDivEq, position203)
						}
						goto l197
					l202:
						position, tokenIndex = position197, tokenIndex197
						{
							position205 := position
							if !_rules[rule_]() {
								goto l204
							}
							if buffer[position] != rune('+') {
								goto l204
							}
							position++
							if buffer[position] != rune('=') {
								goto l204
							}
							position++
							if !_rules[rule_]() {
								goto l204
							}
							add(rulePlusEq, position205)
						}
						goto l197
					l204:
						position, tokenIndex = position197, tokenIndex197
						{
							position207 := position
							if !_rules[rule_]() {
								goto l206
							}
							if buffer[position] != rune('-') {
								goto l206
							}
							position++
							if buffer[position] != rune('=') {
								goto l206
							}
							position++
							if !_rules[rule_]() {
								goto l206
							}
							add(ruleMinusEq, position207)
						}
						goto l197
					l206:
						position, tokenIndex = position197, tokenIndex197
						{
							position209 := position
							if !_rules[rule_]() {
								goto l208
							}
							if buffer[position] != rune('&') {
								goto l208
							}
							position++
							if buffer[position] != rune('=') {
								goto l208
							}
							position++
							if !_rules[rule_]() {
								goto l208
							}
							add(ruleAndEq, position209)
						}
						goto l197
					l208:
						position, tokenIndex = position197, tokenIndex197
						{
							position211 := position
							if !_rules[rule_]() {
								goto l210
							}
							if buffer[position] != rune('|') {
								goto l210
							}
							position++
							if buffer[position] != rune('=') {
								goto l210
							}
							position++
							if !_rules[rule_]() {
								goto l210
							}
							add(ruleOrEq, position211)
						}
						goto l197
					l210:
						position, tokenIndex = position197, tokenIndex197
						{
							position212 := position
							if !_rules[rule_]() {
								goto l193
							}
							if buffer[position] != rune('<') {
								goto l193
							}
							position++
							if buffer[position] != rune('<') {
								goto l193
							}
							position++
							if !_rules[rule_]() {
								goto l193
							}
							add(ruleAppend, position212)
						}
					}
				l197:
					if !_rules[rule_]() {
						goto l193
					}
					add(ruleAssignmentOperator, position196)
				}
				{
					position213 := position
					if !_rules[ruleExpressionSequence]() {
						goto l193
					}
					add(ruleAssignmentRHS, position213)
				}
				add(ruleAssignment, position194)
			}
			return true
		l193:
			position, tokenIndex = position193, tokenIndex193
			return false
		},
		/* 70 AssignmentLHS <- <VariableSequence> */
		nil,
		/* 71 AssignmentRHS <- <ExpressionSequence> */
		nil,
		/* 72 VariableSequence <- <((Variable COMMA)* Variable)> */
		func() bool {
			position216, tokenIndex216 := position, tokenIndex
			{
				position217 := position
			l218:
				{
					position219, tokenIndex219 := position, tokenIndex
					if !_rules[ruleVariable]() {
						goto l219
					}
					if !_rules[ruleCOMMA]() {
						goto l219
					}
					goto l218
				l219:
					position, tokenIndex = position219, tokenIndex219
				}
				if !_rules[ruleVariable]() {
					goto l216
				}
				add(ruleVariableSequence, position217)
			}
			return true
		l216:
			position, tokenIndex = position216, tokenIndex216
			return false
		},
		/* 73 ExpressionSequence <- <((Expression COMMA)* Expression)> */
		func() bool {
			position220, tokenIndex220 := position, tokenIndex
			{
				position221 := position
			l222:
				{
					position223, tokenIndex223 := position, tokenIndex
					if !_rules[ruleExpression]() {
						goto l223
					}
					if !_rules[ruleCOMMA]() {
						goto l223
					}
					goto l222
				l223:
					position, tokenIndex = position223, tokenIndex223
				}
				if !_rules[ruleExpression]() {
					goto l220
				}
				add(ruleExpressionSequence, position221)
			}
			return true
		l220:
			position, tokenIndex = position220, tokenIndex220
			return false
		},
		/* 74 Expression <- <(_ ExpressionLHS ExpressionRHS? _)> */
		func() bool {
			position224, tokenIndex224 := position, tokenIndex
			{
				position225 := position
				if !_rules[rule_]() {
					goto l224
				}
				{
					position226 := position
					{
						position227 := position
						{
							position228, tokenIndex228 := position, tokenIndex
							{
								position230 := position
								{
									position231, tokenIndex231 := position, tokenIndex
									if !_rules[ruleArray]() {
										goto l232
									}
									goto l231
								l232:
									position, tokenIndex = position231, tokenIndex231
									if !_rules[ruleObject]() {
										goto l233
									}
									goto l231
								l233:
									position, tokenIndex = position231, tokenIndex231
									if !_rules[ruleRegularExpression]() {
										goto l234
									}
									goto l231
								l234:
									position, tokenIndex = position231, tokenIndex231
									if !_rules[ruleScalarType]() {
										goto l229
									}
								}
							l231:
								add(ruleType, position230)
							}
							goto l228
						l229:
							position, tokenIndex = position228, tokenIndex228
							if !_rules[ruleVariable]() {
								goto l224
							}
						}
					l228:
						add(ruleValueYielding, position227)
					}
					add(ruleExpressionLHS, position226)
				}
				{
					position235, tokenIndex235 := position, tokenIndex
					{
						position237 := position
						{
							position238 := position
							if !_rules[rule_]() {
								goto l235
							}
							{
								position239, tokenIndex239 := position, tokenIndex
								{
									position241 := position
									if !_rules[rule_]() {
										goto l240
									}
									if buffer[position] != rune('*') {
										goto l240
									}
									position++
									if buffer[position] != rune('*') {
										goto l240
									}
									position++
									if !_rules[rule_]() {
										goto l240
									}
									add(ruleExponentiate, position241)
								}
								goto l239
							l240:
								position, tokenIndex = position239, tokenIndex239
								{
									position243 := position
									if !_rules[rule_]() {
										goto l242
									}
									if buffer[position] != rune('*') {
										goto l242
									}
									position++
									if !_rules[rule_]() {
										goto l242
									}
									add(ruleMultiply, position243)
								}
								goto l239
							l242:
								position, tokenIndex = position239, tokenIndex239
								{
									position245 := position
									if !_rules[rule_]() {
										goto l244
									}
									if buffer[position] != rune('/') {
										goto l244
									}
									position++
									if !_rules[rule_]() {
										goto l244
									}
									add(ruleDivide, position245)
								}
								goto l239
							l244:
								position, tokenIndex = position239, tokenIndex239
								{
									position247 := position
									if !_rules[rule_]() {
										goto l246
									}
									if buffer[position] != rune('%') {
										goto l246
									}
									position++
									if !_rules[rule_]() {
										goto l246
									}
									add(ruleModulus, position247)
								}
								goto l239
							l246:
								position, tokenIndex = position239, tokenIndex239
								{
									position249 := position
									if !_rules[rule_]() {
										goto l248
									}
									if buffer[position] != rune('+') {
										goto l248
									}
									position++
									if !_rules[rule_]() {
										goto l248
									}
									add(ruleAdd, position249)
								}
								goto l239
							l248:
								position, tokenIndex = position239, tokenIndex239
								{
									position251 := position
									if !_rules[rule_]() {
										goto l250
									}
									if buffer[position] != rune('-') {
										goto l250
									}
									position++
									if !_rules[rule_]() {
										goto l250
									}
									add(ruleSubtract, position251)
								}
								goto l239
							l250:
								position, tokenIndex = position239, tokenIndex239
								{
									position253 := position
									if !_rules[rule_]() {
										goto l252
									}
									if buffer[position] != rune('&') {
										goto l252
									}
									position++
									if !_rules[rule_]() {
										goto l252
									}
									add(ruleBitwiseAnd, position253)
								}
								goto l239
							l252:
								position, tokenIndex = position239, tokenIndex239
								{
									position255 := position
									if !_rules[rule_]() {
										goto l254
									}
									if buffer[position] != rune('|') {
										goto l254
									}
									position++
									if !_rules[rule_]() {
										goto l254
									}
									add(ruleBitwiseOr, position255)
								}
								goto l239
							l254:
								position, tokenIndex = position239, tokenIndex239
								{
									position257 := position
									if !_rules[rule_]() {
										goto l256
									}
									if buffer[position] != rune('~') {
										goto l256
									}
									position++
									if !_rules[rule_]() {
										goto l256
									}
									add(ruleBitwiseNot, position257)
								}
								goto l239
							l256:
								position, tokenIndex = position239, tokenIndex239
								{
									position258 := position
									if !_rules[rule_]() {
										goto l235
									}
									if buffer[position] != rune('^') {
										goto l235
									}
									position++
									if !_rules[rule_]() {
										goto l235
									}
									add(ruleBitwiseXor, position258)
								}
							}
						l239:
							if !_rules[rule_]() {
								goto l235
							}
							add(ruleOperator, position238)
						}
						if !_rules[ruleExpression]() {
							goto l235
						}
						add(ruleExpressionRHS, position237)
					}
					goto l236
				l235:
					position, tokenIndex = position235, tokenIndex235
				}
			l236:
				if !_rules[rule_]() {
					goto l224
				}
				add(ruleExpression, position225)
			}
			return true
		l224:
			position, tokenIndex = position224, tokenIndex224
			return false
		},
		/* 75 ExpressionLHS <- <ValueYielding> */
		nil,
		/* 76 ExpressionRHS <- <(Operator Expression)> */
		nil,
		/* 77 ValueYielding <- <(Type / Variable)> */
		nil,
		/* 78 Directive <- <(DirectiveUnset / DirectiveInclude / DirectiveDeclare)> */
		nil,
		/* 79 DirectiveUnset <- <(UNSET VariableSequence)> */
		nil,
		/* 80 DirectiveInclude <- <(INCLUDE String)> */
		nil,
		/* 81 DirectiveDeclare <- <(DECLARE VariableSequence)> */
		nil,
		/* 82 Command <- <(_ CommandName __ ((CommandFirstArg __ CommandSecondArg) / CommandFirstArg / CommandSecondArg)? (_ CommandResultAssignment)?)> */
		func() bool {
			position266, tokenIndex266 := position, tokenIndex
			{
				position267 := position
				if !_rules[rule_]() {
					goto l266
				}
				{
					position268 := position
					{
						position269, tokenIndex269 := position, tokenIndex
						if !_rules[ruleIdentifier]() {
							goto l269
						}
						{
							position271 := position
							if buffer[position] != rune(':') {
								goto l269
							}
							position++
							if buffer[position] != rune(':') {
								goto l269
							}
							position++
							add(ruleSCOPE, position271)
						}
						goto l270
					l269:
						position, tokenIndex = position269, tokenIndex269
					}
				l270:
					if !_rules[ruleIdentifier]() {
						goto l266
					}
					add(ruleCommandName, position268)
				}
				if !_rules[rule__]() {
					goto l266
				}
				{
					position272, tokenIndex272 := position, tokenIndex
					{
						position274, tokenIndex274 := position, tokenIndex
						if !_rules[ruleCommandFirstArg]() {
							goto l275
						}
						if !_rules[rule__]() {
							goto l275
						}
						if !_rules[ruleCommandSecondArg]() {
							goto l275
						}
						goto l274
					l275:
						position, tokenIndex = position274, tokenIndex274
						if !_rules[ruleCommandFirstArg]() {
							goto l276
						}
						goto l274
					l276:
						position, tokenIndex = position274, tokenIndex274
						if !_rules[ruleCommandSecondArg]() {
							goto l272
						}
					}
				l274:
					goto l273
				l272:
					position, tokenIndex = position272, tokenIndex272
				}
			l273:
				{
					position277, tokenIndex277 := position, tokenIndex
					if !_rules[rule_]() {
						goto l277
					}
					{
						position279 := position
						{
							position280 := position
							if !_rules[rule_]() {
								goto l277
							}
							if buffer[position] != rune('-') {
								goto l277
							}
							position++
							if buffer[position] != rune('>') {
								goto l277
							}
							position++
							if !_rules[rule_]() {
								goto l277
							}
							add(ruleASSIGN, position280)
						}
						if !_rules[ruleVariable]() {
							goto l277
						}
						add(ruleCommandResultAssignment, position279)
					}
					goto l278
				l277:
					position, tokenIndex = position277, tokenIndex277
				}
			l278:
				add(ruleCommand, position267)
			}
			return true
		l266:
			position, tokenIndex = position266, tokenIndex266
			return false
		},
		/* 83 CommandName <- <((Identifier SCOPE)? Identifier)> */
		nil,
		/* 84 CommandFirstArg <- <(Variable / ScalarType)> */
		func() bool {
			position282, tokenIndex282 := position, tokenIndex
			{
				position283 := position
				{
					position284, tokenIndex284 := position, tokenIndex
					if !_rules[ruleVariable]() {
						goto l285
					}
					goto l284
				l285:
					position, tokenIndex = position284, tokenIndex284
					if !_rules[ruleScalarType]() {
						goto l282
					}
				}
			l284:
				add(ruleCommandFirstArg, position283)
			}
			return true
		l282:
			position, tokenIndex = position282, tokenIndex282
			return false
		},
		/* 85 CommandSecondArg <- <Object> */
		func() bool {
			position286, tokenIndex286 := position, tokenIndex
			{
				position287 := position
				if !_rules[ruleObject]() {
					goto l286
				}
				add(ruleCommandSecondArg, position287)
			}
			return true
		l286:
			position, tokenIndex = position286, tokenIndex286
			return false
		},
		/* 86 CommandResultAssignment <- <(ASSIGN Variable)> */
		nil,
		/* 87 Conditional <- <(IfStanza ElseIfStanza* ElseStanza?)> */
		nil,
		/* 88 IfStanza <- <(IF ConditionalExpression OPEN Block* CLOSE)> */
		func() bool {
			position290, tokenIndex290 := position, tokenIndex
			{
				position291 := position
				{
					position292 := position
					if !_rules[rule_]() {
						goto l290
					}
					if buffer[position] != rune('i') {
						goto l290
					}
					position++
					if buffer[position] != rune('f') {
						goto l290
					}
					position++
					if !_rules[rule_]() {
						goto l290
					}
					add(ruleIF, position292)
				}
				if !_rules[ruleConditionalExpression]() {
					goto l290
				}
				if !_rules[ruleOPEN]() {
					goto l290
				}
			l293:
				{
					position294, tokenIndex294 := position, tokenIndex
					if !_rules[ruleBlock]() {
						goto l294
					}
					goto l293
				l294:
					position, tokenIndex = position294, tokenIndex294
				}
				if !_rules[ruleCLOSE]() {
					goto l290
				}
				add(ruleIfStanza, position291)
			}
			return true
		l290:
			position, tokenIndex = position290, tokenIndex290
			return false
		},
		/* 89 ElseIfStanza <- <(ELSE IfStanza)> */
		nil,
		/* 90 ElseStanza <- <(ELSE OPEN Block* CLOSE)> */
		nil,
		/* 91 Loop <- <(LOOP ((OPEN Block* CLOSE) / (LoopConditionFixedLength OPEN Block* CLOSE) / (LoopConditionIterable OPEN Block* CLOSE) / (LoopConditionBounded OPEN Block* CLOSE) / (LoopConditionTruthy OPEN Block* CLOSE)))> */
		nil,
		/* 92 LoopConditionFixedLength <- <(COUNT (Integer / Variable))> */
		nil,
		/* 93 LoopConditionIterable <- <(LoopIterableLHS IN LoopIterableRHS)> */
		nil,
		/* 94 LoopIterableLHS <- <VariableSequence> */
		nil,
		/* 95 LoopIterableRHS <- <(Command / Variable)> */
		nil,
		/* 96 LoopConditionBounded <- <(Command SEMI ConditionalExpression SEMI Command)> */
		nil,
		/* 97 LoopConditionTruthy <- <ConditionalExpression> */
		nil,
		/* 98 ConditionalExpression <- <(NOT? (ConditionWithAssignment / ConditionWithCommand / ConditionWithRegex / ConditionWithComparator))> */
		func() bool {
			position304, tokenIndex304 := position, tokenIndex
			{
				position305 := position
				{
					position306, tokenIndex306 := position, tokenIndex
					{
						position308 := position
						if !_rules[rule_]() {
							goto l306
						}
						if buffer[position] != rune('n') {
							goto l306
						}
						position++
						if buffer[position] != rune('o') {
							goto l306
						}
						position++
						if buffer[position] != rune('t') {
							goto l306
						}
						position++
						if !_rules[rule__]() {
							goto l306
						}
						add(ruleNOT, position308)
					}
					goto l307
				l306:
					position, tokenIndex = position306, tokenIndex306
				}
			l307:
				{
					position309, tokenIndex309 := position, tokenIndex
					{
						position311 := position
						if !_rules[ruleAssignment]() {
							goto l310
						}
						if !_rules[ruleSEMI]() {
							goto l310
						}
						if !_rules[ruleConditionalExpression]() {
							goto l310
						}
						add(ruleConditionWithAssignment, position311)
					}
					goto l309
				l310:
					position, tokenIndex = position309, tokenIndex309
					{
						position313 := position
						if !_rules[ruleCommand]() {
							goto l312
						}
						{
							position314, tokenIndex314 := position, tokenIndex
							if !_rules[ruleSEMI]() {
								goto l314
							}
							if !_rules[ruleConditionalExpression]() {
								goto l314
							}
							goto l315
						l314:
							position, tokenIndex = position314, tokenIndex314
						}
					l315:
						add(ruleConditionWithCommand, position313)
					}
					goto l309
				l312:
					position, tokenIndex = position309, tokenIndex309
					{
						position317 := position
						if !_rules[ruleExpression]() {
							goto l316
						}
						{
							position318 := position
							{
								position319, tokenIndex319 := position, tokenIndex
								{
									position321 := position
									if !_rules[rule_]() {
										goto l320
									}
									if buffer[position] != rune('=') {
										goto l320
									}
									position++
									if buffer[position] != rune('~') {
										goto l320
									}
									position++
									if !_rules[rule_]() {
										goto l320
									}
									add(ruleMatch, position321)
								}
								goto l319
							l320:
								position, tokenIndex = position319, tokenIndex319
								{
									position322 := position
									if !_rules[rule_]() {
										goto l316
									}
									if buffer[position] != rune('!') {
										goto l316
									}
									position++
									if buffer[position] != rune('~') {
										goto l316
									}
									position++
									if !_rules[rule_]() {
										goto l316
									}
									add(ruleUnmatch, position322)
								}
							}
						l319:
							add(ruleMatchOperator, position318)
						}
						if !_rules[ruleRegularExpression]() {
							goto l316
						}
						add(ruleConditionWithRegex, position317)
					}
					goto l309
				l316:
					position, tokenIndex = position309, tokenIndex309
					{
						position323 := position
						{
							position324 := position
							if !_rules[ruleExpression]() {
								goto l304
							}
							add(ruleConditionWithComparatorLHS, position324)
						}
						{
							position325, tokenIndex325 := position, tokenIndex
							{
								position327 := position
								{
									position328 := position
									if !_rules[rule_]() {
										goto l325
									}
									{
										position329, tokenIndex329 := position, tokenIndex
										{
											position331 := position
											if !_rules[rule_]() {
												goto l330
											}
											if buffer[position] != rune('=') {
												goto l330
											}
											position++
											if buffer[position] != rune('=') {
												goto l330
											}
											position++
											if !_rules[rule_]() {
												goto l330
											}
											add(ruleEquality, position331)
										}
										goto l329
									l330:
										position, tokenIndex = position329, tokenIndex329
										{
											position333 := position
											if !_rules[rule_]() {
												goto l332
											}
											if buffer[position] != rune('!') {
												goto l332
											}
											position++
											if buffer[position] != rune('=') {
												goto l332
											}
											position++
											if !_rules[rule_]() {
												goto l332
											}
											add(ruleNonEquality, position333)
										}
										goto l329
									l332:
										position, tokenIndex = position329, tokenIndex329
										{
											position335 := position
											if !_rules[rule_]() {
												goto l334
											}
											if buffer[position] != rune('>') {
												goto l334
											}
											position++
											if buffer[position] != rune('=') {
												goto l334
											}
											position++
											if !_rules[rule_]() {
												goto l334
											}
											add(ruleGreaterEqual, position335)
										}
										goto l329
									l334:
										position, tokenIndex = position329, tokenIndex329
										{
											position337 := position
											if !_rules[rule_]() {
												goto l336
											}
											if buffer[position] != rune('<') {
												goto l336
											}
											position++
											if buffer[position] != rune('=') {
												goto l336
											}
											position++
											if !_rules[rule_]() {
												goto l336
											}
											add(ruleLessEqual, position337)
										}
										goto l329
									l336:
										position, tokenIndex = position329, tokenIndex329
										{
											position339 := position
											if !_rules[rule_]() {
												goto l338
											}
											if buffer[position] != rune('>') {
												goto l338
											}
											position++
											if !_rules[rule_]() {
												goto l338
											}
											add(ruleGreaterThan, position339)
										}
										goto l329
									l338:
										position, tokenIndex = position329, tokenIndex329
										{
											position341 := position
											if !_rules[rule_]() {
												goto l340
											}
											if buffer[position] != rune('<') {
												goto l340
											}
											position++
											if !_rules[rule_]() {
												goto l340
											}
											add(ruleLessThan, position341)
										}
										goto l329
									l340:
										position, tokenIndex = position329, tokenIndex329
										{
											position343 := position
											if !_rules[rule_]() {
												goto l342
											}
											if buffer[position] != rune('i') {
												goto l342
											}
											position++
											if buffer[position] != rune('n') {
												goto l342
											}
											position++
											if !_rules[rule_]() {
												goto l342
											}
											add(ruleMembership, position343)
										}
										goto l329
									l342:
										position, tokenIndex = position329, tokenIndex329
										{
											position344 := position
											if !_rules[rule_]() {
												goto l325
											}
											if buffer[position] != rune('n') {
												goto l325
											}
											position++
											if buffer[position] != rune('o') {
												goto l325
											}
											position++
											if buffer[position] != rune('t') {
												goto l325
											}
											position++
											if !_rules[rule__]() {
												goto l325
											}
											if buffer[position] != rune('i') {
												goto l325
											}
											position++
											if buffer[position] != rune('n') {
												goto l325
											}
											position++
											if !_rules[rule_]() {
												goto l325
											}
											add(ruleNonMembership, position344)
										}
									}
								l329:
									if !_rules[rule_]() {
										goto l325
									}
									add(ruleComparisonOperator, position328)
								}
								if !_rules[ruleExpression]() {
									goto l325
								}
								add(ruleConditionWithComparatorRHS, position327)
							}
							goto l326
						l325:
							position, tokenIndex = position325, tokenIndex325
						}
					l326:
						add(ruleConditionWithComparator, position323)
					}
				}
			l309:
				add(ruleConditionalExpression, position305)
			}
			return true
		l304:
			position, tokenIndex = position304, tokenIndex304
			return false
		},
		/* 99 ConditionWithAssignment <- <(Assignment SEMI ConditionalExpression)> */
		nil,
		/* 100 ConditionWithCommand <- <(Command (SEMI ConditionalExpression)?)> */
		nil,
		/* 101 ConditionWithRegex <- <(Expression MatchOperator RegularExpression)> */
		nil,
		/* 102 ConditionWithComparator <- <(ConditionWithComparatorLHS ConditionWithComparatorRHS?)> */
		nil,
		/* 103 ConditionWithComparatorLHS <- <Expression> */
		nil,
		/* 104 ConditionWithComparatorRHS <- <(ComparisonOperator Expression)> */
		nil,
		/* 105 ScalarType <- <(Boolean / Float / Integer / String / NullValue)> */
		func() bool {
			position351, tokenIndex351 := position, tokenIndex
			{
				position352 := position
				{
					position353, tokenIndex353 := position, tokenIndex
					{
						position355 := position
						{
							position356, tokenIndex356 := position, tokenIndex
							if buffer[position] != rune('t') {
								goto l357
							}
							position++
							if buffer[position] != rune('r') {
								goto l357
							}
							position++
							if buffer[position] != rune('u') {
								goto l357
							}
							position++
							if buffer[position] != rune('e') {
								goto l357
							}
							position++
							goto l356
						l357:
							position, tokenIndex = position356, tokenIndex356
							if buffer[position] != rune('f') {
								goto l354
							}
							position++
							if buffer[position] != rune('a') {
								goto l354
							}
							position++
							if buffer[position] != rune('l') {
								goto l354
							}
							position++
							if buffer[position] != rune('s') {
								goto l354
							}
							position++
							if buffer[position] != rune('e') {
								goto l354
							}
							position++
						}
					l356:
						add(ruleBoolean, position355)
					}
					goto l353
				l354:
					position, tokenIndex = position353, tokenIndex353
					{
						position359 := position
						if !_rules[ruleInteger]() {
							goto l358
						}
						{
							position360, tokenIndex360 := position, tokenIndex
							if buffer[position] != rune('.') {
								goto l360
							}
							position++
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l360
							}
							position++
						l362:
							{
								position363, tokenIndex363 := position, tokenIndex
								if c := buffer[position]; c < rune('0') || c > rune('9') {
									goto l363
								}
								position++
								goto l362
							l363:
								position, tokenIndex = position363, tokenIndex363
							}
							goto l361
						l360:
							position, tokenIndex = position360, tokenIndex360
						}
					l361:
						add(ruleFloat, position359)
					}
					goto l353
				l358:
					position, tokenIndex = position353, tokenIndex353
					if !_rules[ruleInteger]() {
						goto l364
					}
					goto l353
				l364:
					position, tokenIndex = position353, tokenIndex353
					if !_rules[ruleString]() {
						goto l365
					}
					goto l353
				l365:
					position, tokenIndex = position353, tokenIndex353
					{
						position366 := position
						if buffer[position] != rune('n') {
							goto l351
						}
						position++
						if buffer[position] != rune('u') {
							goto l351
						}
						position++
						if buffer[position] != rune('l') {
							goto l351
						}
						position++
						if buffer[position] != rune('l') {
							goto l351
						}
						position++
						add(ruleNullValue, position366)
					}
				}
			l353:
				add(ruleScalarType, position352)
			}
			return true
		l351:
			position, tokenIndex = position351, tokenIndex351
			return false
		},
		/* 106 Identifier <- <(([a-z] / [A-Z] / '_') ([a-z] / [A-Z] / ([0-9] / [0-9]) / '_')*)> */
		func() bool {
			position367, tokenIndex367 := position, tokenIndex
			{
				position368 := position
				{
					position369, tokenIndex369 := position, tokenIndex
					if c := buffer[position]; c < rune('a') || c > rune('z') {
						goto l370
					}
					position++
					goto l369
				l370:
					position, tokenIndex = position369, tokenIndex369
					if c := buffer[position]; c < rune('A') || c > rune('Z') {
						goto l371
					}
					position++
					goto l369
				l371:
					position, tokenIndex = position369, tokenIndex369
					if buffer[position] != rune('_') {
						goto l367
					}
					position++
				}
			l369:
			l372:
				{
					position373, tokenIndex373 := position, tokenIndex
					{
						position374, tokenIndex374 := position, tokenIndex
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l375
						}
						position++
						goto l374
					l375:
						position, tokenIndex = position374, tokenIndex374
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l376
						}
						position++
						goto l374
					l376:
						position, tokenIndex = position374, tokenIndex374
						{
							position378, tokenIndex378 := position, tokenIndex
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l379
							}
							position++
							goto l378
						l379:
							position, tokenIndex = position378, tokenIndex378
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l377
							}
							position++
						}
					l378:
						goto l374
					l377:
						position, tokenIndex = position374, tokenIndex374
						if buffer[position] != rune('_') {
							goto l373
						}
						position++
					}
				l374:
					goto l372
				l373:
					position, tokenIndex = position373, tokenIndex373
				}
				add(ruleIdentifier, position368)
			}
			return true
		l367:
			position, tokenIndex = position367, tokenIndex367
			return false
		},
		/* 107 Float <- <(Integer ('.' [0-9]+)?)> */
		nil,
		/* 108 Boolean <- <(('t' 'r' 'u' 'e') / ('f' 'a' 'l' 's' 'e'))> */
		nil,
		/* 109 Integer <- <('-'? PositiveInteger)> */
		func() bool {
			position382, tokenIndex382 := position, tokenIndex
			{
				position383 := position
				{
					position384, tokenIndex384 := position, tokenIndex
					if buffer[position] != rune('-') {
						goto l384
					}
					position++
					goto l385
				l384:
					position, tokenIndex = position384, tokenIndex384
				}
			l385:
				if !_rules[rulePositiveInteger]() {
					goto l382
				}
				add(ruleInteger, position383)
			}
			return true
		l382:
			position, tokenIndex = position382, tokenIndex382
			return false
		},
		/* 110 PositiveInteger <- <[0-9]+> */
		func() bool {
			position386, tokenIndex386 := position, tokenIndex
			{
				position387 := position
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l386
				}
				position++
			l388:
				{
					position389, tokenIndex389 := position, tokenIndex
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l389
					}
					position++
					goto l388
				l389:
					position, tokenIndex = position389, tokenIndex389
				}
				add(rulePositiveInteger, position387)
			}
			return true
		l386:
			position, tokenIndex = position386, tokenIndex386
			return false
		},
		/* 111 String <- <(StringLiteral / StringInterpolated / Heredoc)> */
		func() bool {
			position390, tokenIndex390 := position, tokenIndex
			{
				position391 := position
				{
					position392, tokenIndex392 := position, tokenIndex
					{
						position394 := position
						if buffer[position] != rune('\'') {
							goto l393
						}
						position++
					l395:
						{
							position396, tokenIndex396 := position, tokenIndex
							{
								position397, tokenIndex397 := position, tokenIndex
								if buffer[position] != rune('\'') {
									goto l397
								}
								position++
								goto l396
							l397:
								position, tokenIndex = position397, tokenIndex397
							}
							if !matchDot() {
								goto l396
							}
							goto l395
						l396:
							position, tokenIndex = position396, tokenIndex396
						}
						if buffer[position] != rune('\'') {
							goto l393
						}
						position++
						add(ruleStringLiteral, position394)
					}
					goto l392
				l393:
					position, tokenIndex = position392, tokenIndex392
					{
						position399 := position
						if buffer[position] != rune('"') {
							goto l398
						}
						position++
					l400:
						{
							position401, tokenIndex401 := position, tokenIndex
							{
								position402, tokenIndex402 := position, tokenIndex
								if buffer[position] != rune('"') {
									goto l402
								}
								position++
								goto l401
							l402:
								position, tokenIndex = position402, tokenIndex402
							}
							if !matchDot() {
								goto l401
							}
							goto l400
						l401:
							position, tokenIndex = position401, tokenIndex401
						}
						if buffer[position] != rune('"') {
							goto l398
						}
						position++
						add(ruleStringInterpolated, position399)
					}
					goto l392
				l398:
					position, tokenIndex = position392, tokenIndex392
					{
						position403 := position
						{
							position404 := position
							if !_rules[rule_]() {
								goto l390
							}
							if buffer[position] != rune('b') {
								goto l390
							}
							position++
							if buffer[position] != rune('e') {
								goto l390
							}
							position++
							if buffer[position] != rune('g') {
								goto l390
							}
							position++
							if buffer[position] != rune('i') {
								goto l390
							}
							position++
							if buffer[position] != rune('n') {
								goto l390
							}
							position++
							add(ruleBEGIN, position404)
						}
						if buffer[position] != rune('\n') {
							goto l390
						}
						position++
						{
							position405 := position
						l406:
							{
								position407, tokenIndex407 := position, tokenIndex
								{
									position408, tokenIndex408 := position, tokenIndex
									if !_rules[ruleEND]() {
										goto l408
									}
									goto l407
								l408:
									position, tokenIndex = position408, tokenIndex408
								}
								if !matchDot() {
									goto l407
								}
								goto l406
							l407:
								position, tokenIndex = position407, tokenIndex407
							}
							add(ruleHeredocBody, position405)
						}
						if !_rules[ruleEND]() {
							goto l390
						}
						add(ruleHeredoc, position403)
					}
				}
			l392:
				add(ruleString, position391)
			}
			return true
		l390:
			position, tokenIndex = position390, tokenIndex390
			return false
		},
		/* 112 StringLiteral <- <('\'' (!'\'' .)* '\'')> */
		nil,
		/* 113 StringInterpolated <- <('"' (!'"' .)* '"')> */
		nil,
		/* 114 Heredoc <- <(BEGIN '\n' HeredocBody END)> */
		nil,
		/* 115 HeredocBody <- <(!END .)*> */
		nil,
		/* 116 NullValue <- <('n' 'u' 'l' 'l')> */
		nil,
		/* 117 Object <- <(OPEN (_ KeyValuePair _)* CLOSE)> */
		func() bool {
			position414, tokenIndex414 := position, tokenIndex
			{
				position415 := position
				if !_rules[ruleOPEN]() {
					goto l414
				}
			l416:
				{
					position417, tokenIndex417 := position, tokenIndex
					if !_rules[rule_]() {
						goto l417
					}
					{
						position418 := position
						{
							position419 := position
							if !_rules[ruleIdentifier]() {
								goto l417
							}
							add(ruleKey, position419)
						}
						{
							position420 := position
							if !_rules[rule_]() {
								goto l417
							}
							if buffer[position] != rune(':') {
								goto l417
							}
							position++
							if !_rules[rule_]() {
								goto l417
							}
							add(ruleCOLON, position420)
						}
						{
							position421 := position
							{
								position422, tokenIndex422 := position, tokenIndex
								if !_rules[ruleArray]() {
									goto l423
								}
								goto l422
							l423:
								position, tokenIndex = position422, tokenIndex422
								if !_rules[ruleObject]() {
									goto l424
								}
								goto l422
							l424:
								position, tokenIndex = position422, tokenIndex422
								if !_rules[ruleExpression]() {
									goto l417
								}
							}
						l422:
							add(ruleKValue, position421)
						}
						{
							position425, tokenIndex425 := position, tokenIndex
							if !_rules[ruleCOMMA]() {
								goto l425
							}
							goto l426
						l425:
							position, tokenIndex = position425, tokenIndex425
						}
					l426:
						add(ruleKeyValuePair, position418)
					}
					if !_rules[rule_]() {
						goto l417
					}
					goto l416
				l417:
					position, tokenIndex = position417, tokenIndex417
				}
				if !_rules[ruleCLOSE]() {
					goto l414
				}
				add(ruleObject, position415)
			}
			return true
		l414:
			position, tokenIndex = position414, tokenIndex414
			return false
		},
		/* 118 Array <- <('[' _ ExpressionSequence COMMA? ']')> */
		func() bool {
			position427, tokenIndex427 := position, tokenIndex
			{
				position428 := position
				if buffer[position] != rune('[') {
					goto l427
				}
				position++
				if !_rules[rule_]() {
					goto l427
				}
				if !_rules[ruleExpressionSequence]() {
					goto l427
				}
				{
					position429, tokenIndex429 := position, tokenIndex
					if !_rules[ruleCOMMA]() {
						goto l429
					}
					goto l430
				l429:
					position, tokenIndex = position429, tokenIndex429
				}
			l430:
				if buffer[position] != rune(']') {
					goto l427
				}
				position++
				add(ruleArray, position428)
			}
			return true
		l427:
			position, tokenIndex = position427, tokenIndex427
			return false
		},
		/* 119 RegularExpression <- <('/' (!'/' .)+ '/' ('i' / 'l' / 'm' / 's' / 'u')*)> */
		func() bool {
			position431, tokenIndex431 := position, tokenIndex
			{
				position432 := position
				if buffer[position] != rune('/') {
					goto l431
				}
				position++
				{
					position435, tokenIndex435 := position, tokenIndex
					if buffer[position] != rune('/') {
						goto l435
					}
					position++
					goto l431
				l435:
					position, tokenIndex = position435, tokenIndex435
				}
				if !matchDot() {
					goto l431
				}
			l433:
				{
					position434, tokenIndex434 := position, tokenIndex
					{
						position436, tokenIndex436 := position, tokenIndex
						if buffer[position] != rune('/') {
							goto l436
						}
						position++
						goto l434
					l436:
						position, tokenIndex = position436, tokenIndex436
					}
					if !matchDot() {
						goto l434
					}
					goto l433
				l434:
					position, tokenIndex = position434, tokenIndex434
				}
				if buffer[position] != rune('/') {
					goto l431
				}
				position++
			l437:
				{
					position438, tokenIndex438 := position, tokenIndex
					{
						position439, tokenIndex439 := position, tokenIndex
						if buffer[position] != rune('i') {
							goto l440
						}
						position++
						goto l439
					l440:
						position, tokenIndex = position439, tokenIndex439
						if buffer[position] != rune('l') {
							goto l441
						}
						position++
						goto l439
					l441:
						position, tokenIndex = position439, tokenIndex439
						if buffer[position] != rune('m') {
							goto l442
						}
						position++
						goto l439
					l442:
						position, tokenIndex = position439, tokenIndex439
						if buffer[position] != rune('s') {
							goto l443
						}
						position++
						goto l439
					l443:
						position, tokenIndex = position439, tokenIndex439
						if buffer[position] != rune('u') {
							goto l438
						}
						position++
					}
				l439:
					goto l437
				l438:
					position, tokenIndex = position438, tokenIndex438
				}
				add(ruleRegularExpression, position432)
			}
			return true
		l431:
			position, tokenIndex = position431, tokenIndex431
			return false
		},
		/* 120 KeyValuePair <- <(Key COLON KValue COMMA?)> */
		nil,
		/* 121 Key <- <Identifier> */
		nil,
		/* 122 KValue <- <(Array / Object / Expression)> */
		nil,
		/* 123 Type <- <(Array / Object / RegularExpression / ScalarType)> */
		nil,
	}
	p.rules = _rules
}
