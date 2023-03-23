package models

import (
	"fmt"
	"net/http"
)

type HttpTemplate struct {
	Config   Config                 `yaml:"config"`
	Requests map[string]HttpRequest `yaml:"requests"`
}

type Config struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

type PreRequest struct {
	Name string `yaml:"name"`
}

type HttpRequest struct {
	Name        string
	Method      string            `yaml:"method"`
	Path        string            `yaml:"path"`
	JsonBody    string            `yaml:"jsonBody"`
	Headers     map[string]string `yaml:"headers"`
	PreRequests []PreRequest      `yaml:"pre"`
}

type HttpResponse struct {
	RawResponse *http.Response
	RawBody     []byte
}

func (r *HttpRequest) Sanitize() {
	if r.Method == "" {
		r.Method = "GET"
	}

	if r.Path == "" {
		r.Path = "/"
	}
}

func (r *HttpRequest) Validate(httpTemplate *HttpTemplate) error {
	// Validate existence of pre requests
	for _, preRequest := range r.PreRequests {
		if _, ok := httpTemplate.Requests[preRequest.Name]; !ok {
			return fmt.Errorf("%s requires a pre request %s which is not defined", r.Name, preRequest.Name)
		}
	}

	return nil
}
