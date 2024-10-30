package cli

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/gurleensethi/yurl/internal/app"
	"github.com/gurleensethi/yurl/internal/variable"
	"github.com/gurleensethi/yurl/pkg/models"
	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v3"
)

var (
	Version = "v0.3.0"

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

const (
	FlagVerbose       = "verbose"
	FlagVariable      = "variable"
	FlagVariableFile  = "variable-file"
	FlagFile          = "file"
	FlagListVariables = "list-variables"
	FlagPath          = "path"
)

type CliApp struct {
	app *app.App
}

func NewApp() *CliApp {
	return &CliApp{}
}

func (a *CliApp) Build() *cli.App {
	return &cli.App{
		Name:        "yurl",
		Description: "Write your http requests in yaml.",
		Usage:       "Make http requests from command line using yaml.",
		UsageText:   UsageText,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    FlagVerbose,
				Aliases: []string{"v"},
				Usage:   "verbose output",
			},
			&cli.StringSliceFlag{
				Name:    FlagVariable,
				Usage:   "variable to be used in the request. ",
				Aliases: []string{"var"},
			},
			&cli.StringSliceFlag{
				Name:    FlagVariableFile,
				Usage:   "loads variables from the given file.",
				Aliases: []string{"var-file"},
			},
			&cli.StringFlag{
				Name:    FlagFile,
				Usage:   "path of file to read http requests from",
				Aliases: []string{"f"},
			},
			&cli.BoolFlag{
				Name:    FlagListVariables,
				Usage:   "list all variables in the request",
				Aliases: []string{"lv", "list-vars"},
			},
		},
		Commands: []*cli.Command{
			{
				Name:  "init",
				Usage: "Init a new config file",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  FlagPath,
						Usage: "destination to initalize new configuration at",
					},
				},
				Action: InitConfigFile,
			},
			{
				Name:  "version",
				Usage: "print the version of yurl",
				Action: func(c *cli.Context) error {
					fmt.Println(Version)
					return nil
				},
			},
			{
				Name:    "list-requests",
				Aliases: []string{"ls"},
				Usage:   "list all requests in the requests (yaml) file",
				Action: func(c *cli.Context) error {
					return a.app.ListRequests(c.Context)
				},
			},
			// {
			// 	Name:    "request",
			// 	Aliases: []string{"req"},
			// 	Usage:   "make a request on the fly. ",
			// 	Flags: []cli.Flag{
			// 		&cli.StringFlag{
			// 			Name:    "method",
			// 			Usage:   "http method",
			// 			Aliases: []string{"m"},
			// 		},
			// 		&cli.StringFlag{
			// 			Name:    "path",
			// 			Usage:   "path of the request",
			// 			Aliases: []string{"p"},
			// 		},
			// 		&cli.StringFlag{
			// 			Name:    "body",
			// 			Usage:   "body of the request",
			// 			Aliases: []string{"b"},
			// 		},
			// 		&cli.StringFlag{
			// 			Name:    "json",
			// 			Usage:   "json body of the request",
			// 			Aliases: []string{"j"},
			// 		},
			// 		&cli.StringSliceFlag{
			// 			Name:    "header",
			// 			Usage:   "header of the request",
			// 			Aliases: []string{"head"},
			// 		},
			// 		// TODO: add ability to add pre requests
			// 	},
			// 	Action: func(cliCtx *cli.Context) error {
			// 		method := cliCtx.String("method")
			// 		path := cliCtx.String("path")
			// 		body := cliCtx.String("body")
			// 		jsonBody := cliCtx.String("json")
			// 		headers := cliCtx.StringSlice("header")

			// 		// Parse headers.
			// 		// Raw format: "key:value"
			// 		parsedHeaders := make(map[string]string)
			// 		for _, header := range headers {
			// 			parts := strings.Split(header, ":")
			// 			if len(parts) != 2 {
			// 				return fmt.Errorf("invalid header: %s", header)
			// 			}
			// 			parts[0] = strings.TrimSpace(parts[0])
			// 			parts[1] = strings.TrimSpace(parts[1])
			// 			parsedHeaders[parts[0]] = parts[1]
			// 		}

			// 		variables := a.FileVars

			// 		// Parse variables from command line
			// 		for _, variable := range cliCtx.StringSlice("variable") {
			// 			parts := strings.Split(variable, "=")
			// 			if len(parts) >= 2 {
			// 				key := parts[0]
			// 				value := strings.Join(parts[1:], "=")

			// 				variables[key] = models.Variable{
			// 					Value:  value,
			// 					Source: models.VariableSourceCLI,
			// 				}
			// 			}
			// 		}

			// 		httpTemplateRequest := models.HttpRequestTemplate{
			// 			Method:   method,
			// 			Path:     path,
			// 			Body:     body,
			// 			JsonBody: jsonBody,
			// 			Headers:  parsedHeaders,
			// 		}

			// 		return a.app.ExecuteRequest(cliCtx.Context, httpTemplateRequest, app.ExecuteRequestOpts{
			// 			Verbose:   cliCtx.Bool("verbose"),
			// 			Variables: variables,
			// 		})
			// 	},
			// },
		},
		Before: func(cliCtx *cli.Context) error {
			// Don't try to load http.yaml file if command is an
			// initialization command.
			if cliCtx.NArg() >= 1 && cliCtx.Args().First() == "init" {
				return nil
			}

			filePath := cliCtx.String(FlagFile)
			if filePath == "" {
				filePath = DefaultHTTPYamlFile
			}

			httpTemplate, err := a.parseHTTPYamlFile(cliCtx.Context, filePath)
			if err != nil {
				return err
			}

			httpTemplate.Sanitize()

			err = httpTemplate.Validate()
			if err != nil {
				return err
			}

			variablesFilePaths := cliCtx.StringSlice(FlagVariableFile)

			fileVariables, err := a.parseVariablesFromFiles(cliCtx.Context, variablesFilePaths)
			if err != nil {
				return err
			}

			a.app = app.New(*httpTemplate, fileVariables)

			return nil
		},
		Action: func(cliCtx *cli.Context) error {
			if cliCtx.Args().Len() == 0 {
				fmt.Println("Use `yurl -h` for help")
				return nil
			}

			cliVariables := variable.NewVariables()

			// Parse variables from command line
			for _, v := range cliCtx.StringSlice(FlagVariable) {
				parsedVariable, err := variable.ParseStringWithSource(v, variable.SourceCLI)
				if err != nil && !errors.As(err, &variable.ErrInvalidFormat{}) {
					return err
				}
				cliVariables.Add(parsedVariable)
			}

			requestName := cliCtx.Args().First()

			return a.app.ExecuteRequest(cliCtx.Context, requestName, app.ExecuteRequestOpts{
				ListVariables: cliCtx.Bool(FlagListVariables),
				Verbose:       cliCtx.Bool(FlagVerbose),
				Variables:     cliVariables,
			})
		},
	}
}

