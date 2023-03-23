package app

import (
	"context"
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

	// Execute all the pre required requests
	for _, preRequest := range request.PreRequests {
		_, _, err := a.executeRequest(ctx, a.HTTPTemplate.Requests[preRequest.Name], opts.Verbose)
		if err != nil {
			return err
		}
	}

	_, response, err := a.executeRequest(ctx, request, opts.Verbose)
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

func (a *app) executeRequest(ctx context.Context, request models.HttpRequest, verbose bool) (*http.Request, *models.HttpResponse, error) {
	replaceJsonBody, err := replaceWithUserInput(request.JsonBody)
	if err != nil {
		return nil, nil, err
	}
	request.JsonBody = replaceJsonBody

	httpReq, err := a.buildRequest(ctx, request)
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

	if verbose {
		a.logHttpResponse(ctx, request, httpResp)
	}

	bodyBytes, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, nil, err
	}

	if verbose {
		fmt.Println(string(bodyBytes))
	}

	return httpReq, &models.HttpResponse{
		RawResponse: httpResp,
		RawBody:     bodyBytes,
	}, nil
}

// buildRequest builds a http request from the request template
func (a *app) buildRequest(ctx context.Context, request models.HttpRequest) (*http.Request, error) {
	request.Sanitize()

	// Prepare request URL
	reqURL := url.URL{
		Scheme: "http",
		Host:   fmt.Sprintf("%s:%d", a.HTTPTemplate.Config.Host, a.HTTPTemplate.Config.Port),
		Path:   request.Path,
	}

	httpReq, err := http.NewRequest(request.Method, reqURL.String(), strings.NewReader(request.JsonBody))
	if err != nil {
		return nil, err
	}

	if request.JsonBody != "" {
		httpReq.Header.Add("Content-Type", "application/json")
	}

	for key, value := range request.Headers {
		httpReq.Header.Add(key, value)
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

func (a *app) logHttpResponse(ctx context.Context, request models.HttpRequest, httpResp *http.Response) {
	c := color.New(color.FgMagenta)

	c.Println("\n<<< Response")
	c.Println("------------")
	c.Println(httpResp.Status)
	for key, value := range httpResp.Header {
		c.Printf("%s: %s\n", key, strings.Join(value, "; "))
	}
}
