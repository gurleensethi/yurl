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

	request.Sanitize()

	// Prepare request URL
	reqURL := url.URL{
		Scheme: "http",
		Host:   fmt.Sprintf("%s:%d", a.HTTPTemplate.Config.Host, a.HTTPTemplate.Config.Port),
		Path:   request.Path,
	}

	// Prepare http request
	body, err := replaceWithUserInput(request.JsonBody)
	if err != nil {
		return err
	}

	httpReq, err := http.NewRequest(request.Method, reqURL.String(), strings.NewReader(body))
	if err != nil {
		return err
	}

	if request.JsonBody != "" {
		httpReq.Header.Add("Content-Type", "application/json")
	}

	for key, value := range request.Headers {
		httpReq.Header.Add(key, value)
	}

	// Execute request
	if opts.Verbose {
		fmt.Println("\nRequest\n-------")
		fmt.Printf("%s %s\n", request.Method, reqURL.String())
		for headerName, headerValue := range httpReq.Header {
			fmt.Printf("%s: %s\n", headerName, strings.Join(headerValue, ";"))
		}
	}

	httpClient := http.Client{}
	resp, err := httpClient.Do(httpReq)
	if err != nil {
		return err
	}

	// Print response
	if opts.Verbose {
		fmt.Println("\nResponse\n--------")
		fmt.Println(resp.Status)
		for key, value := range resp.Header {
			fmt.Printf("%s: %s\n", key, value)
		}
	}

	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	fmt.Println(string(bodyBytes))

	return nil
}
