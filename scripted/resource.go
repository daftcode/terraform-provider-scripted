package scripted

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"github.com/armon/circbuf"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/mitchellh/go-linereader"
	"io"
	"os"
	"os/exec"
	"sort"
	"strings"
	"syscall"
	"text/template"
)

type Environment struct {
	old []string
	new []string
}

type State struct {
	c   *Config
	d   *schema.ResourceData
	ctx map[string]interface{}
	env *Environment
	op  string
}

func getScriptedResource() *schema.Resource {
	ret := &schema.Resource{
		Create: resourceScriptedCreate,
		Read:   resourceScriptedRead,
		Update: resourceScriptedUpdate,
		Delete: resourceScriptedDelete,
		Exists: resourceScriptedExists,

		Schema: map[string]*schema.Schema{
			"log_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "Resource name to display in log messages",
				// Hack so it doesn't ever change
				ForceNew: true,
				StateFunc: func(v interface{}) string {
					return ""
				},
			},
			"environment_templates": {
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Environment keys that are themselves templates to be rendered",
			},
			"context": {
				Type:        schema.TypeMap,
				Optional:    true,
				Description: "Template context for rendering commands",
			},
			"environment": {
				Type:        schema.TypeMap,
				Optional:    true,
				Default:     []string{},
				Description: "Environment to run commands in",
			},
			"output": {
				Type:        schema.TypeMap,
				Computed:    true,
				Description: "Output from the read command",
			},
		},
	}
	return ret
}

func getEnv(s *State, environment map[string]interface{}, environmentTemplates []interface{}, old bool) ([]string, error) {
	var env []string
	if s.c.IncludeParentEnvironment {
		env = os.Environ()
	}
	cur := s.ctx["cur"]

	if old {
		s.ctx["cur"] = s.ctx["old"]
	} else {
		s.ctx["cur"] = s.ctx["new"]
	}
	renderedEnv := map[string]string{}
	for _, k := range environmentTemplates {
		key := k.(string)
		tpl := environment[key].(string)
		rendered, err := renderTemplate(tpl, s.ctx)
		if err != nil {
			return nil, err
		}
		writeLog(s, hclog.Debug, "rendering envvar", "key", key, "template", tpl, "rendered", rendered)
		renderedEnv[key] = rendered
	}
	s.ctx["cur"] = cur

	for key, value := range environment {
		rendered, ok := renderedEnv[key]
		if ok {
			value = rendered
		}
		env = append(env, key+"="+value.(string))
	}

	return env, nil
}

func makeEnv(s *State) (*Environment, error) {
	o, n := s.d.GetChange("environment")
	oKeys, nKeys := s.d.GetChange("environment_templates")

	oldEnv, err := getEnv(s, o.(map[string]interface{}), oKeys.([]interface{}), true)
	if err != nil {
		return nil, err
	}
	newEnv, err := getEnv(s, n.(map[string]interface{}), nKeys.([]interface{}), false)
	if err != nil {
		return nil, err
	}
	return &Environment{
		old: oldEnv,
		new: newEnv,
	}, nil
}

func makeState(d *schema.ResourceData, meta interface{}, operation string, old bool) (*State, error) {
	s := &State{
		c:  meta.(*Config),
		d:  d,
		op: operation,
	}
	s.ctx = getContext(s, operation)
	if old {
		s.ctx["cur"] = s.ctx["old"]
	}
	env, err := makeEnv(s)
	s.env = env
	return s, err
}

func copyState(s *State) *State {
	return &State{
		c:   s.c,
		d:   s.d,
		op:  s.op,
		ctx: s.ctx,
		env: &Environment{
			old: s.env.old,
			new: s.env.new,
		},
	}
}

func resourceScriptedCreate(d *schema.ResourceData, meta interface{}) error {
	s, err := makeState(d, meta, "create", false)
	if err != nil {
		return err
	}
	return resourceScriptedCreateBase(s)
}

func resourceScriptedCreateBase(s *State) error {
	command, err := renderTemplate(
		prepareCommands(s, s.c.CommandPrefix, s.c.CreateCommand),
		s.ctx)
	if err != nil {
		return err
	}
	writeLog(s, hclog.Debug, "creating resource")
	_, err = runCommand(s, s.env.new, command)
	if err != nil {
		return err
	}

	s.d.SetId(makeId(s.d, s.env.new))
	writeLog(s, hclog.Debug, "created generic resource", "id", s.d.Id())

	return resourceScriptedReadBase(s)
}

func resourceScriptedRead(d *schema.ResourceData, meta interface{}) error {
	s, err := makeState(d, meta, "read", false)
	if err != nil {
		return err
	}
	return resourceScriptedReadBase(s)
}

