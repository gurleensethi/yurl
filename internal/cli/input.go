package cli

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/gurleensethi/yurl/internal/variable"
	"github.com/gurleensethi/yurl/pkg/styles"
)

var (
	inputRegex = regexp.MustCompile(`{{\s+?([a-zA-Z0-9]+):?(string|int|float|bool)?\s+?}}`)
)

func findVariables(s string) []string {
	matches := inputRegex.FindAllStringSubmatch(s, -1)
	if len(matches) == 0 {
		return nil
	}

	vars := make([]string, 0)
	for _, match := range matches {
		variable := match[1]
		if len(match) > 2 && match[2] != "" {
			variable += " (" + match[2] + ")"
		}

		vars = append(vars, variable)
	}

	return vars
}

func replaceVariables(s string, vars variable.Variables) (string, error) {
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
		if variable, ok := vars[key]; ok {
			s = strings.ReplaceAll(s, match[0], fmt.Sprintf("%v", variable.Value))
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
			if input != "true" && input != "false" {
				return "", fmt.Errorf("input for `%s` must be of type bool", key)
			}
		case "string":
		case "":
			// Esacpe quotes
			input = strings.ReplaceAll(input, `"`, `\"`)
		}

		s = strings.ReplaceAll(s, match[0], input)

		// Add variable to vars
		vars[key] = variable.Variable{
			Value:  input,
			Source: variable.SourceInput,
		}
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
