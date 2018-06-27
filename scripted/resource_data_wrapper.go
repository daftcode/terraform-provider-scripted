package scripted

import "github.com/hashicorp/terraform/helper/schema"

type ResourceData struct {
	*schema.ResourceData
}

func (d *ResourceData) SetIdErr(value string) error {
	d.SetId(value)
	return nil
}

func WrapResourceData(data *schema.ResourceData) ResourceInterface {
	return &ResourceData{data}
}
