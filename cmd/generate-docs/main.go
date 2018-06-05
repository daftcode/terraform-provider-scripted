package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/daftcode/terraform-provider-scripted/scripted"
	"io/ioutil"
	"os"
	"path"
	"text/template"
)

func render(tpl string, context interface{}, outPath string) string {
	t := template.New(outPath)
	t = t.Funcs(scripted.TemplateFuncs)
	t, err := t.Parse(tpl)
	exitIf(err)
	var buf bytes.Buffer
	err = t.Execute(&buf, context)
	exitIf(err)
	return buf.String()
}

func writeTemplate(tpl string, context interface{}, outPath string) {
	write(outPath, render(tpl, context, outPath))
}

func write(path, data string) {
	exitIf(ioutil.WriteFile(path, []byte(data), 0644))
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

const readme = `# {{ .name }} {{ .version }}
- [provider.md](provider.md)
{{- range $name, $data := .resources }}
- [{{ $name }}.md]({{ $name }}.md)
{{- end }}
{{- range $name, $data := (index . "data-sources") }}
- [{{ $name }}.md]({{ $name }}.md)
{{- end }}
`

const commands = `## Argument reference
{{- `

func main() {
	context := getJson(arg(1))
	docs := argOrDefault(2, "docs")
	writeTemplate(readme, context, path.Join(docs, "README.md"))
}
