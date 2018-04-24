package shell

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os/exec"
	"strings"

	"bytes"
	"encoding/base64"
	"github.com/armon/circbuf"
	"github.com/hashicorp/terraform/helper/schema"
	"text/template"
)

func resourceGenericShellCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	command, err := interpolateCommand(
		mergeCommands(config, config.CommandPrefix, config.CreateCommand),
		getContext(d, "create"))
	if err != nil {
		return err
	}
	writeLog("DEBUG", "creating resource")
	_, _, err = runCommand(config, command)
	if err != nil {
		return err
	}

	d.SetId(hash(command))
	writeLog("DEBUG", "created generic resource: %s", d.Id())

	return resourceShellRead(d, meta)
}

func getOutputsText(output string, prefix string) map[string]string {
	outputs := make(map[string]string)
	split := strings.Split(output, "\n")
	for _, varline := range split {
		writeLog("DEBUG", "reading output line: %s", varline)

		if varline == "" {
			continue
		}

		if prefix != "" {
			if !strings.HasPrefix(varline, prefix) {
				writeLog("INFO", "ignoring line without prefix `%s`: \"%s\"", prefix, varline)
				continue
			}
			varline = strings.TrimPrefix(varline, prefix)
		}

		pos := strings.Index(varline, "=")
		if pos == -1 {
			writeLog("INFO", "ignoring line without equal sign: \"%s\"", varline)
			continue
		}

		key := varline[:pos]
		value := varline[pos+1:]
		writeLog("DEBUG", "read output entry (raw): \"%s\" = \"%s\"", key, value)
		outputs[key] = value
	}
	return outputs
}

func getOutputsBase64(output, prefix string) map[string]string {
	outputs := make(map[string]string)
	for key, value := range getOutputsText(output, prefix) {
		decoded, err := base64.StdEncoding.DecodeString(value)
		if err != nil {
			writeLog("WARN", "error decoding %s", err)
			continue
		}
		writeLog("DEBUG", "read output entry (decoded): \"%s\" = \"%s\" (%s)", key, decoded, value)
		outputs[key] = string(decoded[:])
	}
	return outputs
}

func resourceShellRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	command, err := interpolateCommand(
		mergeCommands(config, config.CommandPrefix, config.ReadCommand),
		getContext(d, "read"))
	if err != nil {
		return err
	}
	writeLog("DEBUG", "reading resource")
	output, stderr, err := runCommand(config, command)
	if err != nil {
		writeLog("INFO", "command returned error (%s), marking resource deleted: %s, stderr: %s", err, output, stderr)
		if config.ReadDeleteOnFailure {
			d.SetId("")
			return nil
		}
		return err
	}
	var outputs map[string]string

	switch config.ReadFormat {
	case "base64":
		outputs = getOutputsBase64(output, config.ReadLinePrefix)
	default:
		fallthrough
	case "raw":
		outputs = getOutputsText(output, config.ReadLinePrefix)
	}
	d.Set("output", outputs)

	return nil
}

func resourceGenericShellUpdate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx := getContext(d, "update")
	deleteCommand, _ := interpolateCommand(
		mergeCommands(config, config.CommandPrefix, config.DeleteCommand),
		mergeMaps(ctx, map[string]interface{}{"cur": ctx["old"]}))
	createCommand, _ := interpolateCommand(mergeCommands(config, config.CommandPrefix, config.CreateCommand), ctx)
	command, err := interpolateCommand(
		mergeCommands(config, config.CommandPrefix, config.UpdateCommand),
		mergeMaps(ctx, map[string]interface{}{
			"delete_command": deleteCommand,
			"create_command": createCommand,
		}))
	if err != nil {
		return err
	}
	writeLog("DEBUG", "updating resource: %s", command)
	stdout, stderr, err := runCommand(config, command)
	if err != nil {
		writeLog("WARN", "update command returned error: %s, stderr: %s", stdout, stderr)
		return nil
	}

	return resourceShellRead(d, meta)
}

func resourceGenericShellExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	config := meta.(*Config)

	command, err := interpolateCommand(
		mergeCommands(config, config.CommandPrefix, config.ExistsCommand),
		getContext(d, "exists"))
	if err != nil {
		return false, err
	}
	writeLog("DEBUG", "resource exists")
	stdout, stderr, err := runCommand(config, command)
	if err != nil {
		writeLog("WARN", "command returned error: %s, stderr: %s", stdout, stderr)
	}

	return stdout == "true", err
}

func resourceGenericShellDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	command, err := interpolateCommand(
		mergeCommands(config, config.CommandPrefix, config.DeleteCommand),
		getContext(d, "delete"))
	if err != nil {
		return err
	}
	writeLog("DEBUG", "reading resource")
	_, _, err = runCommand(config, command)
	if err != nil {
		return err
	}

	d.SetId("")
	return nil
}

func getContext(d *schema.ResourceData, operation string) map[string]interface{} {
	return getContextFull(d, operation, false)
}
func getContextFull(d *schema.ResourceData, operation string, oldIsCurrent bool) map[string]interface{} {
	o, n := d.GetChange("context")
	cur := n
	if oldIsCurrent {
		cur = o
	}
	return map[string]interface{}{
		"operation": operation,
		"old":       o,
		"new":       n,
		"cur":       cur,
	}
}

func mergeMaps(maps ...map[string]interface{}) map[string]interface{} {
	ctx := map[string]interface{}{}
	for _, m := range maps {
		for k, v := range m {
			ctx[k] = v
		}
	}
	return ctx
}

func interpolateCommand(command string, context map[string]interface{}) (string, error) {
	t := template.Must(template.New("command").Parse(command))
	var buf bytes.Buffer
	err := t.Execute(&buf, context)
	return buf.String(), err
}

func runCommand(config *Config, commands ...string) (string, string, error) {
	// Setup the command
	interpreter := config.Interpreter[0]
	command := mergeCommands(config, commands...)
	args := append(config.Interpreter[1:], command)
	cmd := exec.Command(interpreter, args...)
	stdout, _ := circbuf.NewBuffer(config.BufferSize)
	cmd.Stdout = io.Writer(stdout)
	stderr, _ := circbuf.NewBuffer(config.BufferSize)
	cmd.Stderr = io.Writer(stderr)

	// Output what we're about to run
	writeLog("DEBUG", "executing: %s %s", interpreter, strings.Join(args, " "))

	// Run the command to completion
	err := cmd.Run()

	if err != nil {
		return "", "", fmt.Errorf("error running command '%s': '%v'. stdout: %s, stderr: %s",
			command, err, stdout.Bytes(), stderr.Bytes())
	}

	writeLog("DEBUG", "STDOUT: \"%s\"", stdout)
	writeLog("DEBUG", "STDERR: \"%s\"", stderr)

	return stdout.String(), stderr.String(), nil
}

func mergeCommands(config *Config, commands ...string) string {
	return strings.Join(commands, config.CommandSeparator)
}

func hash(s string) string {
	sha := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sha[:])
}

func writeLog(level, format string, v ...interface{}) {
	log.Output(2, strings.Join([]string{
		fmt.Sprintf("[%s] [terraform-provider-shell]", level),
		fmt.Sprintf(format, v...),
	}, " "))
}
