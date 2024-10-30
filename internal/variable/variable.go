package variable

import (
	"strings"
)

type Source string

const (
	SourceUnknown     Source = "unknown"
	SourceVarFile     Source = "variable file"
	SourceRequestFile Source = "request file"
	SourceCLI         Source = "cli"
	SourceInput       Source = "input"
	SourceExports     Source = "request exports"
)

type Variable struct {
	Key    string
	Value  any
	Source Source
}

type Variables map[string]Variable

func (vars Variables) Add(v Variable) {
	vars[v.Key] = v
}

func (vars Variables) Remove(key string) {
	delete(vars, key)
}

func (vars Variables) Get(key string) (Variable, bool) {
	v, ok := vars[key]
	return v, ok
}

type ErrInvalidFormat struct {
	format string
}

func (e ErrInvalidFormat) Error() string {
	return "invalid variable format: " + e.format
}

func ParseStringWithSource(v string, source Source) (Variable, error) {
	parts := strings.Split(v, "=")

	if len(parts) != 2 {
		return Variable{}, ErrInvalidFormat{format: v}
	}

	key := parts[0]
	value := strings.Join(parts[1:], "=")
	return Variable{
		Key:    key,
		Value:  value,
		Source: source,
	}, nil
}

func ParseString(v string) (Variable, error) {
	return ParseStringWithSource(v, SourceUnknown)
}
