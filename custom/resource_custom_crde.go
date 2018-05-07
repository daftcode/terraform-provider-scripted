package custom

import (
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceCustomCRDE() *schema.Resource {
	return getResource(false, true)
}
