package webfriend

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/ghetzel/go-stockutil/log"
	"github.com/ghetzel/go-stockutil/stringutil"
	"github.com/ghetzel/go-webfriend/commands"
)

type parsedStructField struct {
	Name             string
	FriendscriptName string
	Type             string
	Docs             string
	Default          string
}

type parsedStruct struct {
	Name   string
	Fields []parsedStructField
}

type parsedArg struct {
	Name      string
	Type      string
	Primitive bool
}

type parsedFunc struct {
	Name             string
	FriendscriptName string
	Docs             string
	Args             []parsedArg
	Return           parsedArg
}

type parsedSource struct {
	Functions map[string]parsedFunc
	Structs   map[string]parsedStruct
}

func (self *Environment) Documentation() []commands.ModuleDoc {
	docs := make([]commands.ModuleDoc, 0)

	for name, module := range self.modules {
		moduleT := reflect.TypeOf(module)
		parsed := &parsedSource{
			Functions: make(map[string]parsedFunc),
			Structs:   make(map[string]parsedStruct),
		}

		if sources, err := filepath.Glob(fmt.Sprintf("commands/%s/*.go", name)); err == nil {
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
								Args:             make([]parsedArg, len(fnDecl.Type.Params.List)),
							}

							for i, inParam := range fnDecl.Type.Params.List {
								pFunc.Args[i] = parsedArg{
									Name: inParam.Names[0].Name,
								}
							}

							// for _, outParam := range fnDecl.Type.Results.List {
							// 	if
							// }

							parsed.Functions[key] = pFunc

						case *ast.GenDecl:
							if gdecl := decl.(*ast.GenDecl); len(gdecl.Specs) > 0 {
								if typeSpec, ok := gdecl.Specs[0].(*ast.TypeSpec); ok {
									if structType, ok := typeSpec.Type.(*ast.StructType); ok {
										key := typeSpec.Name.Name
										stc := parsedStruct{
											Name:   key,
											Fields: make([]parsedStructField, 0),
										}

										for _, sfield := range structType.Fields.List {
											stc.Fields = append(stc.Fields, parsedStructField{
												Name:             sfield.Names[0].Name,
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
					log.Fatalf("Failed to parse source for module %v: %v", name, err)
					return nil
				}
			}
		} else {
			log.Fatalf("Failed to read source files for module %v: %v", name, err)
			return nil
		}

		for i := 0; i < moduleT.NumMethod(); i++ {
			fn := moduleT.Method(i)
			key := stringutil.Underscore(fn.Name)

			log.Debugf("[doc] %v::%v", name, key)

			if fnDoc, ok := parsed.Functions[key]; ok && fnDoc.Docs != `` {
				log.Debugf("[doc]   %v", fnDoc.Docs)
			} else {
				log.Debugf("[doc]   UNDOCUMENTED")
			}

			log.Debugf("[doc]")
		}
	}

	return docs
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
