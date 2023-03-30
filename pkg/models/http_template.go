package models

import (
	"fmt"
	"net/http"
	"strings"
)

type VariableSource string

const (
	VariableSourceVarFile     VariableSource = "variable file"
	VariableSourceRequestFile VariableSource = "request file"
	VariableSourceCLI         VariableSource = "cli"
	VariableSourceInput       VariableSource = "input"
	VariableSourceExports     VariableSource = "request exports"
)

type Variable struct {
	Value  any
	Source VariableSource
}

type Variables map[string]Variable

type HttpTemplate struct {
	Config   Config                         `yaml:"config"`
	Requests map[string]HttpRequestTemplate `yaml:"requests"`
}

func (t *HttpTemplate) Sanitize() {
	t.Config.Sanitize()

	for _, request := range t.Requests {
		request.Sanitize()
	}
}

func (t *HttpTemplate) Validate() error {
	err := t.Config.Validate()
	if err != nil {
		return err
	}

	for _, request := range t.Requests {
		err := request.Validate(t)
		if err != nil {
			return err
		}
	}

	return nil
}

type Config struct {
	Host   string `yaml:"host"`
	Port   int    `yaml:"port"`
	Scheme string `yaml:"scheme"`
}

func (c Config) Validate() error {
	if c.Host == "" {
		return fmt.Errorf("config.host is required")
	}

	if c.Port == 0 {
		return fmt.Errorf("config.port is required")
	}

	if c.Scheme != "http" && c.Scheme != "https" {
		return fmt.Errorf("config.scheme must be http or https")
	}

	return nil
}

func (c *Config) Sanitize() {
	if c.Scheme == "" {
		c.Scheme = "http"
	}
}

type PreRequest struct {
	Name string `yaml:"name"`
}

type Export struct {
	JSON string `yaml:"json"`
}

type HttpRequestTemplate struct {
	Name        string
	Description string            `yaml:"description"`
	Method      string            `yaml:"method"`
	Path        string            `yaml:"path"`
	Body        string            `yaml:"body"`
	JsonBody    string            `yaml:"jsonBody"`
	Headers     map[string]string `yaml:"headers"`
	Query       map[string]string `yaml:"query"`
	PreRequests []PreRequest      `yaml:"pre"`
	Exports     map[string]Export `yaml:"exports"`
}

type HttpRequest struct {
	Template   *HttpRequestTemplate
	RawRequest *http.Request
	Variables  Variables
}

type HttpResponse struct {
	Request     *HttpRequest
	RawResponse *http.Response
	RawBody     []byte
	Exports     map[string]any
}

func (r *HttpRequestTemplate) Sanitize() {
	r.Method = strings.ToUpper(r.Method)
	if r.Method == "" {
		r.Method = "GET"
	}

	if r.Path == "" {
		r.Path = "/"
	}
}

func (r *HttpRequestTemplate) Validate(httpTemplate *HttpTemplate) error {
	// Validate existence of pre requests
	for _, preRequest := range r.PreRequests {
		if _, ok := httpTemplate.Requests[preRequest.Name]; !ok {
			return fmt.Errorf("'%s' requires a pre request '%s' which is not defined", r.Name, preRequest.Name)
		}
	}

	return nil
}
