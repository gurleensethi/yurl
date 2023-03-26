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
	"sort"
	"strings"

	"github.com/gurleensethi/yurl/pkg/models"
	"github.com/gurleensethi/yurl/pkg/styles"
	"github.com/urfave/cli/v2"
	"github.com/yalp/jsonpath"
	"gopkg.in/yaml.v3"
)

var (
	DefaultHTTPYamlFile = "http.yaml"

	UsageText = `
Use default file (http.yaml)

  yurl <request name>

Specify a request file

  yurl -f requests-file.yaml <request name>

Use variables

  yurl -var email=test@test.com <request name>

Use a variable file

  yurl -var-file=local.vars <request name>
`
)

type app struct {
	HTTPTemplate models.HttpTemplate
	FileVars     map[string]any
}

func New() *app {
	return &app{
		FileVars: make(map[string]any),
	}
}

func (a *app) BuildCliApp() *cli.App {
	return &cli.App{
		Name:        "yurl",
		Description: "Write your http requests in yaml.",
		Usage:       "Make http requests from command line using yaml.",
		UsageText:   UsageText,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "verbose",
				Aliases: []string{"v"},
				Usage:   "verbose output",
			},
			&cli.StringSliceFlag{
				Name:    "variable",
				Usage:   "variable to be used in the request. ",
				Aliases: []string{"var"},
			},
			&cli.StringSliceFlag{
				Name:    "variable-file",
				Usage:   "loads variables from the given file.",
				Aliases: []string{"var-file"},
			},
			&cli.StringFlag{
				Name:    "file",
				Usage:   "path of file to read http requests from",
				Aliases: []string{"f"},
			},
		},
		Commands: []*cli.Command{
			{
				Name:    "list-requests",
				Aliases: []string{"ls"},
				Usage:   "list all requests in the requests (yaml) file",
				Action: func(c *cli.Context) error {
					return a.ListRequests(c.Context)
				},
			},
		},
		Before: func(c *cli.Context) error {
			filePath := c.String("file")
			if filePath == "" {
				filePath = DefaultHTTPYamlFile
			}

			err := a.parseHTTPYamlFile(c.Context, filePath)
			if err != nil {
				return err
			}

			variablesFilePaths := c.StringSlice("variable-file")

			err = a.parseVariablesFromFiles(c.Context, variablesFilePaths)
			if err != nil {
				return err
			}

			a.HTTPTemplate.Sanitize()

			err = a.HTTPTemplate.Validate()
			if err != nil {
				return err
			}

			return nil
		},
		Action: func(c *cli.Context) error {
			if c.Args().Len() == 0 {
				fmt.Println("Use `yurl -h` for help")
				return nil
			}

			variables := a.FileVars

			// Parse variables from command line
			for _, variable := range c.StringSlice("variable") {
				parts := strings.Split(variable, "=")
				if len(parts) >= 2 {
					variables[parts[0]] = strings.Join(parts[1:], "=")
				}
			}

			return a.ExecuteRequest(c.Context, c.Args().First(), ExecuteRequestOpts{
				Verbose:   c.Bool("verbose"),
				Variables: variables,
			})
		},
	}
}

func (a *app) parseHTTPYamlFile(ctx context.Context, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		if err == os.ErrNotExist {
			return fmt.Errorf("no request file found at %s", filePath)
		}
		return err
	}

	err = yaml.NewDecoder(file).Decode(&a.HTTPTemplate)
	if err != nil {
		return err
	}

	return nil
}

func (a *app) parseVariablesFromFiles(ctx context.Context, filePaths []string) error {
	for _, filePath := range filePaths {
		file, err := os.Open(filePath)
		if err != nil {
			if err == os.ErrNotExist {
				return fmt.Errorf("no request file found at %s", filePath)
			}
			return err
		}

		reader := bufio.NewReader(file)

		for {
			lineBytes, _, err := reader.ReadLine()
			if err == io.EOF {
				break
			} else if err != nil {
				return err
			}

			line := string(lineBytes)
			parts := strings.Split(line, "=")
			if len(parts) >= 2 {
				a.FileVars[parts[0]] = strings.Join(parts[1:], "=")
			}
		}
	}

	return nil
}

type ExecuteRequestOpts struct {
	Verbose   bool
	Variables models.Variables
}

