package scripted

import "github.com/hashicorp/terraform/helper/schema"

func getScriptedResource() *schema.Resource {
	ret := &schema.Resource{
		Create: resourceScriptedCreate,
		Read:   resourceScriptedRead,
		Update: resourceScriptedUpdate,
		Delete: resourceScriptedDelete,
		Exists: resourceScriptedExists,

		Schema: map[string]*schema.Schema{
			"log_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "Resource name to display in log messages",
				// Hack so it doesn't ever change
				ForceNew: true,
				StateFunc: func(v interface{}) string {
					return ""
				},
			},
			"context": {
				Type:        schema.TypeMap,
				Optional:    true,
				Description: "Template context for rendering commands",
			},
			"environment": {
				Type:        schema.TypeMap,
				Optional:    true,
				Default:     []string{},
				Description: "Environment to run commands in",
			},
			"output": {
				Type:        schema.TypeMap,
				Computed:    true,
				Description: "Output from the read command",
			},
		},
	}
	return ret
}
