# Two runs apply

Example of terraform config running in 2 stages (running commands twice).
On the first run creates "scripted_resource.dependency" and on second run creates "scripted_resource.test" without
throwing errors.

```hcl-terraform
provider "scripted" {
	commands_dependencies = <<EOF
{{- if .Cur.dependency_path -}}
[ -f {{ .Cur.dependency_path | quote }} ] && echo -n true || echo -n false
{{- else -}}
echo -n true
{{- end -}}
EOF
	commands_create = "echo -n {{ .Cur.content | quote }} > {{ .Cur.path }}"
	commands_read = "echo \"out=$(cat {{ .Cur.path | quote }})\""
	commands_delete = "rm {{ .Cur.path | quote }}"
}
resource "scripted_resource" "dependency" {
	log_name = "dependency"
	context {
		path = "dependency"
		content = "dependency"
	}
}
resource "scripted_resource" "test" {
	log_name = "test"
	context {
		path = "test_file"
		content = "hi"
		dependency_path = "dependency"
	}
}
```