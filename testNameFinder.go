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

	// selectedBasicLit, found := findSelectedStringBasicLit(fileNode, file, selection)
	testCase, _ := findTestCaseInTestFuncDecl(file, decl, selection)

	return &TestName{
		FuncName: decl.Name.Name,
		TestCase: testCase,
	}, nil
}

func findTestCaseInTestFuncDecl(file *token.File, funcDecl *ast.FuncDecl, selection Selection) (string, bool) {
	// FuncDecl直下のt.Runを行っているRangeStmtを探す
	var rangeStmtForRunTest *ast.RangeStmt
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

			if selectorIdent.Name != "t" {
				continue
			}

			if selectorExpr.Sel.Name != "Run" {
				continue
			}

			rangeStmtForRunTest = rStmt
			break
		}

		if rangeStmtForRunTest != nil {
			break
		}
	}

	if rangeStmtForRunTest == nil {
		return "", false
	}

	tableVariableIdent, ok := rangeStmtForRunTest.X.(*ast.Ident)
	if !ok {
		return "", false
	}

	tableVariableName := tableVariableIdent.Name

	// FuncDecl直下のLhs[0].Ident.NameがtableVariableNameであるAssignStmtを探す
	var assignStmtForTableVariable *ast.AssignStmt
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

		assignStmtForTableVariable = assignStmt
	}

	if assignStmtForTableVariable == nil {
		return "", false
	}

	// assignStmtForTableVariableからテスト名の特定方法を決定する
	compositeLit, ok := assignStmtForTableVariable.Rhs[0].(*ast.CompositeLit)
	if !ok {
		return "", false
	}

	type testNameDecider struct {
		isSlice       bool
		testNameField string
	}
	var decider *testNameDecider
	if len(assignStmtForTableVariable.Rhs) == 0 {
		return "", false
	}

	if arrayType, ok := compositeLit.Type.(*ast.ArrayType); ok {
		if _, ok := arrayType.Elt.(*ast.StructType); ok {
			decider = &testNameDecider{
				isSlice:       true,
				testNameField: "name",
			}
		}
	}
	if mapType, ok := compositeLit.Type.(*ast.MapType); ok {
		if _, ok := mapType.Value.(*ast.StructType); ok {
			decider = &testNameDecider{
				isSlice: false,
			}
		}
	}

	if decider == nil {
		return "", false
	}

	// テスト名を特定する
	var testCase string
	for _, elt := range compositeLit.Elts {
		nodeStartLineNumber := file.Line(elt.Pos())
		nodeEndLineNumber := file.Line(elt.End())

		if nodeStartLineNumber > selection.LineNumber || selection.LineNumber > nodeEndLineNumber {
			continue
		}

		if !decider.isSlice {
			keyValueExpr, ok := elt.(*ast.KeyValueExpr)
			if !ok {
				continue
			}

			keyBasicLit, ok := keyValueExpr.Key.(*ast.BasicLit)
			if !ok {
				continue
			}

			if keyBasicLit.Kind != token.STRING {
				continue
			}

			testCase = strings.Trim(keyBasicLit.Value, "\"")
			break
		}

		compositeLit, ok := elt.(*ast.CompositeLit)
		if !ok {
			continue
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

			if keyIdent.Name != decider.testNameField {
				continue
			}

			basicLit, ok := keyValueExpr.Value.(*ast.BasicLit)
			if !ok {
				continue
			}

			if basicLit.Kind != token.STRING {
				continue
			}

			testCase = strings.Trim(basicLit.Value, "\"")
			break
		}
	}

	return testCase, testCase != ""
}

func findSelectedStringBasicLit(fileNode *ast.File, file *token.File, selection Selection) (*ast.BasicLit, bool) {
	var basicLit *ast.BasicLit

	ast.Inspect(fileNode, func(node ast.Node) bool {
		b, ok := node.(*ast.BasicLit)
		if !ok {
			return true
		}

		lineStartPos := file.LineStart(selection.LineNumber)

		start := lineStartPos + token.Pos(selection.StartCursor)
		end := lineStartPos + token.Pos(selection.EndCursor)

		if b.Pos() <= start && end <= b.End() {
			if b.Kind == token.STRING {
				basicLit = b
			}

			return false
		}

		return true
	})

	if basicLit == nil {
		return nil, false
	}

	return basicLit, true
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
