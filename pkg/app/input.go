package app

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/gurleensethi/yurl/pkg/models"
	"github.com/gurleensethi/yurl/pkg/styles"
)

var (
	inputRegex = regexp.MustCompile(`{{\s+?([a-zA-Z0-9]+):?(string|int|float|bool)?\s+?}}`)
)

func replaceVariables(s string, vars models.Variables) (string, error) {
	matches := inputRegex.FindAllStringSubmatch(s, -1)
	if len(matches) == 0 {
		return s, nil
	}

	for _, match := range matches {
		key := match[1]
		inputType := "" // Input type is the type to be enforced for the input
		if len(match) > 2 {
			inputType = match[2]
		}

		// Check if variable is present in vars
		if value, ok := vars[key]; ok {
			s = strings.ReplaceAll(s, match[0], fmt.Sprintf("%v", value))
			continue
		}

		// Variable not present in vars, prompt user for input
		label := styles.PrimaryText.Render(fmt.Sprintf("`%s`", key))
		if inputType != "" {
			label += styles.SecondaryText.Render(fmt.Sprintf(" (%s)", inputType))
		}

		input, err := getUserInput(label)
		if err != nil {
			return "", err
		}

		switch inputType {
		case "int":
			_, err := strconv.Atoi(input)
			if err != nil {
				return "", fmt.Errorf("input for `%s` must be of type int", key)
			}
		case "float":
			_, err := strconv.ParseFloat(input, 64)
			if err != nil {
				return "", fmt.Errorf("input for `%s` must be of type float", key)
			}
		case "bool":
			_, err := strconv.ParseBool(input)
			if err != nil {
				return "", fmt.Errorf("input for `%s` must be of type bool", key)
			}
		case "string":
		case "":
			// Esacpe quotes
			input = strings.ReplaceAll(input, `"`, `\"`)
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
	fmt.Fprintf(os.Stderr, "Enter %s: ", label)

	reader := bufio.NewReader(os.Stdin)

	line, _, err := reader.ReadLine()
	if err != nil {
		return "", err
	}

	return string(line), err
}
