package main

import (
	"os"

	"github.com/gurleensethi/yurl/pkg/app"
)

func main() {
	err := app.New().BuildCliApp().Run(os.Args)
	if err != nil {
		panic(err)
	}
}
