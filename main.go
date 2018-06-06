package main

import (
	"github.com/daftcode/terraform-provider-scripted/scripted"
	"github.com/hashicorp/terraform/plugin"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: scripted.Provider,
	})
}
