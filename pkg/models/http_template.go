package models

import (
	"fmt"
	"net/http"
)

type Variables map[string]any

type HttpTemplate struct {
	Config   Config                         `yaml:"config"`
	Requests map[string]HttpTemplateRequest `yaml:"requests"`
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

type HttpTemplateRequest struct {
	Name        string
	Method      string            `yaml:"method"`
	Path        string            `yaml:"path"`
	JsonBody    string            `yaml:"jsonBody"`
	Headers     map[string]string `yaml:"headers"`
	PreRequests []PreRequest      `yaml:"pre"`
	Exports     map[string]Export `yaml:"exports"`
}

type HttpRequest struct {
	Template   *HttpTemplateRequest
	RawRequest *http.Request
}

type HttpResponse struct {
	RawResponse *http.Response
	RawBody     []byte
	Exports     map[string]any
}

func (r *HttpTemplateRequest) Sanitize() {
	if r.Method == "" {
		r.Method = "GET"
	}

	if r.Path == "" {
		r.Path = "/"
	}
}

func (r *HttpTemplateRequest) Validate(httpTemplate *HttpTemplate) error {
	// Validate existence of pre requests
	for _, preRequest := range r.PreRequests {
		if _, ok := httpTemplate.Requests[preRequest.Name]; !ok {
			return fmt.Errorf("%s requires a pre request %s which is not defined", r.Name, preRequest.Name)
		}
	}

	return nil
}
