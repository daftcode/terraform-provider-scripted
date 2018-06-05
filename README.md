# `terraform-provider-scripted`

This is a terraform provider that lets you wrap shell/interpreter based tools to [Terraform](https://terraform.io/) resources in a simple way.

## Installing

[Copied from the Terraform documentation](https://www.terraform.io/docs/plugins/basics.html):
> To install a plugin, put the binary somewhere on your filesystem, then configure Terraform to be able to find it. The configuration where plugins are defined is ~/.terraformrc for Unix-like systems and %APPDATA%/terraform.rc for Windows.

Build it from source (instructions below) and move the binary `terraform-provider-scripted` to `bin/` and it should work.

## Using the provider

First, an simple example that is used in tests too

```hcl
provider "scripted" {
  commands_create = "echo \"hi\" > test_file"
  commands_read = "echo -n \"out=$(cat test_file)\""
  commands_delete = "rm test_file"
}

resource "scripted_resource" "test" {
}
```

```console
$ terraform plan
$ terraform apply
$ terraform destroy
```

To create a more complete example add this to the sample example file

```hcl
provider "scripted" {
  alias = "file"
  commands_should_update = <<EOF
[ "$(cat '{{ .Cur.path }}')" == '{{ .Cur.content }}' ] || exit 1
EOF
  commands_create = "echo -n '{{ .Cur.content }}' > '{{ .Cur.path }}'"
  commands_read = "echo -n \"out=$(cat '{{ .Cur.path }}')\""
  commands_delete = "rm '{{ .Cur.path }}'"
}

resource "scripted_resource" "test" {
  provider = "scripted.file"
  context {
    path = "test_file"
    content = "hi"
  }
}
```

Parameters can be used to change the resources. More info in [docs](docs/README.md)

## Building from source

1.  [Install Go](https://golang.org/doc/install) on your machine
1.  `go get github.com/daftcode/terraform-provider-scripted`
1.  `make build`. You will now find the binary at `dist/terraform-provider-scripted`.

## Running acceptance tests

```console
make test
```

## Known Problems

* The provider will error instead of removing the resource if the delete command fails. However, this is a safe default.
* Changes in provider do not issue resource rebuilds. Please parametrize all parameters that will change.

## Authors

* Krzysztof Nazarewski

Based on [`terraform-provider-shell`](https://github.com/toddnni/terraform-provider-shell) by Toni Ylenius.


## License

* MIT, See LICENSE file
