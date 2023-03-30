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

	"github.com/gurleensethi/yurl/pkg/logger"
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

	ErrParsingExports = errors.New("error parsing exports")
)

type app struct {
	HTTPTemplate models.HttpTemplate
	FileVars     models.Variables
}

func New() *app {
	return &app{
		FileVars: make(models.Variables),
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
			&cli.BoolFlag{
				Name:    "list-variables",
				Usage:   "list all variables in the request",
				Aliases: []string{"lv", "list-vars"},
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
			{
				Name:    "request",
				Aliases: []string{"req"},
				Usage:   "make a request on the fly. ",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "method",
						Usage:   "http method",
						Aliases: []string{"m"},
					},
					&cli.StringFlag{
						Name:    "path",
						Usage:   "path of the request",
						Aliases: []string{"p"},
					},
					&cli.StringFlag{
						Name:    "body",
						Usage:   "body of the request",
						Aliases: []string{"b"},
					},
					&cli.StringFlag{
						Name:    "json",
						Usage:   "json body of the request",
						Aliases: []string{"j"},
					},
					&cli.StringSliceFlag{
						Name:    "header",
						Usage:   "header of the request",
						Aliases: []string{"head"},
					},
					// TODO: add ability to add pre requests
				},
				Action: func(cliCtx *cli.Context) error {
					method := cliCtx.String("method")
					path := cliCtx.String("path")
					body := cliCtx.String("body")
					jsonBody := cliCtx.String("json")
					headers := cliCtx.StringSlice("header")

					// Parse headers.
					// Raw format: "key:value"
					parsedHeaders := make(map[string]string)
					for _, header := range headers {
						parts := strings.Split(header, ":")
						if len(parts) != 2 {
							return fmt.Errorf("invalid header: %s", header)
						}
						parts[0] = strings.TrimSpace(parts[0])
						parts[1] = strings.TrimSpace(parts[1])
						parsedHeaders[parts[0]] = parts[1]
					}

					variables := a.FileVars

					// Parse variables from command line
					for _, variable := range cliCtx.StringSlice("variable") {
						parts := strings.Split(variable, "=")
						if len(parts) >= 2 {
							key := parts[0]
							value := strings.Join(parts[1:], "=")

							variables[key] = models.Variable{
								Value:  value,
								Source: models.VariableSourceCLI,
							}
						}
					}

					httpTemplateRequest := models.HttpRequestTemplate{
						Method:   method,
						Path:     path,
						Body:     body,
						JsonBody: jsonBody,
						Headers:  parsedHeaders,
					}

					return a.ExecuteRequest(cliCtx.Context, httpTemplateRequest, ExecuteRequestOpts{
						Verbose:   cliCtx.Bool("verbose"),
						Variables: variables,
					})
				},
			},
		},
		Before: func(cliCtx *cli.Context) error {
			filePath := cliCtx.String("file")
			if filePath == "" {
				filePath = DefaultHTTPYamlFile
			}

			err := a.parseHTTPYamlFile(cliCtx.Context, filePath)
			if err != nil {
				return err
			}

			variablesFilePaths := cliCtx.StringSlice("variable-file")

			err = a.parseVariablesFromFiles(cliCtx.Context, variablesFilePaths)
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
		Action: func(cliCtx *cli.Context) error {
			if cliCtx.Args().Len() == 0 {
				fmt.Println("Use `yurl -h` for help")
				return nil
			}

			name := cliCtx.Args().First()
			request, ok := a.HTTPTemplate.Requests[name]
			if !ok {
				return errors.New("request not found")
			}

			variables := a.FileVars

			// Parse variables from command line
			for _, variable := range cliCtx.StringSlice("variable") {
				parts := strings.Split(variable, "=")
				if len(parts) >= 2 {
					key := parts[0]
					value := strings.Join(parts[1:], "=")
					variables[key] = models.Variable{
						Value:  value,
						Source: models.VariableSourceCLI,
					}
				}
			}

			// List variables in the http request
			listVariables := cliCtx.Bool("list-variables")
			if listVariables {
				return a.ListRequestVariables(cliCtx.Context, request)
			}

			return a.ExecuteRequest(cliCtx.Context, request, ExecuteRequestOpts{
				Verbose:   cliCtx.Bool("verbose"),
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

	// Set name for each request
	for name, req := range a.HTTPTemplate.Requests {
		req.Name = name
		a.HTTPTemplate.Requests[name] = req
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
				key := parts[0]
				value := strings.Join(parts[1:], "=")
				a.FileVars[key] = models.Variable{
					Value:  value,
					Source: models.VariableSourceVarFile,
				}
			}
		}
	}

	return nil
}

type ExecuteRequestOpts struct {
	Verbose   bool
	Variables models.Variables
}

func (a *app) ExecuteRequest(ctx context.Context, request models.HttpRequestTemplate, opts ExecuteRequestOpts) error {
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
			vars[key] = models.Variable{
				Value:  value,
				Source: models.VariableSourceExports,
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

func (a *app) ListRequestVariables(ctx context.Context, request models.HttpRequestTemplate) error {
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

func (a *app) executeRequest(ctx context.Context, requestTemplate models.HttpRequestTemplate, vars models.Variables, verbose bool) (*http.Request, *models.HttpResponse, error) {
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
			var parsedBody map[string]any
			exportsErr = json.Unmarshal([]byte(bodyBytes), &parsedBody)
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

// buildRequest builds a http request from the request template
func (a *app) buildRequest(ctx context.Context, request models.HttpRequestTemplate, vars models.Variables) (*models.HttpRequest, error) {
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
	}, nil
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
		fmt.Println("  ", styles.Url.Copy().Bold(false).Render(request.Method+" "+request.Path))

		if i < len(keys)-1 {
			fmt.Println()
		}
	}

	return nil
}
