package scripted

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
)

type ResourceDiff struct {
	*schema.ResourceDiff
}

func (rd *ResourceDiff) GetChange(key string) (o interface{}, n interface{}) {
	defer func() {
		if r := recover(); r != nil {
			o = nil
			defer func() {
				if r := recover(); r != nil {
					n = nil
				}
			}()
			n = rd.Get(key)
			return
		}
	}()
	o, n = rd.ResourceDiff.GetChange(key)
	return o, n
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
