package main

import (
	"github.com/hashicorp/terraform/plugin"
	"github.com/daftcode/terraform-provider-scripted/scripted"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: scripted.Provider,
	})
}
