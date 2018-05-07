package custom

import (
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceCustomCRUDE() *schema.Resource {
	return getResource(true, true)
}
