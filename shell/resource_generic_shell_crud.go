package shell

import (
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceGenericShellCRUD() *schema.Resource {
	return getResource(true, false)
}