func (a *CliApp) parseHTTPYamlFile(_ context.Context, filePath string) (*models.HttpTemplate, error) {
	file, err := os.Open(filePath)
	if err != nil {
		if err == os.ErrNotExist {
			return nil, fmt.Errorf("no request file found at %s", filePath)
		}
		return nil, err
	}

	var template models.HttpTemplate

	err = yaml.NewDecoder(file).Decode(&template)
	if err != nil {
		return nil, err
	}

	// Set name for each request
	for name, req := range template.Requests {
		req.Name = name
		template.Requests[name] = req
	}

	return &template, nil
}

func (a *CliApp) parseVariablesFromFiles(_ context.Context, filePaths []string) (variable.Variables, error) {
	variables := variable.NewVariables()

	for _, filePath := range filePaths {
		file, err := os.Open(filePath)
		if err != nil {
			if err == os.ErrNotExist {
				return nil, fmt.Errorf("no request file found at %s", filePath)
			}
			return nil, err
		}

		reader := bufio.NewReader(file)

		for {
			lineBytes, _, err := reader.ReadLine()
			if err == io.EOF {
				break
			} else if err != nil {
				return nil, err
			}

			line := string(lineBytes)

			parsedVariable, err := variable.ParseStringWithSource(line, variable.SourceInput)
			if err != nil && !errors.As(err, &variable.ErrInvalidFormat{}) {
				variables.Add(parsedVariable)
			}
		}
	}

	return variables, nil
}
