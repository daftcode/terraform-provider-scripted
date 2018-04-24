package shell

import (
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceGenericShellCRUDE() *schema.Resource {
	return &schema.Resource{
		Create: resourceGenericShellCreate,
		Read:   resourceShellRead,
		Update: resourceGenericShellUpdate,
		Delete: resourceGenericShellDelete,
		Exists: resourceGenericShellExists,

		Schema: map[string]*schema.Schema{
			"context": {
				Type:        schema.TypeMap,
				Optional:    true,
				Description: "Template context for rendering commands",
			},
			"output": {
				Type:        schema.TypeMap,
				Computed:    true,
				Description: "Output from the read command",
			},
		},
	}
}
