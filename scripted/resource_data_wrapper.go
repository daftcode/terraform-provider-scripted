package scripted

import (
	"github.com/hashicorp/terraform/helper/schema"
	"strings"
)

type ResourceData struct {
	*schema.ResourceData
}

func (d *ResourceData) SetIdErr(value string) error {
	d.SetId(value)
	return nil
}
func (d *ResourceData) GetChangedKeysPrefix(prefix string) []string {
	var ret []string
	for key := range resourceSchema {
		if strings.HasPrefix(key, prefix) && d.ResourceData.HasChange(key) {
			ret = append(ret, key)
		}
	}
	return ret
}

func WrapResourceData(data *schema.ResourceData) ResourceInterface {
	return &ResourceData{data}
}
