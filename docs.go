package webfriend

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"reflect"
	"sort"
	"strings"

	"github.com/ghetzel/go-stockutil/log"
	"github.com/ghetzel/go-stockutil/sliceutil"
	"github.com/ghetzel/go-stockutil/stringutil"
)

type DocItem struct {
	Name         string      `json:"name,omitempty"`
	Type         string      `json:"types"`
	Required     bool        `json:"required,omitempty"`
	Description  string      `json:"description,omitempty"`
	DefaultValue interface{} `json:"default,omitempty"`
	Examples     []string    `json:"examples,omitempty"`
	Parameters   []*DocItem  `json:"parameters,omitempty"`
}
type CallDoc struct {
	Name        string     `json:"name"`
	Description string     `json:"description,omitempty"`
	Argument    *DocItem   `json:"argument,omitempty"`
	Options     []*DocItem `json:"options,omitempty"`
	Return      *DocItem   `json:"return,omitempty"`
}

type CallDocSet []*CallDoc

type ModuleDoc struct {
	DisplayName string     `json:"display_name"`
	Name        string     `json:"name"`
	Summary     string     `json:"summary,omitempty"`
	Description string     `json:"description,omitempty"`
	Commands    CallDocSet `json:"commands"`
	commandSet  map[string]*CallDoc
}

func (self *ModuleDoc) AddCommand(name string, doc *CallDoc) {
	doc.Description = processComment(doc.Description)

	if len(self.commandSet) == 0 {
		self.commandSet = make(map[string]*CallDoc)
	}

	if doc.Description == `` {
		return
	} else {
		self.commandSet[name] = doc
	}

	self.Commands = nil

	for _, cmd := range self.commandSet {
		self.Commands = append(self.Commands, cmd)
	}

	sort.Slice(self.Commands, func(i int, j int) bool {
		return (self.Commands[i].Name < self.Commands[j].Name)
	})

	log.Infof("  added command %q // %s...", name, stringutil.ElideWords(doc.Description, 4))
}

type parsedStructField struct {
	Name             string
	FriendscriptName string
	Type             string
	Docs             string
	Default          string
}

type parsedStruct struct {
	Name   string
	Fields []*parsedStructField
}

type parsedArg struct {
	Name      string
	Type      string
	Primitive bool
}

func (self parsedArg) String() string {
	return strings.TrimSpace(fmt.Sprintf("%v %v", self.Name, self.Type))
}

type parsedArgSet []*parsedArg

func (self parsedArgSet) String() string {
	return strings.Join(sliceutil.Stringify(self), ` `)
}

type parsedFunc struct {
	Name             string
	FriendscriptName string
	Docs             string
	Args             parsedArgSet
	Return           *parsedArg
	Skip             bool
}

type parsedSource struct {
	Docs      string
	Summary   string
	Functions map[string]parsedFunc
	Structs   map[string]parsedStruct
}

