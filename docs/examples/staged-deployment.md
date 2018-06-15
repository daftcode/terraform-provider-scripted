# Staged deployment (based on a variable)

## Command 1 - apply - skip creation

This apply does not create resource even though it is defined

```hcl-terraform
locals {
  stage1 = ""
}

resource "scripted_resource" "stage1" {
	context {
		path = "test_file"
	}
}

provider "scripted" {
	commands_needs_delete = "[ -n '${local.stage1}' ] || echo -n true"
	commands_create = "echo -n 'hi' > {{ .Cur.path }}"
	commands_read = "echo \"out=$(cat {{ .Cur.path | quote }})\""
	commands_delete = "rm {{ .Cur.path | quote }}"
}
```

## Command 2 - apply - create resource

This apply creates resource normally

```hcl-terraform
locals {
  stage1 = "1"
}

resource "scripted_resource" "stage1" {
	context {
		path = "test_file"
	}
}

provider "scripted" {
	commands_needs_delete = "[ -n '${local.stage1}' ] || echo -n true"
	commands_create = "echo -n 'hi' > {{ .Cur.path }}"
	commands_read = "echo \"out=$(cat {{ .Cur.path | quote }})\""
	commands_delete = "rm {{ .Cur.path | quote }}"
}
```

## Command 3 - plan - empty plan

This plan doesn't do anything as it should be

```hcl-terraform
locals {
  stage1 = "1"
}

resource "scripted_resource" "stage1" {
	context {
		path = "test_file"
	}
}

provider "scripted" {
	commands_needs_delete = "[ -n '${local.stage1}' ] || echo -n true"
	commands_create = "echo -n 'hi' > {{ .Cur.path }}"
	commands_read = "echo \"out=$(cat {{ .Cur.path | quote }})\""
	commands_delete = "rm {{ .Cur.path | quote }}"
}
```

# Command 4 - apply - undo stage1

This apply undoes stage1

```hcl-terraform
locals {
  stage1 = ""
}

resource "scripted_resource" "stage1" {
	context {
		path = "test_file"
	}
}

provider "scripted" {
	commands_needs_delete = "[ -n '${local.stage1}' ] || echo -n true"
	commands_create = "echo -n 'hi' > {{ .Cur.path }}"
	commands_read = "echo \"out=$(cat {{ .Cur.path | quote }})\""
	commands_delete = "rm {{ .Cur.path | quote }}"
}
```