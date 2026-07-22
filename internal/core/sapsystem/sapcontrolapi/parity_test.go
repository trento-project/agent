// SPDX-FileCopyrightText: SUSE LLC
// SPDX-License-Identifier: Apache-2.0

package sapcontrolapi

import (
	"go/ast"
	"go/parser"
	"go/token"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// Guards webservice.go's hand-extracted subset of the SAPControl web service against drift from the
// gowsdl-generated reference in _generated_wsdl.go.
// Both files are parsed as plain source rather than compiled, since they declare the same type names.
// Operations and their referenced types are discovered from webservice.go's WebService interface and matched by XML.

// typeUniverse indexes one file's declarations for comparison.
type typeUniverse struct {
	file       *ast.File
	structs    map[string]*ast.StructType // `type X struct {...}`, keyed by name
	scalars    map[string]bool            // `type X string`-style wire enums, keyed by name
	arrayItems map[string]string          // gowsdl's `ArrayOfX{ Item []Y }` -> Y
}

// buildTypeUniverse collects the structs, scalars, and array wrappers in a file.
func buildTypeUniverse(file *ast.File) *typeUniverse {
	structs := structsByName(file)
	return &typeUniverse{
		file:       file,
		structs:    structs,
		scalars:    namedScalarTypes(file),
		arrayItems: arrayItemTypes(structs),
	}
}

// operation describes a single WebService method's request and response types.
type operation struct {
	Request  string
	Response string
}

// TestFeatureParityWithGeneratedWSDL checks that webservice.go's hand-extracted subset of the SAPControl web service matches the gowsdl-generated reference in _generated_wsdl.go.
func TestFeatureParityWithGeneratedWSDL(t *testing.T) {
	handFile := parseGoFile(t, "webservice.go")
	genFile := parseGoFile(t, "_generated_wsdl.go")

	hand := buildTypeUniverse(handFile)
	gen := buildTypeUniverse(genFile)

	handOps := interfaceOperations(handFile)
	genOps := interfaceOperations(genFile)
	handBodies := methodBodies(handFile)
	require.NotEmpty(t, handOps, "found no operations in webservice.go's WebService interface to check")

	for methodName, handOp := range handOps {
		t.Run(methodName, func(t *testing.T) {
			genOp, ok := genOps[methodName]
			require.Truef(t, ok,
				"gowsdl reference does not define %s -- was this operation removed from (or never generated for) the WSDL?",
				methodName)

			visited := map[[2]string]bool{}
			assertTypesEquivalent(t, handOp.Request, genOp.Request, hand, gen, visited)
			assertTypesEquivalent(t, handOp.Response, genOp.Response, hand, gen, visited)

			body, ok := handBodies[methodName]
			require.Truef(t, ok, "webservice.go's WebService interface declares %s but has no method implementing it", methodName)
			assertCallContextPassesResponseDirectly(t, methodName, body)
		})
	}

	logUnwrappedOperations(t, handOps, genOps)
}

// logUnwrappedOperations logs any operations that exist in the gowsdl reference, but are not wrapped by webservice.go.
func logUnwrappedOperations(t *testing.T, handOps, genOps map[string]operation) {
	t.Helper()

	var unwrapped []string
	for methodName := range genOps {
		if _, ok := handOps[methodName]; !ok {
			unwrapped = append(unwrapped, methodName)
		}
	}
	sort.Strings(unwrapped)

	t.Logf("%d/%d gowsdl operations are not wrapped by webservice.go: %s",
		len(unwrapped), len(genOps), strings.Join(unwrapped, ", "))
}

// methodBodies maps method name -> body, for every method declared in the file.
func methodBodies(file *ast.File) map[string]*ast.BlockStmt {
	out := map[string]*ast.BlockStmt{}
	for _, decl := range file.Decls {
		funcDecl, ok := decl.(*ast.FuncDecl)
		if !ok || funcDecl.Recv == nil || funcDecl.Body == nil {
			continue
		}
		out[funcDecl.Name.Name] = funcDecl.Body
	}
	return out
}

// assertCallContextPassesResponseDirectly fails if the method's CallContext call takes the address of its response argument rather than passing it directly.
func assertCallContextPassesResponseDirectly(t *testing.T, methodName string, body *ast.BlockStmt) {
	t.Helper()

	found := false
	ast.Inspect(body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok || sel.Sel.Name != "CallContext" || len(call.Args) == 0 {
			return true
		}

		found = true
		_, isAddressOf := call.Args[len(call.Args)-1].(*ast.UnaryExpr)
		require.Falsef(t, isAddressOf,
			"%s passes its response to CallContext by address (e.g. &response); "+
				"response is already a pointer from new(...) and must be passed as-is",
			methodName)
		return true
	})

	require.Truef(t, found, "%s has no CallContext invocation to check", methodName)
}

