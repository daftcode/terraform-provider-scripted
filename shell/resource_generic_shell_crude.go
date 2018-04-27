package shell

import (
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceGenericShellCRUDE() *schema.Resource {
	return getResource(true, true)
}
