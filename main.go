// Package main contain application entry point
package main

import (
	"fmt"
	"os"

	"github.com/nikitamarchenko/golang-template-api-server/cmd"
)

func main() {
	err := cmd.Execute()
	if err != nil {
		_, _ = fmt.Printf("error: %v\n", err) //nolint:forbidigo // we don't have logger here

		os.Exit(cmd.ErrorExitCodeCommon)
	}
}