func (self *Environment) Documentation() []*ModuleDoc {
	var mods = make(map[string]*ModuleDoc)
	var modnames = []string{`core`}
	var remaining = make([]string, 0)
	var modules = self.Modules()

	for name := range modules {
		if name == `core` {
			continue
		} else {
			remaining = append(remaining, name)
		}
	}

	sort.Strings(remaining)
	modnames = append(modnames, remaining...)

	for _, name := range modnames {
		var module = modules[name]
		var mod *ModuleDoc

		if m, ok := mods[name]; ok {
			mod = m
		} else {
			mod = &ModuleDoc{
				DisplayName: mangleModName(name),
				Name:        name,
				Commands:    make(CallDocSet, 0),
			}

			mods[name] = mod
		}

		var sourcePaths = []string{
			fmt.Sprintf("../friendscript/commands/%s/*.go", name),
			fmt.Sprintf("commands/%s/*.go", name),
		}

		for _, sourcePath := range sourcePaths {
			log.Debugf("Parsing %q (%T) from %v", name, module, sourcePath)

			if parsed, err := parseCommandSourceCode(sourcePath); err == nil {
				mod.Summary = parsed.Summary
				mod.Description = parsed.Docs
				var moduleT = reflect.TypeOf(module)

				log.Debugf("Methods: %d", moduleT.NumMethod())

				for i := 0; i < moduleT.NumMethod(); i++ {
					var fn = moduleT.Method(i)
					var key = stringutil.Underscore(fn.Name)

					switch key {
					case `execute_command`, `format_command_name`, `set_instance`, `new`:
						continue
					}

					cmdDoc := CallDoc{
						Name: key,
					}

					if fnDoc, ok := parsed.Functions[key]; ok && fnDoc.Docs != `` {
						if fnDoc.Skip {
							continue
						}

						cmdDoc.Description = fnDoc.Docs

						if r := fnDoc.Return; r != nil {
							cmdDoc.Return = &DocItem{
								Type: fmt.Sprintf("%v", r),
							}

							if subargs, ok := parsed.Structs[cmdDoc.Return.Type]; ok {
								for _, arg := range subargs.Fields {
									cmdDoc.Return.Parameters = append(cmdDoc.Return.Parameters, &DocItem{
										Name:         arg.FriendscriptName,
										Type:         arg.Type,
										Description:  arg.Docs,
										DefaultValue: stringutil.Autotype(arg.Default),
									})
								}
							}
						}

						for i, arg := range fnDoc.Args {
							if subargs, ok := parsed.Structs[arg.Type]; ok {
								for _, arg := range subargs.Fields {
									cmdDoc.Options = append(cmdDoc.Options, &DocItem{
										Name:         arg.FriendscriptName,
										Type:         arg.Type,
										Description:  arg.Docs,
										DefaultValue: stringutil.Autotype(arg.Default),
									})
								}
							} else {

								if i == 0 {
									cmdDoc.Argument = &DocItem{
										Name: arg.Name,
										Type: arg.Type,
									}
								} else {
									cmdDoc.Options = append(cmdDoc.Options, &DocItem{
										Name: arg.Name,
										Type: arg.Type,
									})
								}
							}
						}
					}

					mod.AddCommand(cmdDoc.Name, &cmdDoc)
				}
			} else {
				log.Errorf("Error parsing source: %v", err)
			}

			log.Infof("Documented %q: %d commands", mod.Name, len(mod.Commands))
		}
	}

	var sortedMods = make([]*ModuleDoc, 0)

	for _, mod := range mods {
		sortedMods = append(sortedMods, mod)
	}

	sort.Slice(sortedMods, func(i int, j int) bool {
		return (sortedMods[i].Name < sortedMods[j].Name)
	})

	return sortedMods
}

