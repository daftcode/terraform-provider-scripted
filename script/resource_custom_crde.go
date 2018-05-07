package script

import (
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceScriptCRDE() *schema.Resource {
	return getResource(false, true)
}