func resourceScriptedReadBase(s *State) error {
	command, err := renderTemplate(
		prepareCommands(s, s.c.CommandPrefix, s.c.ReadCommand),
		s.ctx)
	if err != nil {
		return err
	}
	writeLog(s, hclog.Debug, "reading resource")
	stdout, err := runCommand(s, s.env.new, command)
	if err != nil {
		writeLog(s, hclog.Info, "command returned error, marking resource deleted", "error", err, "stdout", stdout)
		if s.c.DeleteOnReadFailure {
			s.d.SetId("")
			return nil
		}
		return err
	}
	var outputs map[string]string

	switch s.c.ReadFormat {
	case "base64":
		outputs = getOutputsBase64(s, stdout, s.c.ReadLinePrefix)
	default:
		fallthrough
	case "raw":
		outputs = getOutputsText(s, stdout, s.c.ReadLinePrefix)
	}
	s.d.Set("output", outputs)

	return nil
}

func resourceScriptedUpdate(d *schema.ResourceData, meta interface{}) error {
	s, err := makeState(d, meta, "update", false)
	if err != nil {
		return err
	}

	if s.c.DeleteBeforeUpdate {
		if err := resourceScriptedDeleteBase(s); err != nil {
			return err
		}
	}

	if s.c.CreateBeforeUpdate {
		if err := resourceScriptedCreateBase(s); err != nil {
			return err
		}
	}

	if s.c.UpdateCommand != "" {
		deleteCommand, _ := renderTemplate(
			wrapCommands(s, s.c.CommandPrefix, s.c.DeleteCommand),
			mergeMaps(s.ctx, map[string]interface{}{"cur": s.ctx["old"]}))
		createCommand, _ := renderTemplate(wrapCommands(s, s.c.CommandPrefix, s.c.CreateCommand), s.ctx)
		command, err := renderTemplate(
			prepareCommands(s, s.c.CommandPrefix, s.c.UpdateCommand),
			mergeMaps(s.ctx, map[string]interface{}{
				"delete_command": deleteCommand,
				"create_command": createCommand,
			}))
		if err != nil {
			return err
		}
		writeLog(s, hclog.Debug, "updating resource", "command", command)
		_, err = runCommand(s, s.env.new, command)
		if err != nil {
			writeLog(s, hclog.Warn, "update command returned error", "error", err)
			return nil
		}
		d.SetId(makeId(d, s.env.new))
	}

	if s.c.CreateAfterUpdate {
		if err := resourceScriptedCreateBase(s); err != nil {
			return err
		}
	}

	return resourceScriptedReadBase(s)
}

func resourceScriptedExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	s, err := makeState(d, meta, "exists", false)
	if err != nil {
		return false, err
	}
	if s.c.ExistsCommand == "" {
		return true, nil
	}
	command, err := renderTemplate(
		prepareCommands(s, s.c.CommandPrefix, s.c.ExistsCommand),
		s.ctx)
	if err != nil {
		return false, err
	}
	writeLog(s, hclog.Debug, "resource exists")
	_, err = runCommand(s, s.env.new, command)
	if err != nil {
		writeLog(s, hclog.Warn, "command returned error", "error", err)
	}
	exists := getExitStatus(err) == s.c.ExistsExpectedStatus
	if s.c.ExistsExpectedStatus == 0 {
		exists = err == nil
	}
	if !exists && s.c.DeleteOnNotExists {
		s.d.SetId("")
	}
	return exists, nil
}

func resourceScriptedDelete(d *schema.ResourceData, meta interface{}) error {
	s, err := makeState(d, meta, "delete", true)
	if err != nil {
		return err
	}
	return resourceScriptedDeleteBase(s)
}

func resourceScriptedDeleteBase(s *State) error {
	s = copyState(s)
	if s.op != "delete" {
		s.ctx = mergeMaps(s.ctx, map[string]interface{}{"cur": s.ctx["old"]})
	}
	command, err := renderTemplate(
		prepareCommands(s, s.c.CommandPrefix, s.c.DeleteCommand),
		s.ctx)
	if err != nil {
		return err
	}
	writeLog(s, hclog.Debug, "reading resource")
	_, err = runCommand(s, s.env.old, command)
	if err != nil {
		return err
	}

	s.d.SetId("")
	return nil
}

func getExitStatus(err error) int {
	if exiterr, ok := err.(*exec.ExitError); ok {
		if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
			return status.ExitStatus()
		}
	}
	return -1
}

func getOutputsText(s *State, output string, prefix string) map[string]string {
	outputs := make(map[string]string)
	split := strings.Split(output, "\n")
	for _, varline := range split {
		writeLog(s, hclog.Debug, "reading output line", "line", varline)

		if varline == "" {
			writeLog(s, hclog.Debug, "skipping empty line")
			continue
		}

		if prefix != "" {
			if !strings.HasPrefix(varline, prefix) {
				writeLog(s, hclog.Info, "ignoring line without prefix", "prefix", prefix, "line", varline)
				continue
			}
			varline = strings.TrimPrefix(varline, prefix)
		}

		pos := strings.Index(varline, "=")
		if pos == -1 {
			writeLog(s, hclog.Info, "ignoring line without equal sign", "line", varline)
			continue
		}

		key := varline[:pos]
		value := varline[pos+1:]
		writeLog(s, hclog.Debug, "read output entry (raw)", "key", key, key, value)
		outputs[key] = value
	}
	return outputs
}

