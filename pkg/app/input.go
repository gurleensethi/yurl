package app

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

var (
	inputRegex = regexp.MustCompile(`{{\s+?([a-zA-Z]+)\s+?}}`)
)

func replaceWithUserInput(s string) (string, error) {
	matches := inputRegex.FindAllStringSubmatch(s, -1)
	if len(matches) == 0 {
		return s, nil
	}

	for _, match := range matches {
		key := match[1]
		input, err := getUserInput(key)
		if err != nil {
			return "", err
		}
		s = strings.ReplaceAll(s, match[0], input)
	}

	return s, nil
}

func getUserInput(label string) (string, error) {
	fmt.Fprintf(os.Stderr, "Enter `%s`: ", label)
	reader := bufio.NewReader(os.Stdin)
	line, _, err := reader.ReadLine()
	if err != nil {
		return "", err
	}
	return string(line), err
}