func (a *app) ExecuteRequest(ctx context.Context, name string, opts ExecuteRequestOpts) error {
	request, ok := a.HTTPTemplate.Requests[name]
	if !ok {
		return errors.New("request not found")
	}

	request.Name = name
	request.Sanitize()

	err := request.Validate(&a.HTTPTemplate)
	if err != nil {
		return err
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
			vars[key] = value
		}
	}

	_, response, err := a.executeRequest(ctx, request, vars, opts.Verbose)
	if err != nil {
		return err
	}

	// When verbose is set executeRequest will print the request and response.
	// So we need to print it only when verbose is not set.
	if !opts.Verbose {
		fmt.Println(string(response.RawBody))
	}

	return nil
}

func (a *app) executeRequest(ctx context.Context, requestTemplate models.HttpTemplateRequest, vars models.Variables, verbose bool) (*http.Request, *models.HttpResponse, error) {
	httpRequest, err := a.buildRequest(ctx, requestTemplate, vars)
	if err != nil {
		return nil, nil, err
	}

	httpReq := httpRequest.RawRequest

	if verbose {
		a.logHttpRequest(ctx, httpRequest)
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

	// Parse out exports
	exports := make(map[string]any)

	for name, export := range requestTemplate.Exports {
		if export.JSON != "" {
			var parsedBody map[string]any
			err := json.Unmarshal([]byte(bodyBytes), &parsedBody)
			if err != nil {
				return nil, nil, err
			}

			value, err := jsonpath.Read(parsedBody, export.JSON)
			if err != nil {
				return nil, nil, err
			}

			exports[name] = value
		}
	}

	httpResponse := &models.HttpResponse{
		RawResponse: httpResp,
		RawBody:     bodyBytes,
		Exports:     exports,
	}

	if verbose {
		a.logHttpResponse(ctx, requestTemplate, httpResponse)
	}

	return httpReq, httpResponse, nil
}

// buildRequest builds a http request from the request template
func (a *app) buildRequest(ctx context.Context, request models.HttpTemplateRequest, vars models.Variables) (*models.HttpRequest, error) {
	request.Sanitize()

	// Prepare request URL
	replacedPath, err := replaceVariables(request.Path, vars)
	if err != nil {
		return nil, err
	}

	rawURL := fmt.Sprintf("http://%s:%d%s", a.HTTPTemplate.Config.Host, a.HTTPTemplate.Config.Port, replacedPath)
	reqURL, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}

	reqURL.Scheme = a.HTTPTemplate.Config.Scheme

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
	}, nil
}

func (a *app) logHttpRequest(ctx context.Context, request *models.HttpRequest) {
	fmt.Println(styles.SectionHeader.Render("Request"))

	protocol := styles.Url.Render(request.RawRequest.Proto)
	method := styles.Url.Render(request.RawRequest.Method)
	completeUrl := styles.Url.Render(request.RawRequest.URL.String())
	fmt.Printf("%s %s %s\n", method, completeUrl, protocol)

	for headerName, headerValue := range request.RawRequest.Header {
		fmt.Printf("%s: %s\n", styles.HeaderName.Render(headerName), strings.Join(headerValue, ";"))
	}

	fmt.Println(request.Template.JsonBody)
}

func (a *app) logHttpResponse(ctx context.Context, request models.HttpTemplateRequest, httpResponse *models.HttpResponse) {
	fmt.Println(styles.SectionHeader.Render("Response"))

	protocol := styles.Url.Render(httpResponse.RawResponse.Proto)
	status := styles.Url.Render(httpResponse.RawResponse.Status)
	fmt.Println(protocol, status)

	for key, value := range httpResponse.RawResponse.Header {
		fmt.Printf("%s: %s\n", styles.HeaderName.Render(key), strings.Join(value, "; "))
	}
	fmt.Println(string(httpResponse.RawBody))

	fmt.Println(styles.SectionHeader.Render("Exports"))

	if len(httpResponse.Exports) == 0 {
		fmt.Println("  No exports")
	}
	for key, value := range httpResponse.Exports {
		fmt.Println(key, ":", value)
	}
}

func (a *app) ListRequests(ctx context.Context) error {
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
		fmt.Println("  ", styles.Url.Bold(false).Render(request.Method+" "+request.Path))

		if i < len(keys)-1 {
			fmt.Println()
		}
	}

	return nil
}
