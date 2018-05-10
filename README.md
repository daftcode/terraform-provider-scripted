# Terraform custom provider

This is a terraform provider that lets you wrap shell/interpreter based tools to [Terraform](https://terraform.io/) resources in a simple way.

## Naming

The naming of this provider has been hard. The provider is about wrapping functionality by running shell or any other interpreter-based scripts. Originally the name was `generic_shell_wrapper` then `shell`, but currently the name is just `script`.

## Installing

[Copied from the Terraform documentation](https://www.terraform.io/docs/plugins/basics.html):
> To install a plugin, put the binary somewhere on your filesystem, then configure Terraform to be able to find it. The configuration where plugins are defined is ~/.terraformrc for Unix-like systems and %APPDATA%/terraform.rc for Windows.

Build it from source (instructions below) and move the binary `terraform-provider-scripted` to `bin/` and it should work.

## Using the provider

First, an simple example that is used in tests too

```hcl
provider "scripted" {
  create_command = "echo \"hi\" > test_file"
  read_command = "echo -n \"out=$(cat test_file)\""
  delete_command = "rm test_file"
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
  alias = "write_to_file"
  create_command = "echo \"{{.new.input}}\" > {{.new.file}}"
  read_command = "echo -n \"out=$(cat '{{.new.file}}')\""
  delete_command = "rm {{.old.file}}"
}

resource "scripted_resource" "filetest" {
  provider = "script.write_to_file"
  context {
    input = "this to the file"
    file = "test_file2"
  }
}
```

Parameters can by used to change the resources.

## Building from source

1.  [Install Go](https://golang.org/doc/install) on your machine
2.  [Set up Gopath](https://golang.org/doc/code.html)
3.  `git clone` this repository into `$GOPATH/src/github.com/nazarewk/terraform-provider-scripted`
4.  Get the dependencies. Run `go get`
6.  `make install`. You will now find the
    binary at `$GOPATH/bin/terraform-provider-scripted`.

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