// assertTypesEquivalent recursively checks that handType and genType describe the same.
func assertTypesEquivalent(
	t *testing.T,
	handType, genType string,
	hand, gen *typeUniverse,
	visited map[[2]string]bool,
) {
	t.Helper()

	key := [2]string{handType, genType}
	if visited[key] {
		return
	}
	visited[key] = true

	handStruct, handIsStruct := hand.structs[handType]
	genStruct, genIsStruct := gen.structs[genType]
	require.Equalf(t, genIsStruct, handIsStruct,
		"webservice.go's %s and the generated reference's %s disagree on whether they're a struct",
		handType, genType)

	if handIsStruct {
		assertStructFieldsEquivalent(t, handType, genType, handStruct, genStruct, hand, gen, visited)
		return
	}

	if hand.scalars[handType] || gen.scalars[genType] {
		genValues := constValuesByType(gen.file, genType)
		require.NotEmptyf(t, genValues, "gowsdl reference's %s has no declared string constants", genType)

		handValues := constValuesByType(hand.file, handType)
		require.Equalf(t, genValues, handValues,
			"webservice.go's %s enum values (generated: %s) drifted from the gowsdl reference; "+
				"update webservice.go's constants to match",
			handType, genType)
		return
	}

	require.Equalf(t, genType, handType,
		"underlying type drifted from the gowsdl reference: webservice.go uses %s where generated uses %s",
		handType, genType)
}

// assertStructFieldsEquivalent checks that two struct types have the same XML elements and field types.
func assertStructFieldsEquivalent(
	t *testing.T,
	handName, genName string,
	handStruct, genStruct *ast.StructType,
	hand, gen *typeUniverse,
	visited map[[2]string]bool,
) {
	t.Helper()

	handFields := elementFieldTypes(handStruct, hand.arrayItems)
	genFields := elementFieldTypes(genStruct, gen.arrayItems)

	require.Equalf(t, sortedKeys(genFields), sortedKeys(handFields),
		"webservice.go's %s XML elements drifted from the gowsdl reference's %s; update webservice.go to match",
		handName, genName)

	for element, genFieldType := range genFields {
		handFieldType, ok := handFields[element]
		if !ok {
			continue // already reported by the key-set check above
		}
		assertTypesEquivalent(t, handFieldType, genFieldType, hand, gen, visited)
	}

	require.Equalf(t, xmlNameTag(genStruct), xmlNameTag(handStruct),
		"webservice.go's %s XMLName tag drifted from the gowsdl reference's %s", handName, genName)
}

// sortedKeys returns the keys of a map in sorted order.
func sortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// parseGoFile parses a Go source file and returns its AST.
func parseGoFile(t *testing.T, path string) *ast.File {
	t.Helper()
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	require.NoErrorf(t, err, "failed to parse %s", path)
	return file
}

// structsByName collects every `type X struct {...}` declaration, keyed by name.
func structsByName(file *ast.File) map[string]*ast.StructType {
	out := map[string]*ast.StructType{}
	forEachTypeSpec(file, func(spec *ast.TypeSpec) {
		if structType, ok := spec.Type.(*ast.StructType); ok {
			out[spec.Name.Name] = structType
		}
	})
	return out
}

// namedScalarTypes collects every `type X <builtin>` declaration.
func namedScalarTypes(file *ast.File) map[string]bool {
	out := map[string]bool{}
	forEachTypeSpec(file, func(spec *ast.TypeSpec) {
		if _, ok := spec.Type.(*ast.Ident); ok {
			out[spec.Name.Name] = true
		}
	})
	return out
}

// interfaceOperations finds every method shaped like in the file's interfaces and returns its request/response type names.
func interfaceOperations(file *ast.File) map[string]operation {
	out := map[string]operation{}
	forEachTypeSpec(file, func(spec *ast.TypeSpec) {
		ifaceType, ok := spec.Type.(*ast.InterfaceType)
		if !ok {
			return
		}
		for _, m := range ifaceType.Methods.List {
			funcType, ok := m.Type.(*ast.FuncType)
			if !ok || len(m.Names) == 0 {
				continue
			}
			if funcType.Params == nil || len(funcType.Params.List) != 2 {
				continue
			}
			if exprString(funcType.Params.List[0].Type) != "context.Context" {
				continue
			}
			if funcType.Results == nil || len(funcType.Results.List) != 2 {
				continue
			}
			if exprString(funcType.Results.List[1].Type) != "error" {
				continue
			}

			out[m.Names[0].Name] = operation{
				Request:  baseType(exprString(funcType.Params.List[1].Type)),
				Response: baseType(exprString(funcType.Results.List[0].Type)),
			}
		}
	})
	return out
}

