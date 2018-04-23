package shell

import (
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceGenericShellCRUD() *schema.Resource {
	return &schema.Resource{
		Create: resourceGenericShellCreate,
		Read:   resourceShellRead,
		Update: resourceGenericShellUpdate,
		Delete: resourceGenericShellDelete,

		// desc: will always recreate the resource if something is changed
		// will output variables but we don't define them here
		// eg. if contains access_ipv4

		Schema: map[string]*schema.Schema{
			"data": {
				Type:        schema.TypeMap,
				Optional:    true,
				Description: "The input data for commands",
			},
			"output": {
				Type:        schema.TypeMap,
				Computed:    true,
				Description: "Output from the read command",
			},
		},
	}
}