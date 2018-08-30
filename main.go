package main

import (
	"fmt"
	"github.com/daftcode/terraform-provider-scripted/scripted"
	"github.com/hashicorp/terraform/plugin"
	"os"
)

func main() {
	args := os.Args[1:]
	versionArgs := []string{
		"version",
		"--version",
		"-v",
	}

	if len(args) == 1 {
		for _, cmd := range versionArgs {
			if args[0] == cmd {
				fmt.Print(scripted.Version)
				return
			}
		}
	}

	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: scripted.Provider,
	})
}
