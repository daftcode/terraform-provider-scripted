package scripted

import (
	"github.com/hashicorp/terraform/helper/schema"
	"strings"
)

type ResourceData struct {
	*schema.ResourceData
}

func (d *ResourceData) GetChange(key string) (interface{}, interface{}) {
	o, n := d.ResourceData.GetChange(key)
	return deterraformify(o), deterraformify(n)
}
func (d *ResourceData) Get(key string) interface{} {
	return deterraformify(d.ResourceData.Get(key))
}
func (d *ResourceData) GetOk(key string) (interface{}, bool) {
	value, ok := d.ResourceData.GetOk(key)
	return deterraformify(value), ok
}
func (d *ResourceData) Set(key string, value interface{}) error {
	return d.ResourceData.Set(key, terraformify(value))
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