// forEachTypeSpec calls fn for every `type X ...` declaration in the file.
func forEachTypeSpec(file *ast.File, fn func(spec *ast.TypeSpec)) {
	for _, decl := range file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.TYPE {
			continue
		}
		for _, spec := range genDecl.Specs {
			if typeSpec, ok := spec.(*ast.TypeSpec); ok {
				fn(typeSpec)
			}
		}
	}
}

// exprString returns a string representation of an ast.Expr, e.g. "[]*OSProcess".
func exprString(e ast.Expr) string {
	switch v := e.(type) {
	case *ast.Ident:
		return v.Name
	case *ast.StarExpr:
		return "*" + exprString(v.X)
	case *ast.SelectorExpr:
		return exprString(v.X) + "." + v.Sel.Name
	case *ast.ArrayType:
		return "[]" + exprString(v.Elt)
	default:
		return "<unsupported>"
	}
}

// structField represents a single field in a struct, including its name, type, and XML tag.
type structField struct {
	Name string
	Type string
	Tag  string
}

// structFields returns a slice of structField for each field in the given struct type.
func structFields(s *ast.StructType) []structField {
	var out []structField
	for _, f := range s.Fields.List {
		tag := ""
		if f.Tag != nil {
			tag = f.Tag.Value
		}
		name := ""
		if len(f.Names) > 0 {
			name = f.Names[0].Name
		}
		out = append(out, structField{Name: name, Type: exprString(f.Type), Tag: tag})
	}
	return out
}

// xmlTag extracts the XML tag from a struct field's raw tag string.
func xmlTag(rawTag string) string {
	unquoted, err := strconv.Unquote(rawTag)
	if err != nil {
		return ""
	}
	return reflect.StructTag(unquoted).Get("xml")
}

// xmlNameTag returns the XMLName tag of a struct.
func xmlNameTag(s *ast.StructType) string {
	for _, f := range structFields(s) {
		if f.Name == "XMLName" {
			return xmlTag(f.Tag)
		}
	}
	return ""
}

// firstTagSegment extracts the top-level element name from an xml tag.
func firstTagSegment(tag string) string {
	tag = strings.SplitN(tag, ",", 2)[0]
	if idx := strings.Index(tag, ">"); idx >= 0 {
		tag = tag[:idx]
	}
	return tag
}

// baseType strips leading "*"/"[]" wrappers down to the named type.
func baseType(t string) string {
	for {
		switch {
		case strings.HasPrefix(t, "[]"):
			t = t[2:]
		case strings.HasPrefix(t, "*"):
			t = t[1:]
		default:
			return t
		}
	}
}

// elementFieldTypes maps each field's XML element name to its referenced type.
func elementFieldTypes(s *ast.StructType, arrayItems map[string]string) map[string]string {
	out := map[string]string{}
	for _, f := range structFields(s) {
		if f.Name == "XMLName" {
			continue
		}

		elementName := firstTagSegment(xmlTag(f.Tag))
		if elementName == "" || elementName == "-" {
			continue // no xml tag means it's not part of the wire shape
		}

		fieldType := baseType(f.Type)
		if itemType, ok := arrayItems[fieldType]; ok {
			fieldType = itemType
		}

		out[elementName] = fieldType
	}
	return out
}

// arrayItemTypes maps each gowsdl `ArrayOfX{ Item []Y }` wrapper to Y.
func arrayItemTypes(structs map[string]*ast.StructType) map[string]string {
	out := map[string]string{}
	for name, s := range structs {
		if !strings.HasPrefix(name, "ArrayOf") {
			continue
		}
		fs := structFields(s)
		if len(fs) != 1 || fs[0].Name != "Item" {
			continue
		}
		out[name] = baseType(fs[0].Type)
	}
	return out
}

// constValuesByType returns the string values of every const declared with the given named type.
func constValuesByType(file *ast.File, typeName string) map[string]bool {
	out := map[string]bool{}
	for _, decl := range file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.CONST {
			continue
		}
		for _, spec := range genDecl.Specs {
			valueSpec, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}
			ident, ok := valueSpec.Type.(*ast.Ident)
			if !ok || ident.Name != typeName {
				continue
			}
			for _, v := range valueSpec.Values {
				lit, ok := v.(*ast.BasicLit)
				if !ok {
					continue
				}
				if unquoted, err := strconv.Unquote(lit.Value); err == nil {
					out[unquoted] = true
				}
			}
		}
	}
	return out
}
