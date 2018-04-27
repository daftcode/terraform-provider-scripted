package shell

import (
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceGenericShellCRD() *schema.Resource {
	return getResource(false, false)
}
