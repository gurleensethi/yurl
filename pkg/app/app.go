package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/gurleensethi/yurl/pkg/models"
	"github.com/urfave/cli/v2"
	"github.com/yalp/jsonpath"
	"gopkg.in/yaml.v3"
)

var (
	DefaultHTTPYamlFile = "http.yaml"
)

type app struct {
	HTTPTemplate models.HttpTemplate
}

func New() *app {
	return &app{}
}

func (a *app) BuildCliApp() *cli.App {
	return &cli.App{
		Name:        "yurl",
		Description: "Write your http requests in yaml.",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "verbose",
				Aliases: []string{"v"},
				Usage:   "verbose output",
			},
		},
		Before: func(c *cli.Context) error {
			return a.ParseHTTPYamlFile(c.Context)
		},
		Action: func(c *cli.Context) error {
			if c.Args().Len() == 0 {
				fmt.Println("Use `yurl -h` for help")
				return nil
			}

			return a.ExecuteRequest(c.Context, c.Args().First(), ExecuteRequestOpts{
				Verbose: c.Bool("verbose"),
			})
		},
	}
}

func (a *app) ParseHTTPYamlFile(ctx context.Context) error {
	file, err := os.Open(DefaultHTTPYamlFile)
	if err != nil {
		return err
	}

	err = yaml.NewDecoder(file).Decode(&a.HTTPTemplate)
	if err != nil {
		return err
	}

	return nil
}

type ExecuteRequestOpts struct {
	Verbose bool
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

	vars := make(map[string]any)

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

func (a *app) executeRequest(ctx context.Context, request models.HttpRequest, vars models.Variables, verbose bool) (*http.Request, *models.HttpResponse, error) {
	httpReq, err := a.buildRequest(ctx, request, vars)
	if err != nil {
		return nil, nil, err
	}

	if verbose {
		a.logHttpRequest(ctx, request, httpReq)
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

	for name, export := range request.Exports {
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
		a.logHttpResponse(ctx, request, httpResponse)
	}

	return httpReq, httpResponse, nil
}

// buildRequest builds a http request from the request template
func (a *app) buildRequest(ctx context.Context, request models.HttpRequest, vars models.Variables) (*http.Request, error) {
	request.Sanitize()

	replacedJsonBody, err := replaceVariables(request.JsonBody, vars)
	if err != nil {
		return nil, err
	}
	request.JsonBody = replacedJsonBody

	// Prepare request URL
	replacedPath, err := replaceVariables(request.Path, vars)
	if err != nil {
		return nil, err
	}

	reqURL := url.URL{
		Scheme: "http",
		Host:   fmt.Sprintf("%s:%d", a.HTTPTemplate.Config.Host, a.HTTPTemplate.Config.Port),
		Path:   replacedPath,
	}

	httpReq, err := http.NewRequest(request.Method, reqURL.String(), strings.NewReader(request.JsonBody))
	if err != nil {
		return nil, err
	}

	if request.JsonBody != "" {
		httpReq.Header.Add("Content-Type", "application/json")
	}

	for key, value := range request.Headers {
		replacedValue, err := replaceVariables(value, vars)
		if err != nil {
			return nil, err
		}

		httpReq.Header.Add(key, replacedValue)
	}

	return httpReq, nil
}

func (a *app) logHttpRequest(ctx context.Context, request models.HttpRequest, httpReq *http.Request) {
	c := color.New(color.FgHiYellow)

	c.Println("\n>>> Request")
	c.Println("-----------")
	c.Printf("%s %s\n", request.Method, httpReq.URL.String())
	for headerName, headerValue := range httpReq.Header {
		c.Printf("%s: %s\n", headerName, strings.Join(headerValue, ";"))
	}
	c.Println(request.JsonBody)
}

func (a *app) logHttpResponse(ctx context.Context, request models.HttpRequest, httpResponse *models.HttpResponse) {
	c := color.New(color.FgMagenta)

	c.Println("\n<<< Response")
	c.Println("------------")
	c.Println(httpResponse.RawResponse.Status)
	for key, value := range httpResponse.RawResponse.Header {
		c.Printf("%s: %s\n", key, strings.Join(value, "; "))
	}
	c.Println("\nExports:")
	c.Println("--------")
	if len(httpResponse.Exports) == 0 {
		c.Println("  No exports")
	}
	for key, value := range httpResponse.Exports {
		c.Println("  ", key, ":", value)
	}
}
