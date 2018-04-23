package shell

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os/exec"
	"strings"

	"github.com/armon/circbuf"
	"github.com/hashicorp/terraform/helper/schema"
	"text/template"
	"bytes"
	"encoding/base64"
)

func resourceGenericShellCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	command, err := interpolateCommand(config.CreateCommand, getArguments(d))
	if err != nil {
		return err
	}
	writeLog("DEBUG", "creating generic resource: %s", command)
	_, _, err = runCommand(command, config)
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
			if ! strings.HasPrefix(varline, prefix) {
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

	command, err := interpolateCommand(config.ReadCommand, getArguments(d))
	if err != nil {
		return err
	}
	writeLog("DEBUG", "reading resource: %s", command)
	output, stderr, err := runCommand(command, config)
	if err != nil {
		writeLog("INFO", "command returned error (%s), marking resource deleted: %s, stderr: %s", err, output, stderr)
		d.SetId("")
		return nil
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

	o, n := d.GetChange("context")
	deleteCommand, _ := interpolateCommand(config.DeleteCommand, o.(map[string]interface{}))
	createCommand, _ := interpolateCommand(config.CreateCommand, n.(map[string]interface{}))
	command, err := interpolateCommand(config.UpdateCommand, map[string]interface{}{
		"old":            o,
		"new":            n,
		"delete_command": deleteCommand,
		"create_command": createCommand,
	})
	if err != nil {
		return err
	}
	writeLog("DEBUG", "updating generic resource: %s", command)
	stdout, stderr, err := runCommand(command, config)
	if err != nil {
		writeLog("WARN", "update command returned error: %s, stderr: %s", stdout, stderr)
		return nil
	}

	return resourceShellRead(d, meta)
}

func resourceGenericShellExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	config := meta.(*Config)

	command, err := interpolateCommand(config.ExistsCommand, getArguments(d))
	if err != nil {
		return false, err
	}
	writeLog("DEBUG", "resource exists: %s", command)
	stdout, stderr, err := runCommand(command, config)
	if err != nil {
		writeLog("WARN", "command returned error: %s, stderr: %s", stdout, stderr)
	}

	return stdout == "true", err
}

func resourceGenericShellDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	command, err := interpolateCommand(config.DeleteCommand, getArguments(d))
	if err != nil {
		return err
	}
	writeLog("DEBUG", "deleting generic resource: %s", command)
	_, _, err = runCommand(command, config)
	if err != nil {
		return err
	}

	d.SetId("")
	return nil
}

func getArguments(d *schema.ResourceData) map[string]interface{} {
	return d.Get("context").(map[string]interface{})
}

func interpolateCommand(command string, context map[string]interface{}) (string, error) {
	t := template.Must(template.New("command").Parse(command))
	var buf bytes.Buffer
	err := t.Execute(&buf, context)
	return buf.String(), err
}

func runCommand(command string, config *Config) (string, string, error) {
	// Setup the command
	interpreter := config.Interpreter[0]
	args := append(config.Interpreter[1:], mergeCommands(
		fmt.Sprintf("cd %s", config.WorkingDirectory),
		config.CommandPrefix,
		command,
	))
	cmd := exec.Command(interpreter, args...)
	stdout, _ := circbuf.NewBuffer(config.BufferSize)
	cmd.Stdout = io.Writer(stdout)
	stderr, _ := circbuf.NewBuffer(config.BufferSize)
	cmd.Stderr = io.Writer(stderr)

	// Output what we're about to run
	writeLog("going to execute: %s %s", interpreter, strings.Join(args, " "))

	// Run the command to completion
	err := cmd.Run()

	if err != nil {
		return "", "", fmt.Errorf("error running command '%s': '%v'. stdout: %s, stderr: %s",
			command, err, stdout.Bytes(), stderr.Bytes())
	}

	writeLog("DEBUG", "stdout was: \"%s\"", stdout)
	writeLog("DEBUG", "stderr was: \"%s\"", stderr)

	return stdout.String(), stderr.String(), nil
}

func mergeCommands(commands... string) string {
	return strings.Join(commands, "\n")
}

func hash(s string) string {
	sha := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sha[:])
}

func writeLog(level, format string, v ... interface{}) {
	log.Output(2, strings.Join([]string{
		fmt.Sprintf("[%s] [terraform-provider-shell]", level),
		fmt.Sprintf(format, v...),
	}, " "))
}
