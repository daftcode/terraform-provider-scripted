package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/daftcode/terraform-provider-scripted/scripted"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"text/template"
)

type Template struct {
	tpl *template.Template
	ctx interface{}
}

type Templates struct {
	templates map[string]*Template
}

func arg(i int) string {
	if len(os.Args) <= i {
		exit(fmt.Sprintf("Argument %d is missing", i))
	}
	return os.Args[i]
}

func argOrDefault(i int, def string) string {
	if len(os.Args) > i {
		return os.Args[i]
	}
	return def
}
func getJson(path string) interface{} {
	content, err := ioutil.ReadFile(path)
	exitIf(err)
	var data interface{}
	exitIf(json.Unmarshal(content, &data))
	return data
}

func exit(msg string) {
	fmt.Println(msg)
	os.Exit(1)
}

func exitIf(err error) {
	if err != nil {
		exit(err.Error())
	}
}

func makeBackticks(s string) string {
	return strings.Replace(s, "\\'", "`", -1)
}

const readme = `# {{ .name }} {{ .version }}
- [provider_scripted.md](provider_scripted.md)
{{- range $name, $data := .resources }}
- [{{ $name }}.md]({{ $name }}.md)
{{- end }}
{{- range $name, $data := (index . "data-sources") }}
- [{{ $name }}.md]({{ $name }}.md)
{{- end }}
`

var description = makeBackticks(`## Argument reference

| Argument | Type | Description | Default |
|:---      | ---  | ---         | ---     |
{{- range $arg, $data := . }}
| \'{{ $arg }}\' | 
{{- "" }} [{{ $data.Type }}](https://www.terraform.io/docs/extend/schemas/schema-types.html#{{ $data.Type | lower }}) | 
{{- "" }} {{ $data.Description }} | 
{{- "" }} {{ if hasKey $data.Default "Value" }}{{ $default := default "" $data.Default.Value -}}
	{{ if contains "\'" $default | not }}\'{{ end }}
    {{- $default -}}
 	{{ if contains "\'" $default | not }}\'{{ end }}
{{- else }}not set{{ end }} | 
{{- end }}
`)

func (t *Templates) set(name, content string, context interface{}) {
	tpl := template.New(name)
	tpl = tpl.Funcs(scripted.TemplateFuncs)
	tpl, err := tpl.Parse(content)
	exitIf(err)
	t.templates[name] = &Template{
		tpl: tpl,
		ctx: context,
	}
}

func (t *Templates) write(baseDir string, context interface{}) {
	for name, tpl := range t.templates {
		tpl.write(path.Join(baseDir, name))
	}
}

func (t *Template) write(path string) {
	var buf bytes.Buffer
	exitIf(t.tpl.Execute(&buf, t.ctx))
	exitIf(ioutil.WriteFile(path, buf.Bytes(), 0644))
}

func get(data interface{}, path ... string) interface{} {
	cur := data.(map[string]interface{})
	var ok bool
	for _, key := range path {
		if cur, ok = cur[key].(map[string]interface{}); !ok {
			return nil
		}
	}
	return cur
}

func main() {
	t := &Templates{templates: map[string]*Template{}}
	data := getJson(arg(1))
	t.set("README.md", readme, data)
	t.set("provider_scripted.md", description, get(data, "provider"))
	for name, values := range get(data, "resources").(map[string]interface{}) {
		t.set(name+".md", description, values)
	}
	for name, values := range get(data, "data-sources").(map[string]interface{}) {
		t.set(name+".md", description, values)
	}
	t.write(argOrDefault(2, "docs"), data)
}
