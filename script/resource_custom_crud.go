package script

import (
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceScriptCRUD() *schema.Resource {
	return getResource(true, false)
}
