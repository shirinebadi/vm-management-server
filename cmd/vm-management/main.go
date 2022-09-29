package main

import (
	"os"

	"github.com/shirinebadi/vm-management-server/internal/app/vm-management/cmd"
)

const (
	exitFailure = 1
)

func main() {
	root := cmd.NewRootCommand()

	if root != nil {
		if err := root.Execute(); err != nil {
			os.Exit(exitFailure)
		}
	}
}
