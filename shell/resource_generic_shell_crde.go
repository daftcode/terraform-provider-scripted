package shell

import (
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceGenericShellCRDE() *schema.Resource {
	return &schema.Resource{
		Create: resourceGenericShellCreate,
		Read:   resourceShellRead,
		Delete: resourceGenericShellDelete,
		Exists: resourceGenericShellExists,

		// desc: will always recreate the resource if something is changed
		// will output variables but we don't define them here
		// eg. if contains access_ipv4

		Schema: map[string]*schema.Schema{
			"context": {
				Type:        schema.TypeMap,
				Optional:    true,
				Description: "Template context for rendering commands",
				ForceNew:    true,
			},
			"output": {
				Type:        schema.TypeMap,
				Computed:    true,
				Description: "Output from the read command",
			},
		},
	}
}
