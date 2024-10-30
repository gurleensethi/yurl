package models

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gurleensethi/yurl/internal/variable"
)

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

	// DFS to check any cycles on pre requests
	for _, request := range t.Requests {
		if ok, requestChain := t.findPreRequestCycles(request.Name, make([]string, 0), make(map[string]struct{})); ok {
			return fmt.Errorf("cycle detected in pre-requests: %s", strings.Join(requestChain, " -> "))
		}
	}

	return nil
}

// findPreRequestCycles checks if there are any cycles in pre request chain.
func (t *HttpTemplate) findPreRequestCycles(requestName string, requestChain []string, visited map[string]struct{}) (bool, []string) {
	request := t.Requests[requestName]
	visited[requestName] = struct{}{}
	requestChain = append(requestChain, requestName)

	for _, preRequest := range request.PreRequests {
		if _, ok := visited[preRequest.Name]; ok {
			return true, append(requestChain, preRequest.Name)
		}

		isCycleDetected, requestChain := t.findPreRequestCycles(preRequest.Name, requestChain, visited)
		if isCycleDetected {
			return true, requestChain
		}
	}

	return false, nil
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
	Variables  variable.Variables
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
