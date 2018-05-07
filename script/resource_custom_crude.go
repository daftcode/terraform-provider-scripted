package script

import (
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceScriptCRUDE() *schema.Resource {
	return getResource(true, true)
}
