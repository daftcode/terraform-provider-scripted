package shell

import (
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceGenericShellCRDE() *schema.Resource {
	return getResource(false, true)
}