func parseCommandSourceCode(fileglob string) (*parsedSource, error) {
	var parsed = &parsedSource{
		Functions: make(map[string]parsedFunc),
		Structs:   make(map[string]parsedStruct),
	}

	if sources, err := filepath.Glob(fileglob); err == nil {
		for _, sourcefile := range sources {
			log.Infof("Processing file %s", sourcefile)

			if source, err := parser.ParseFile(
				token.NewFileSet(),
				sourcefile,
				nil,
				parser.ParseComments,
			); err == nil {
				var lines = strings.Split(astCommentGroupToString(source.Doc), "\n")
				var goToDocs = false

				for _, line := range lines {
					if strings.TrimSpace(line) == `` {
						goToDocs = true
						continue
					}

					if goToDocs {
						parsed.Docs += line + "\n"
					} else {
						parsed.Summary += line + "\n"
					}
				}

				parsed.Summary = strings.TrimSpace(parsed.Summary)
				parsed.Docs = strings.TrimSpace(parsed.Docs)

				for _, decl := range source.Decls {
					switch decl.(type) {
					case *ast.FuncDecl: // describe a function declaration from its source code

						fnDecl := decl.(*ast.FuncDecl)
						key := stringutil.Underscore(fnDecl.Name.Name)

						pFunc := parsedFunc{
							Name:             fnDecl.Name.Name,
							FriendscriptName: key,
							Docs:             astCommentGroupToString(fnDecl.Doc),
							Args:             make(parsedArgSet, len(fnDecl.Type.Params.List)),
						}

						// functions preceded by a "// [SKIP]" command will be omitted
						if strings.Contains(pFunc.Docs, `[SKIP]`) {
							pFunc.Skip = true
						}

						// extract function parameter names
						for i, inParam := range fnDecl.Type.Params.List {
							pFunc.Args[i] = &parsedArg{
								Name: inParam.Names[0].Name,
								Type: astTypeToString(inParam.Type),
							}
						}

						// extract function return types
						if fnDecl.Type.Results != nil && len(fnDecl.Type.Results.List) > 1 {
							pFunc.Return = &parsedArg{
								Type: astTypeToString(fnDecl.Type.Results.List[0].Type),
							}
						}

						parsed.Functions[key] = pFunc

					case *ast.GenDecl:
						if gdecl := decl.(*ast.GenDecl); len(gdecl.Specs) > 0 {
							if typeSpec, ok := gdecl.Specs[0].(*ast.TypeSpec); ok {
								if structType, ok := typeSpec.Type.(*ast.StructType); ok {
									var key = typeSpec.Name.Name
									var stc = parsedStruct{
										Name:   key,
										Fields: make([]*parsedStructField, 0),
									}

									// built list of struct fields, including associated documentation
									for _, sfield := range structType.Fields.List {
										if len(sfield.Names) == 0 {
											continue
										}

										var structField = &parsedStructField{
											Name:             sfield.Names[0].Name,
											Type:             astTypeToString(sfield.Type),
											FriendscriptName: stringutil.Underscore(sfield.Names[0].Name),
											Docs:             astCommentGroupToString(sfield.Doc),
										}

										if sfield.Tag != nil {
											var tags = strings.Split(
												stringutil.Unwrap(sfield.Tag.Value, "`", "`"),
												` `,
											)

											for _, tag := range tags {
												tag = strings.TrimSpace(tag)

												if strings.HasPrefix(tag, `default:`) {
													structField.Default = stringutil.Unwrap(
														strings.TrimPrefix(tag, `default:`),
														`"`,
														`"`,
													)

													var shouldQuote bool

													switch structField.Type {
													case `string`, `Duration`, `Selector`:
														shouldQuote = true
													}

													if shouldQuote {
														structField.Default = stringutil.Wrap(structField.Default, `"`, `"`)
													}

													break
												}
											}
										}

										stc.Fields = append(stc.Fields, structField)
									}

									parsed.Structs[key] = stc
								}
							}
						}
					}
				}
			} else {
				return nil, fmt.Errorf("Failed to parse source for module %v: %v", source, err)
			}
		}
	} else {
		return nil, fmt.Errorf("Failed to read source files: %v", err)
	}

	return parsed, nil
}

func astCommentGroupToString(cg *ast.CommentGroup) string {
	if cg != nil {
		var out string

		for _, c := range cg.List {
			var line = strings.TrimSpace(c.Text)

			if !strings.HasPrefix(line, `//`) {
				continue
			}

			if strings.TrimSpace(line) == `` {
				out += "\n"
			} else {
				out += line + "\n"
			}
		}

		return processComment(out)
	}

	return ``
}

func processComment(out string) string {
	var lines = strings.Split(out, "\n")

	for i := range lines {
		lines[i] = strings.TrimPrefix(lines[i], `//`)
		lines[i] = strings.TrimPrefix(lines[i], ` `)
	}

	return strings.TrimSpace(strings.Join(lines, "\n"))
}

func astTypeToString(ty ast.Expr) string {
	if star, ok := ty.(*ast.StarExpr); ok {
		ty = star.X
	}

	if sel, ok := ty.(*ast.SelectorExpr); ok {
		ty = sel.Sel
	}

	if ident, ok := ty.(*ast.Ident); ok {
		var name = ident.Name

		name = strings.TrimPrefix(name, `*`)
		name = strings.TrimPrefix(name, `u`)
		name = strings.TrimSuffix(name, `8`)
		name = strings.TrimSuffix(name, `16`)
		name = strings.TrimSuffix(name, `32`)
		name = strings.TrimSuffix(name, `64`)
		name = strings.TrimSuffix(name, `128`)

		switch name {
		case `Reader`, `Writer`:
			name = `stream`
		}

		return name
	} else if _, ok := ty.(*ast.InterfaceType); ok {
		return `any`
	} else if _, ok := ty.(*ast.MapType); ok {
		return `map`
	} else if arrayOf, ok := ty.(*ast.ArrayType); ok {
		return fmt.Sprintf("[]%v", astTypeToString(arrayOf.Elt))
	} else {
		return `UNKNOWN`
	}
}

func mangleModName(in string) string {
	switch in {
	case `fmt`:
		return `format`
	default:
		return in
	}
}
