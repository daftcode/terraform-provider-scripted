package scripted

import "github.com/hashicorp/terraform/helper/schema"

func getScriptedDataSource() *schema.Resource {
	resource := getScriptedResource()
	read := resource.Read
	resource.Read = func(d *schema.ResourceData, meta interface{}) error {
		err := read(d, meta)
		d.SetId("-")
		return err
	}
	resource.Create = nil
	resource.Update = nil
	resource.Delete = nil
	resource.Exists = nil
	return resource
}
