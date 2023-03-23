package main

import (
	"fmt"
	"os"

	"github.com/gurleensethi/yurl/pkg/app"
)

func main() {
	err := app.New().BuildCliApp().Run(os.Args)
	if err != nil {
		if os.Getenv("DEBUG") == "true" {
			panic(err)
		}

		fmt.Println(err.Error())
		os.Exit(1)
	}
}
