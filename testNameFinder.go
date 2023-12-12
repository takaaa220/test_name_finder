package testNameFinder

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
)

type Selection struct {
	LineNumber  int
	StartCursor int
	EndCursor   int
}

func (s Selection) String() string {
	return fmt.Sprintf("L%d:%d-%d", s.LineNumber, s.StartCursor, s.EndCursor)
}

func (s Selection) Valid() (string, bool) {
	if s.LineNumber <= 0 {
		return "lineNumber should be greater than 0", false
	}

	if s.StartCursor < 0 {
		return "startCursor should be greater than or equal to 0", false
	}

	if s.StartCursor > s.EndCursor {
		return "startCursor should be less than or equal to endCursor", false
	}

	return "", true
}

type TestName struct {
	FuncName string
	TestCase string
}

func (t TestName) String() string {
	if t.TestCase == "" {
		return t.FuncName
	}

	return fmt.Sprintf("%s/%s", t.FuncName, t.TestCase)
}

func FindTestName(filePath string, selection Selection) (*TestName, error) {
	if msg, ok := selection.Valid(); !ok {
		return nil, fmt.Errorf(msg)
	}

	fset := token.NewFileSet()
	fileNode, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse file: %w", err)
	}

	file := fset.File(fileNode.Pos())

	decl, found := findTargetTestFuncDecl(fileNode, file, selection)
	if !found {
		return nil, fmt.Errorf("failed to find target test function declaration")
	}

	testCase, _ := findTestCaseInTestFuncDecl(file, decl, selection)

	return &TestName{
		FuncName: decl.Name.Name,
		TestCase: testCase,
	}, nil
}

func findTargetTestFuncDecl(fileNode *ast.File, file *token.File, selection Selection) (*ast.FuncDecl, bool) {
	for _, decl := range fileNode.Decls {
		funcDecl, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}

		lineStartPos := file.LineStart(selection.LineNumber)

		start := lineStartPos + token.Pos(selection.StartCursor)
		end := lineStartPos + token.Pos(selection.EndCursor)

		if funcDecl.Pos() <= start && end <= funcDecl.End() {
			if strings.HasPrefix(funcDecl.Name.Name, "Test") {
				return funcDecl, true
			}

			return nil, false
		}
	}

	return nil, false
}

func findTestCaseInTestFuncDecl(file *token.File, funcDecl *ast.FuncDecl, selection Selection) (string, bool) {
	// func TestXXX(t testing.T)
	if len(funcDecl.Type.Params.List) == 0 {
		return "", false
	}
	testingTVariableName := funcDecl.Type.Params.List[0].Names[0].Name

	rangeStmtForRunTest := findRangeStmtForRunTest(funcDecl, testingTVariableName)
	if rangeStmtForRunTest == nil {
		return "", false
	}

	tableVariableIdent, ok := rangeStmtForRunTest.X.(*ast.Ident)
	if !ok {
		return "", false
	}

	tableVariableName := tableVariableIdent.Name

	assignStmtForTableVariable := findAssignStmtOfTableVariable(funcDecl, tableVariableName)
	if assignStmtForTableVariable == nil {
		return "", false
	}

	decider := createTestNameDecider(assignStmtForTableVariable)
	if decider == nil {
		return "", false
	}

	assignStmtRhsCompositeLit, ok := assignStmtForTableVariable.Rhs[0].(*ast.CompositeLit)
	if !ok {
		return "", false
	}

	testCase := findTestCaseFromSelection(assignStmtRhsCompositeLit, file, selection, decider)

	return testCase, testCase != ""
}

// find RangeStmt that has t.Run CallExpr
func findRangeStmtForRunTest(funcDecl *ast.FuncDecl, testingTVariableName string) *ast.RangeStmt {
	for _, stmt := range funcDecl.Body.List {
		rStmt, ok := stmt.(*ast.RangeStmt)
		if !ok {
			continue
		}

		for _, inStmt := range rStmt.Body.List {
			exprStmt, ok := inStmt.(*ast.ExprStmt)
			if !ok {
				continue
			}

			callExpr, ok := exprStmt.X.(*ast.CallExpr)
			if !ok {
				continue
			}

			selectorExpr, ok := callExpr.Fun.(*ast.SelectorExpr)
			if !ok {
				continue
			}

			selectorIdent, ok := selectorExpr.X.(*ast.Ident)
			if !ok {
				continue
			}

			if selectorIdent.Name != testingTVariableName {
				continue
			}

			if selectorExpr.Sel.Name != "Run" {
				continue
			}

			return rStmt
		}
	}

	return nil
}

