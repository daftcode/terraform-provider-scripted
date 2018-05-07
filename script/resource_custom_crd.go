package script

import (
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceScriptCRD() *schema.Resource {
	return getResource(false, false)
}
