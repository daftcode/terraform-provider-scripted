package custom

import (
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceCustomCRUD() *schema.Resource {
	return getResource(true, false)
}
