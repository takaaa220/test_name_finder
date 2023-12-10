package main

import (
	"flag"
	"fmt"
	"os"

	testNameFinder "github.com/takaaa220/test-name-finder"
)

func main() {
	// FilePath, LineNumber, StartCursor, EndCursor をCLIで受け取る
	var filePath string
	var lineNumber int
	var startCursor int
	var endCursor int

	flag.StringVar(&filePath, "f", "", "[required] absolute file path")
	flag.IntVar(&lineNumber, "l", -1, "[required] line number (1 start)")
	flag.IntVar(&startCursor, "s", -1, "[required] start cursor (0 start)")
	flag.IntVar(&endCursor, "e", -1, "[required] end cursor (0 start)")

	flag.Parse()

	// 必須入力のチェック
	if filePath == "" || lineNumber == -1 || startCursor == -1 || endCursor == -1 {
		fmt.Println("Required flags are not set.")
		flag.Usage()

		os.Exit(1)
		return
	}

	testName, err := testNameFinder.FindTestName(filePath, testNameFinder.Selection{
		LineNumber:  lineNumber,
		StartCursor: startCursor,
		EndCursor:   endCursor,
	})
	if err != nil {
		panic(fmt.Errorf("failed to find test name: %w", err))
	}

	fmt.Println(testName)
}
