package scripted

import (
	"encoding/json"
	"fmt"
	"github.com/Masterminds/sprig"
	"github.com/ghodss/yaml"
	"reflect"
	"text/template"
)

var TemplateFuncs = getTemplateFuncs()
var SprigTemplateFuncs = getSprigTemplateFuncs()
var ExtraTemplateFuncs = template.FuncMap{
	"toYaml":              toYaml,
	"fromYaml":            fromYaml,
	"toJson":              toJson,
	"toPrettyJson":        toPrettyJson,
	"fromJson":            fromJson,
	"is":                  is,
	"isSet":               isSet,
	"isFilled":            isFilled,
	"stringifyJsonValues": stringifyJsonValues,
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

func fromJsonMust(value string) interface{} {
	ret, _ := fromJson(value)
	return ret
}

func stringifyJsonValues(value interface{}) interface{} {
	var inner func(interface{}) interface{}
	inner = func(cur interface{}) interface{} {
		curValue := reflect.ValueOf(cur)
		switch curValue.Kind() {
		case reflect.Map:
			mapped := map[string]interface{}{}
			for _, key := range curValue.MapKeys() {
				mapped[key.String()] = inner(curValue.MapIndex(key).Interface())
			}
			return mapped

		case reflect.Slice, reflect.Array:
			var list []interface{}
			for i := 0; i < curValue.Len(); i++ {
				list = append(list, inner(curValue.Index(i).Interface()))
			}
			return list

		default:
			return fmt.Sprintf("%v", cur)
		}
	}
	return inner(value)
}
