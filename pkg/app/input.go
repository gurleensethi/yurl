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

// replaceWithUserInput replaces all instances of {{ <key> }} with user input.
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

// getUserInput prompts user for input and returns it.
func getUserInput(label string) (string, error) {
	// When piping output to other programs, we don't want the intput promots to be a part of it.
	// For example: `yurl Login | jq`, if output is sent to stdin, the input prompts will be part of input
	// to jq. We don't want that.
	fmt.Fprintf(os.Stderr, "Enter `%s`: ", label)

	reader := bufio.NewReader(os.Stdin)

	line, _, err := reader.ReadLine()
	if err != nil {
		return "", err
	}

	return string(line), err
}
