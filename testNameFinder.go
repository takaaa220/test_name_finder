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

	selectedBasicLit, found := findSelectedStringBasicLit(fileNode, file, selection)

	var selectedText string
	if found {
		selectedText = selectedBasicLit.Value[1 : len(selectedBasicLit.Value)-1]
	}

	decl, found := findTargetTestFuncDecl(fileNode, file, selection)
	if !found {
		return nil, fmt.Errorf("failed to find target test function declaration")
	}

	return &TestName{
		FuncName: decl.Name.Name,
		TestCase: selectedText,
	}, nil
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
