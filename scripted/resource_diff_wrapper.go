package scripted

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
)

type ResourceDiff struct {
	*schema.ResourceDiff
}

func (rd *ResourceDiff) Set(key string, value interface{}) error {
	return fmt.Errorf("not implemented")
}

func (rd *ResourceDiff) SetIdErr(string) error {
	return fmt.Errorf("not implemented")
}

func WrapResourceDiff(diff *schema.ResourceDiff) ResourceInterface {
	return &ResourceDiff{diff}
}
