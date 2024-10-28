package cli

import (
	_ "embed"
	"fmt"
	"os"
	"path"

	"github.com/urfave/cli/v2"
)

//go:embed http.yaml.template
var httpYamlTemplateFile []byte

// InitConfigFile creates a new config file at the specified output directory.
func InitConfigFile(c *cli.Context) error {
	fileDestinationPath := c.String("path")

	// Make sure destination path exists
	if fileDestinationPath != "" {
		dir, err := os.Stat(fileDestinationPath)
		if os.IsNotExist(err) {
			return fmt.Errorf("destination path \"%s\" is invalid", fileDestinationPath)
		}
		if err != nil {
			return err
		}

		if !dir.IsDir() {
			return fmt.Errorf("destination path \"%s\" is not a directory", fileDestinationPath)
		}
	} else {
		fileDestinationPath = "."
	}

	// Make sure config file already doesn't exists
	configFilePath := path.Join(fileDestinationPath, DefaultHTTPYamlFile)

	_, err := os.Stat(configFilePath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	if err == nil {
		return fmt.Errorf("http.yaml file already exists at path \"%s\"", fileDestinationPath)
	}

	file, err := os.Create(configFilePath)
	if err != nil {
		return fmt.Errorf("failed to create http.yaml file: %w", err)
	}

	_, err = file.Write(httpYamlTemplateFile)
	if err != nil {
		return fmt.Errorf("failed to write to http.yaml file: %w", err)
	}

	err = file.Close()
	if err != nil {
		return err
	}

	fmt.Printf("http template file initalized: \"%s\"\n", configFilePath)

	return nil
}
