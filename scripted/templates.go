package scripted

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/Masterminds/sprig"
	"github.com/ghodss/yaml"
	"text/template"
)

var TemplateFuncs = getTemplateFuncs()
var SprigTemplateFuncs = getSprigTemplateFuncs()
var ExtraTemplateFuncs = template.FuncMap{
	"toYaml":             toYaml,
	"fromYaml":           fromYaml,
	"toJson":             toJson,
	"toPrettyJson":       toPrettyJson,
	"fromJson":           fromJson,
	"is":                 is,
	"isSet":              isSet,
	"isFilled":           isFilled,
	"terraformifyValues": terraformifyPrimitives,

	"include":  func(string, interface{}) string { return "not implemented" },
	"required": func(string, interface{}) interface{} { return "not implemented" },
}

func NewTemplate(name string) *template.Template {
	t := template.New(name)
	t = t.Funcs(getFuncsForTemplate(t))
	return t
}

func getSprigTemplateFuncs() template.FuncMap {
	ret := sprig.TxtFuncMap()
	delete(ret, "env")
	delete(ret, "expandenv")
	return ret
}

func getFuncsForTemplate(t *template.Template) template.FuncMap {
	funcMap := make(template.FuncMap)
	for k, v := range TemplateFuncs {
		funcMap[k] = v
	}

	// Copied from https://github.com/helm/helm/blob/3f0c6c54049d38e5c86dad1e9475f7dbf783be21/pkg/engine/engine.go#L150-L178

	// Add the 'include' function here so we can close over t.
	funcMap["include"] = func(name string, data interface{}) (string, error) {
		buf := bytes.NewBuffer(nil)
		if err := t.ExecuteTemplate(buf, name, data); err != nil {
			return "", err
		}
		return buf.String(), nil
	}

	// Add the 'required' function here
	funcMap["required"] = func(warn string, val interface{}) (interface{}, error) {
		if val == nil {
			return val, fmt.Errorf(warn)
		} else if _, ok := val.(string); ok {
			if val == "" {
				return val, fmt.Errorf(warn)
			}
		}
		return val, nil
	}
	return funcMap
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
