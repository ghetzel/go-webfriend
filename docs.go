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
}
type CallDoc struct {
	Name        string     `json:"name"`
	Description string     `json:"description,omitempty"`
	Argument    *DocItem   `json:"argument,omitempty"`
	Options     []*DocItem `json:"options,omitempty"`
	Return      *DocItem   `json:"return,omitempty"`
}

type ModuleDoc struct {
	Name        string    `json:"name"`
	Summary     string    `json:"summary,omitempty"`
	Description string    `json:"description,omitempty"`
	Commands    []CallDoc `json:"commands"`
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

func (self *Environment) Documentation() []ModuleDoc {
	docs := make([]ModuleDoc, 0)
	modnames := []string{`core`}
	remaining := make([]string, 0)

	for name, _ := range self.modules {
		if name == `core` {
			continue
		} else {
			remaining = append(remaining, name)
		}
	}

	sort.Strings(remaining)
	modnames = append(modnames, remaining...)

	for _, name := range modnames {
		module := self.modules[name]

		doc := ModuleDoc{
			Name:     name,
			Commands: make([]CallDoc, 0),
		}

		sourcePath := fmt.Sprintf("commands/%s/*.go", name)

		log.Debugf("Parsing %q (%T) from %v", name, module, sourcePath)

		if parsed, err := parseCommandSourceCode(sourcePath); err == nil {
			doc.Summary = parsed.Summary
			doc.Description = parsed.Docs
			moduleT := reflect.TypeOf(module)

			for i := 0; i < moduleT.NumMethod(); i++ {
				fn := moduleT.Method(i)
				key := stringutil.Underscore(fn.Name)

				if key == `execute_command` {
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

				doc.Commands = append(doc.Commands, cmdDoc)
			}
		} else {
			log.Errorf("Error parsing source: %v", err)
		}

		docs = append(docs, doc)
		log.Infof("Documented %q: %d commands", doc.Name, len(doc.Commands))
	}

	return docs
}

func parseCommandSourceCode(fileglob string) (*parsedSource, error) {
	parsed := &parsedSource{
		Functions: make(map[string]parsedFunc),
		Structs:   make(map[string]parsedStruct),
	}

	if sources, err := filepath.Glob(fileglob); err == nil {
		for _, sourcefile := range sources {
			if source, err := parser.ParseFile(
				token.NewFileSet(),
				sourcefile,
				nil,
				parser.ParseComments,
			); err == nil {
				if parsed.Docs == `` {
					lines := strings.Split(astCommentGroupToString(source.Doc), "\n")
					goToDocs := false

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
				}

				for _, decl := range source.Decls {
					switch decl.(type) {
					case *ast.FuncDecl:
						fnDecl := decl.(*ast.FuncDecl)
						key := stringutil.Underscore(fnDecl.Name.Name)

						pFunc := parsedFunc{
							Name:             fnDecl.Name.Name,
							FriendscriptName: key,
							Docs:             astCommentGroupToString(fnDecl.Doc),
							Args:             make(parsedArgSet, len(fnDecl.Type.Params.List)),
						}

						if strings.Contains(pFunc.Docs, `[SKIP]`) {
							pFunc.Skip = true
						}

						for i, inParam := range fnDecl.Type.Params.List {
							pFunc.Args[i] = &parsedArg{
								Name: inParam.Names[0].Name,
								Type: astTypeToString(inParam.Type),
							}
						}

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
									key := typeSpec.Name.Name
									stc := parsedStruct{
										Name:   key,
										Fields: make([]*parsedStructField, 0),
									}

									// built list of struct fields, including associated documentation
									for _, sfield := range structType.Fields.List {
										structField := &parsedStructField{
											Name:             sfield.Names[0].Name,
											Type:             astTypeToString(sfield.Type),
											FriendscriptName: stringutil.Underscore(sfield.Names[0].Name),
											Docs:             astCommentGroupToString(sfield.Doc),
										}

										if sfield.Tag != nil {
											tags := strings.Split(
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
		out := ``

		for _, c := range cg.List {
			line := strings.TrimSpace(c.Text)

			if !strings.HasPrefix(line, `//`) {
				continue
			}

			if strings.TrimSpace(line) == `` {
				out += "\n"
			} else {
				out += line + "\n"
			}
		}

		lines := strings.Split(out, "\n")

		for i, _ := range lines {
			lines[i] = strings.TrimPrefix(lines[i], `//`)
			lines[i] = strings.TrimPrefix(lines[i], ` `)
		}

		return strings.TrimSpace(strings.Join(lines, "\n"))
	}

	return ``
}

func astTypeToString(ty ast.Expr) string {
	if star, ok := ty.(*ast.StarExpr); ok {
		ty = star.X
	}

	if sel, ok := ty.(*ast.SelectorExpr); ok {
		ty = sel.Sel
	}

	if ident, ok := ty.(*ast.Ident); ok {
		name := ident.Name

		name = strings.TrimPrefix(name, `*`)
		name = strings.TrimPrefix(name, `u`)
		name = strings.TrimSuffix(name, `8`)
		name = strings.TrimSuffix(name, `16`)
		name = strings.TrimSuffix(name, `32`)
		name = strings.TrimSuffix(name, `64`)
		name = strings.TrimSuffix(name, `128`)

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
