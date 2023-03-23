package app

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/gurleensethi/yurl/pkg/models"
)

var (
	inputRegex = regexp.MustCompile(`{{\s+?([a-zA-Z]+)\s+?}}`)
)

func replaceVariables(s string, vars models.Variables) (string, error) {
	matches := inputRegex.FindAllStringSubmatch(s, -1)
	if len(matches) == 0 {
		return s, nil
	}

	for _, match := range matches {
		key := match[1]

		// Check if variable is present in vars
		if value, ok := vars[key]; ok {
			s = strings.ReplaceAll(s, match[0], fmt.Sprintf("%v", value))
			continue
		}

		// Variable not present in vars, prompt user for input
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
