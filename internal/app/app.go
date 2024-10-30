package app

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/gurleensethi/yurl/internal/logger"
	"github.com/gurleensethi/yurl/internal/variable"
	"github.com/gurleensethi/yurl/pkg/models"
	"github.com/gurleensethi/yurl/pkg/styles"
	"github.com/yalp/jsonpath"
)

var (
	ErrParsingExports = errors.New("error parsing exports")

	inputRegex = regexp.MustCompile(`{{\s+?([a-zA-Z0-9]+):?(string|int|float|bool)?\s+?}}`)
)

// App represents the main application for performing
// http requests and other operations.
type App struct {
	// HTTPTemplate represents main config file for yurl.
	HTTPTemplate models.HttpTemplate

	// Inital variables that app is initialized with.
	Variables variable.Variables
}

func New(template models.HttpTemplate, vars variable.Variables) *App {
	return &App{
		HTTPTemplate: template,
		Variables:    vars,
	}
}

type ExecuteRequestOpts struct {
	ListVariables bool
	Variables     variable.Variables
	Verbose       bool
}

func (a *App) ListRequests(ctx context.Context) error {
	keys := make([]string, 0, len(a.HTTPTemplate.Requests))
	for name := range a.HTTPTemplate.Requests {
		keys = append(keys, name)
	}

	sort.Strings(keys)

	for i, key := range keys {
		request := a.HTTPTemplate.Requests[key]

		name := styles.HeaderName.Render(key)
		description := styles.Description.Render(request.Description)

		fmt.Println(name, description)
		fmt.Println("  ", styles.Url.Copy().Bold(false).Render(request.Method+" "+request.Path))

		if i < len(keys)-1 {
			fmt.Println()
		}
	}

	return nil
}

func (a *App) ExecuteRequest(ctx context.Context, requestName string, opts ExecuteRequestOpts) error {
	request, ok := a.HTTPTemplate.Requests[requestName]
	if !ok {
		return errors.New("request not found")
	}

	request.Sanitize()

	err := request.Validate(&a.HTTPTemplate)
	if err != nil {
		return err
	}

	if opts.ListVariables {
		err := a.ListRequestVariables(ctx, request)
		if err != nil {
			return err
		}

		return nil
	}

	vars := opts.Variables

	// Execute all the pre required requests
	for _, preRequest := range request.PreRequests {
		_, httpResponse, err := a.executeRequest(ctx, a.HTTPTemplate.Requests[preRequest.Name], vars, opts.Verbose)
		if err != nil {
			return err
		}

		// Merge the exports from the pre request to the vars
		for key, value := range httpResponse.Exports {
			vars[key] = variable.Variable{
				Value:  value,
				Source: variable.SourceExports,
			}
		}

		if opts.Verbose {
			fmt.Println(styles.Divider.Render("------------------------------------------------------------------------------"))
		}
	}

	_, response, err := a.executeRequest(ctx, request, vars, opts.Verbose)
	if err != nil {
		return err
	}

	if !opts.Verbose {
		fmt.Println(string(response.RawBody))
	}

	return nil
}

func (a *App) executeRequest(ctx context.Context, requestTemplate models.HttpRequestTemplate, vars variable.Variables, verbose bool) (*http.Request, *models.HttpResponse, error) {
	httpRequest, err := a.buildRequest(ctx, requestTemplate, vars)
	if err != nil {
		return nil, nil, err
	}

	httpReq := httpRequest.RawRequest

	if verbose {
		logger.LogHttpRequest(ctx, httpRequest)
	}

	httpClient := http.Client{}
	httpResp, err := httpClient.Do(httpReq)
	if err != nil {
		return nil, nil, err
	}

	defer httpResp.Body.Close()

	bodyBytes, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, nil, err
	}

	httpResponse := &models.HttpResponse{
		Request:     httpRequest,
		RawResponse: httpResp,
		RawBody:     bodyBytes,
		Exports:     make(map[string]any),
	}

	// Capturing the exports error to process later on because regardless if there is an error
	// parsing the exports we still want to log the response.
	var exportsErr error

	for name, export := range requestTemplate.Exports {
		if export.JSON != "" {
			var parsedBody interface{}
			exportsErr = json.Unmarshal(bodyBytes, &parsedBody)
			if err != nil {
				exportsErr = err
				break
			}

			var value any
			value, err := jsonpath.Read(parsedBody, export.JSON)
			if err != nil {
				exportsErr = err
				break
			}

			httpResponse.Exports[name] = value
		}
	}
	if verbose {
		logger.LogHttpResponse(ctx, httpResponse)
	}
	if exportsErr != nil {
		return nil, nil, fmt.Errorf("%w: %w", ErrParsingExports, exportsErr)
	}

	return httpReq, httpResponse, nil
}

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

