package models

type HttpTemplate struct {
	Config   Config                 `yaml:"config"`
	Requests map[string]HttpRequest `yaml:"requests"`
}

type Config struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

type HttpRequest struct {
	Name     string
	Method   string            `yaml:"method"`
	Path     string            `yaml:"path"`
	JsonBody string            `yaml:"jsonBody"`
	Headers  map[string]string `yaml:"headers"`
}

func (r *HttpRequest) Sanitize() {
	if r.Method == "" {
		r.Method = "GET"
	}

	if r.Path == "" {
		r.Path = "/"
	}
}
