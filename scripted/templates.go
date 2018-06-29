package scripted

import (
	"encoding/json"
	"github.com/Masterminds/sprig"
	"github.com/ghodss/yaml"
	"text/template"
)

var TemplateFuncs = getTemplateFuncs()
var SprigTemplateFuncs = getSprigTemplateFuncs()
var ExtraTemplateFuncs = template.FuncMap{
	"toYaml":       toYaml,
	"fromYaml":     fromYaml,
	"toJson":       toJson,
	"toPrettyJson": toPrettyJson,
	"fromJson":     fromJson,
	"is":           is,
	"isSet":        isSet,
	"isFilled":     isFilled,
}

func getSprigTemplateFuncs() template.FuncMap {
	ret := sprig.TxtFuncMap()
	delete(ret, "env")
	delete(ret, "expandenv")
	return ret
}

func getTemplateFuncs() template.FuncMap {
	ret := template.FuncMap{}
	for k, v := range SprigTemplateFuncs {
		ret[k] = v
	}
	for k, v := range ExtraTemplateFuncs {
		ret[k] = v
	}
	return ret
}

func toYaml(value interface{}) (string, error) {
	ret, err := yaml.Marshal(value)
	return string(ret[:]), err
}

func fromYaml(value string) (interface{}, error) {
	var ret interface{}
	err := yaml.Unmarshal([]byte(value), &ret)
	return ret, err
}

func toJson(value interface{}) (string, error) {
	ret, err := json.Marshal(value)
	return string(ret[:]), err
}

func toJsonMust(value interface{}) string {
	ret, _ := toJson(value)
	return ret
}

func toPrettyJson(value interface{}) (string, error) {
	ret, err := json.MarshalIndent(value, "", "  ")
	return string(ret[:]), err
}
func toPrettyJsonMust(value interface{}) string {
	ret, _ := toPrettyJson(value)
	return ret
}

func fromJson(value string) (interface{}, error) {
	var ret interface{}
	err := json.Unmarshal([]byte(value), &ret)
	return ret, err
}