func (a *App) ListRequestVariables(ctx context.Context, request models.HttpRequestTemplate) error {
	// Find variables in the path
	pathVars := findVariables(request.Path)

	if len(pathVars) > 0 {
		fmt.Println(styles.PrimaryText.Render("Path"))

		for _, pathVar := range pathVars {
			fmt.Println("  " + pathVar)
		}
	}

	// Find variables in the headers
	headerVars := []string{}
	for _, header := range request.Headers {
		headerVars = append(headerVars, findVariables(header)...)
	}

	if len(headerVars) > 0 {
		fmt.Println(styles.PrimaryText.Render("Headers"))

		for _, headerVar := range headerVars {
			fmt.Println("  " + headerVar)
		}
	}

	// Find variables in the body
	bodyVars := findVariables(request.Body)
	bodyVars = append(bodyVars, findVariables(request.JsonBody)...)

	if len(bodyVars) > 0 {
		fmt.Println(styles.PrimaryText.Render("Body"))

		for _, bodyVar := range bodyVars {
			fmt.Println("  " + bodyVar)
		}
	}

	return nil
}

// buildRequest builds a http request from the request template
func (a *App) buildRequest(ctx context.Context, request models.HttpRequestTemplate, vars variable.Variables) (*models.HttpRequest, error) {
	request.Sanitize()

	// Prepare request URL
	replacedPath, err := replaceVariables(request.Path, vars)
	if err != nil {
		return nil, err
	}

	host := a.HTTPTemplate.Config.Host
	if a.HTTPTemplate.Config.Port != 0 {
		host = fmt.Sprintf("%s:%d", host, a.HTTPTemplate.Config.Port)
	}

	reqURL := url.URL{
		Host:   host,
		Scheme: a.HTTPTemplate.Config.Scheme,
		Path:   replacedPath,
	}

	reqURL.Scheme = a.HTTPTemplate.Config.Scheme

	// Prepare query params
	query := reqURL.Query()

	for key, value := range request.Query {
		replacedParam, err := replaceVariables(value, vars)
		if err != nil {
			return nil, err
		}

		query.Set(key, replacedParam)
	}

	reqURL.RawQuery = query.Encode()

	// Prepare request body
	body := ""
	bodyContentType := ""

	if request.Body != "" {
		replacedBody, err := replaceVariables(request.Body, vars)
		if err != nil {
			return nil, err
		}
		body = replacedBody
		request.Body = replacedBody
		bodyContentType = "text/plain"
	} else if request.JsonBody != "" {
		replacedJsonBody, err := replaceVariables(request.JsonBody, vars)
		if err != nil {
			return nil, err
		}
		body = replacedJsonBody
		request.JsonBody = replacedJsonBody
		bodyContentType = "application/json"
	}

	httpReq, err := http.NewRequest(request.Method, reqURL.String(), strings.NewReader(body))
	if err != nil {
		return nil, err
	}

	if bodyContentType != "" {
		httpReq.Header.Add("Content-Type", bodyContentType)
	}

	for key, value := range request.Headers {
		replacedValue, err := replaceVariables(value, vars)
		if err != nil {
			return nil, err
		}

		httpReq.Header.Set(key, replacedValue)
	}

	return &models.HttpRequest{
		RawRequest: httpReq,
		Template:   &request,
		Variables:  vars,
	}, nil
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
		if variable, ok := vars.Get(key); ok {
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
		vars.Add(variable.Variable{
			Value:  input,
			Source: variable.SourceInput,
		})
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
