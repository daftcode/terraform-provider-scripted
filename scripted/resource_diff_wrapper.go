package scripted

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"runtime/debug"
)

type ResourceDiff struct {
	*schema.ResourceDiff
}

func (rd *ResourceDiff) GetRollbackKeys() []string {
	return rd.GetChangedKeysPrefix("")
}

func (rd *ResourceDiff) IsNew() bool {
	return rd.ResourceDiff.Id() == ""
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

func (rd *ResourceDiff) GetOld(key string) interface{} {
	o, _ := rd.GetChange(key)
	return o
}

func (rd *ResourceDiff) Set(key string, value interface{}) error {
	err := rd.ResourceDiff.SetNew(key, value)
	if err != nil {
		debug.PrintStack()
	}
	return err
}

func (rd *ResourceDiff) SetIdErr(string) error {
	return fmt.Errorf("not implemented")
}

func (rd *ResourceDiff) GetChangedKeysPrefix(prefix string) []string {
	return rd.ResourceDiff.GetChangedKeysPrefix(prefix)
}

func (rd *ResourceDiff) HasChangedKeysPrefix(prefix string) bool {
	return len(rd.GetChangedKeysPrefix(prefix)) > 0
}

func (rd *ResourceDiff) HasChange(key string) bool {
	return !rd.ResourceDiff.NewValueKnown(key) || rd.ResourceDiff.HasChange(key)
}

func WrapResourceDiff(diff *schema.ResourceDiff) ResourceInterface {
	return &ResourceDiff{diff}
}
