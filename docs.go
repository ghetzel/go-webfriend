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
}

type parsedSource struct {
	Docs      string
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

		if parsed, err := parseCommandSourceCode(fmt.Sprintf("commands/%s/*.go", name)); err == nil {
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
			log.Fatal(err)
		}

		docs = append(docs, doc)
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

						for i, inParam := range fnDecl.Type.Params.List {
							pFunc.Args[i] = &parsedArg{
								Name: inParam.Names[0].Name,
								Type: astTypeToString(inParam.Type),
							}
						}

						if len(fnDecl.Type.Results.List) > 1 {
							log.Debugf("ATS: fn=%v rtyp=%T", fnDecl.Name.Name, fnDecl.Type.Results.List[0].Type)

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

										stc.Fields = append(stc.Fields, &parsedStructField{
											Name:             sfield.Names[0].Name,
											Type:             astTypeToString(sfield.Type),
											FriendscriptName: stringutil.Underscore(sfield.Names[0].Name),
											Docs:             astCommentGroupToString(sfield.Doc),
										})
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
			line = strings.TrimPrefix(line, `//`)
			line = strings.TrimSpace(line)

			out += line + ` `
		}

		return strings.TrimSpace(out)
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
		return ident.Name
	} else if _, ok := ty.(*ast.InterfaceType); ok {
		return `any`
	} else if _, ok := ty.(*ast.MapType); ok {
		return `{}`
	} else if arrayOf, ok := ty.(*ast.ArrayType); ok {
		return fmt.Sprintf("[]%v", astTypeToString(arrayOf.Elt))
	} else {
		return `UNKNOWN`
	}
}