func getOutputsBase64(s *State, output, prefix string) map[string]string {
	outputs := make(map[string]string)
	for key, value := range getOutputsText(s, output, prefix) {
		decoded, err := base64.StdEncoding.DecodeString(value)
		if err != nil {
			writeLog(s, hclog.Warn, "error decoding base64", "error", err)
			continue
		}
		writeLog(s, hclog.Debug, "read output entry (decoded)", "key", key, key, string(decoded[:]), "base64", value)
		outputs[key] = string(decoded[:])
	}
	return outputs
}

func getContext(s *State, operation string) map[string]interface{} {
	o, n := s.d.GetChange("context")
	return map[string]interface{}{
		"operation": operation,
		"output":    s.d.Get("output"),
		"old":       o,
		"new":       n,
		"cur":       n,
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

func renderTemplate(tpl string, context map[string]interface{}) (string, error) {
	t := template.Must(template.New("template").Parse(tpl))
	var buf bytes.Buffer
	err := t.Execute(&buf, context)
	return buf.String(), err
}

func runCommand(s *State, env []string, commands ...string) (string, error) {
	interpreter := s.c.Interpreter[0]
	command := prepareCommands(s, commands...)
	args := append(s.c.Interpreter[1:], command)
	cmd := exec.Command(interpreter, args...)
	cmd.Dir = s.c.WorkingDirectory
	cmd.Env = env

	// Setup the reader that will read the output from the command.
	// We use an os.Pipe so that the *os.File can be passed directly to the
	// process, and not rely on goroutines copying the data which may block.
	// See golang.org/issue/18874
	pr, pw, err := os.Pipe()
	if err != nil {
		return "", fmt.Errorf("failed to initialize pipe for output: %s", err)
	}

	stdout, _ := circbuf.NewBuffer(s.c.BufferSize)
	output, _ := circbuf.NewBuffer(8 * 1024)

	cmd.Stdout = io.MultiWriter(pw, stdout)
	cmd.Stderr = pw

	// Write everything we read from the pipe to the output buffer too
	tee := io.TeeReader(pr, output)

	// copy the teed output to the logger
	copyDoneCh := make(chan struct{})
	go copyOutput(s, tee, copyDoneCh)

	// Output what we're about to run
	writeLog(s, hclog.Debug, "executing", "interpreter", interpreter, "arguments", strings.Join(args, " "))

	// Start the command
	err = cmd.Start()
	if err == nil {
		err = cmd.Wait()
	}

	// Close the write-end of the pipe so that the goroutine mirroring output
	// ends properly.
	pw.Close()

	select {
	case <-copyDoneCh:
	}
	if err != nil {
		return stdout.String(), fmt.Errorf("error running command '%s': %v. Output: %s",
			command, err, output.Bytes())
	}
	return stdout.String(), nil
}

func prepareCommands(s *State, commands ...string) string {
	out := ""
	for _, cmd := range commands {
		out = fmt.Sprintf(s.c.CommandJoiner, out, cmd)
	}
	return out
}

func wrapCommands(s *State, commands ...string) string {
	return fmt.Sprintf(s.c.CommandIsolator, prepareCommands(s, commands...))
}

// Retrieve Id from
func makeId(d *schema.ResourceData, env []string) string {
	var keys []string
	ctx := d.Get("context").(map[string]interface{})
	for k := range ctx {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var entries []string
	for _, k := range keys {
		entries = append(entries, hash(hash(k)+hash(ctx[k].(string))))
	}
	for _, entry := range env {
		entries = append(entries, hash(entry))
	}
	return hash(strings.Join(entries, ""))
}

func hash(s string) string {
	sha := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sha[:])
}

func copyOutput(s *State, r io.Reader, doneCh chan<- struct{}) {
	defer close(doneCh)
	lr := linereader.New(r)
	for line := range lr.Ch {
		format := fmt.Sprintf("<LINE>%%-%ds</LINE>", s.c.CommandLogWidth)
		writeLog(s, s.c.CommandLogLevel, fmt.Sprintf(format, line))
	}
}

func selectLogFunction(logger hclog.Logger, level hclog.Level) func(msg string, args ...interface{}) {
	switch level {
	case hclog.Trace:
		return logger.Trace
	case hclog.Debug:
		return logger.Debug
	case hclog.Info:
		return logger.Info
	case hclog.Warn:
		return logger.Warn
	case hclog.Error:
		return logger.Error
	default:
		return logger.Info
	}
}

func getLogFunction(s *State, level hclog.Level) func(msg string, args ...interface{}) {
	fns := []func(msg string, args ...interface{}){
		selectLogFunction(s.c.Logger, level),
	}

	if s.c.FileLogger != nil {
		fns = append(fns, selectLogFunction(s.c.FileLogger, level))
	}

	return func(msg string, args ...interface{}) {
		for _, v := range fns {
			v(msg, args...)
		}
	}
}

func writeLog(s *State, level hclog.Level, msg string, args ...interface{}) {
	fn := getLogFunction(s, level)
	if s.c.LogProviderName != "" {
		args = append(args, "provider", s.c.LogProviderName)
	}
	resourceName := s.d.Get("log_name").(string)
	if resourceName != "" {
		args = append(args, "resource", resourceName)
	}
	fn(msg, args...)
}
