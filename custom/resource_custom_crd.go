package custom

import (
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceCustomCRD() *schema.Resource {
	return getResource(false, false)
}
