package main

import (
	"fmt"
	"os"

	"github.com/gurleensethi/yurl/internal/cli"
)

func main() {
	err := cli.NewApp().Build().Run(os.Args)
	if err != nil {
		if os.Getenv("DEBUG") == "true" {
			panic(err)
		}

		fmt.Println()
		fmt.Println(err.Error())
		os.Exit(1)
	}
}
