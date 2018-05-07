package main

import (
	"github.com/hashicorp/terraform/plugin"
	"github.com/nazarewk/terraform-provider-custom/custom"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: custom.Provider,
	})
}