// find AssignStmt whose first Lhs's name is tableVariableName
func findAssignStmtOfTableVariable(funcDecl *ast.FuncDecl, tableVariableName string) *ast.AssignStmt {
	for _, stmt := range funcDecl.Body.List {
		assignStmt, _ := stmt.(*ast.AssignStmt)
		if assignStmt == nil {
			continue
		}

		if len(assignStmt.Lhs) == 0 {
			continue
		}

		ident, ok := assignStmt.Lhs[0].(*ast.Ident)
		if !ok {
			continue
		}

		if ident.Name != tableVariableName {
			continue
		}

		return assignStmt
	}

	return nil
}

type testNameDecider struct {
	isSlice       bool
	testNameField string
}

func (d testNameDecider) findTest(expr ast.Expr) (string, bool) {
	if d.isSlice {
		return d.findTestFromSlice(expr)
	}

	return d.findTestFromMap(expr)
}

func (d testNameDecider) findTestFromMap(expr ast.Expr) (string, bool) {
	if d.isSlice {
		return "", false
	}

	keyValueExpr, ok := expr.(*ast.KeyValueExpr)
	if !ok {
		return "", false
	}

	basicLit, ok := keyValueExpr.Key.(*ast.BasicLit)
	if !ok {
		return "", false
	}

	if basicLit.Kind != token.STRING {
		return "", false
	}

	return strings.Trim(basicLit.Value, "\""), true
}

func (d testNameDecider) findTestFromSlice(expr ast.Expr) (string, bool) {
	if !d.isSlice {
		return "", false
	}

	compositeLit, ok := expr.(*ast.CompositeLit)
	if !ok {
		return "", false
	}

	for _, elt := range compositeLit.Elts {
		keyValueExpr, ok := elt.(*ast.KeyValueExpr)
		if !ok {
			continue
		}

		keyIdent, ok := keyValueExpr.Key.(*ast.Ident)
		if !ok {
			continue
		}

		if keyIdent.Name != d.testNameField {
			continue
		}

		basicLit, ok := keyValueExpr.Value.(*ast.BasicLit)
		if !ok {
			continue
		}

		if basicLit.Kind != token.STRING {
			continue
		}

		return strings.Trim(basicLit.Value, "\""), true
	}

	return "", false
}

func createTestNameDecider(assignStmt *ast.AssignStmt) *testNameDecider {
	compositeLit, ok := assignStmt.Rhs[0].(*ast.CompositeLit)
	if !ok || len(assignStmt.Lhs) == 0 {
		return nil
	}

	if arrayType, ok := compositeLit.Type.(*ast.ArrayType); ok {
		if _, ok := arrayType.Elt.(*ast.StructType); ok {
			return &testNameDecider{
				isSlice:       true,
				testNameField: "name",
			}
		}
	}
	if mapType, ok := compositeLit.Type.(*ast.MapType); ok {
		if _, ok := mapType.Value.(*ast.StructType); ok {
			return &testNameDecider{
				isSlice: false,
			}
		}
	}

	return nil
}

func findTestCaseFromSelection(assignStmtRhsCompositeLit *ast.CompositeLit, file *token.File, selection Selection, decider *testNameDecider) string {
	for _, elt := range assignStmtRhsCompositeLit.Elts {
		nodeStartLineNumber := file.Line(elt.Pos())
		nodeEndLineNumber := file.Line(elt.End())

		if nodeStartLineNumber > selection.LineNumber || selection.LineNumber > nodeEndLineNumber {
			continue
		}

		testName, found := decider.findTest(elt)
		if !found {
			continue
		}

		return testName
	}

	return ""
}
