package scripted

import (
	"encoding/json"
	"github.com/Masterminds/sprig"
	"github.com/ghodss/yaml"
	"text/template"
)

var TemplateFuncs = getTemplateFuncs()
var templateFuncs = template.FuncMap{
	"toYaml": func(value interface{}) (string, error) {
		ret, err := yaml.Marshal(value)
		return string(ret[:]), err
	},
	"fromYaml": func(value string) (interface{}, error) {
		var ret interface{}
		err := yaml.Unmarshal([]byte(value), &ret)
		return ret, err
	},
	"toJson": func(value interface{}) (string, error) {
		ret, err := json.Marshal(value)
		return string(ret[:]), err
	},
	"toPrettyJson": func(value interface{}) (string, error) {
		ret, err := json.MarshalIndent(value, "", "  ")
		return string(ret[:]), err
	},
	"fromJson": func(value string) (interface{}, error) {
		var ret interface{}
		err := json.Unmarshal([]byte(value), &ret)
		return ret, err
	},
}

func getTemplateFuncs() template.FuncMap {
	ret := sprig.TxtFuncMap()
	delete(ret, "env")
	delete(ret, "expandenv")

	for k, v := range templateFuncs {
		ret[k] = v
	}
	return ret
}
