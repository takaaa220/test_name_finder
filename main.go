package testNameFinder

import (
	"bufio"
	"fmt"
	"os"
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

func FindTestName(filePath string, selection Selection) (string, error) {
	if msg, ok := selection.Valid(); !ok {
		return "", fmt.Errorf(msg)
	}

	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	currentLine := 0

	for scanner.Scan() {
		currentLine++
		if currentLine != selection.LineNumber {
			continue
		}

		line := scanner.Text()
		if len(line) < selection.EndCursor {
			return "", fmt.Errorf("cursor out of range")
		}

		return line[selection.StartCursor:selection.EndCursor], nil
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	return "", fmt.Errorf("not found")
}
