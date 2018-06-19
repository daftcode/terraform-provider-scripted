package scripted

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"github.com/armon/circbuf"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/terraform/helper/schema"
	"io"
	"os"
	"os/exec"
	"strings"
	"text/template"
)

type Operation string

var nextResourceId = 1

const (
	Create Operation = "create"
	Read   Operation = "read"
	Exists Operation = "exists"
	Update Operation = "update"
	Delete Operation = "delete"
)

type Scripted struct {
	pc      *ProviderConfig
	d       *schema.ResourceData
	rc      *ResourceConfig
	op      Operation
	logging *Logging
	old     bool
	oldLog  []bool
}

type ChangeMap struct {
	Old map[string]string
	New map[string]string
	Cur map[string]string
}

type TemplateContext struct {
	*ChangeMap
	Operation    Operation
	EmptyString  string
	StatePrefix  string
	OutputPrefix string
	Output       map[string]string
	State        *ChangeMap
}

type ResourceConfig struct {
	// LogName              string
	EnvironmentTemplates []string
	Context              *ChangeMap
	state                *ChangeMap
	environment          *ChangeMap
}

type TemplateArg struct {
	name     string
	template string
}

func New(d *schema.ResourceData, meta interface{}, operation Operation, old bool) (*Scripted, error) {
	s := (&Scripted{
		pc: meta.(*ProviderConfig),
		d:  d,
		rc: &ResourceConfig{
			// LogName: d.Get("log_name").(string),
			Context: castConfigChangeMap(d.GetChange("context")),
			state:   castConfigChangeMap(d.GetChange("state")),
		},
	}).setOperation(operation)
	s.ensureLogging()
	s.setOld(old)
	s.log(hclog.Trace, "resource initialized")
	s.log(hclog.Trace, "initialized state", "old", s.rc.state.Old, "new", s.rc.state.New)

	return s, nil
}

func (s *Scripted) ensureLogging() *Scripted {
	s.logging = s.pc.Logging.Clone()

	args := []interface{}{
		"operation", s.op,
	}
	if s.pc.Commands.Output.LogIids {
		args = append(args, "riid", nextResourceId)
	}
	nextResourceId++
	// if s.rc.LogName != "" {
	// 	args = append(args, "resource", s.rc.LogName)
	// }
	s.logging.Push(args...)
	return s
}

func (s *Scripted) renderEnv(old bool) error {
	s.addOld(old)
	defer s.removeOld()
	env := s.rc.environment.Cur

	var prefix string
	if s.old {
		prefix = "old"
	} else {
		prefix = "new"
	}
	for key, tpl := range env {
		if !strings.Contains(tpl, s.pc.Templates.LeftDelim) {
			continue
		}
		rendered, err := s.template(fmt.Sprintf("env.%s.%s", prefix, key), tpl)
		if err != nil {
			if !s.old {
				return err
			}
			rendered = fmt.Sprintf("<ERROR: %s>", err.Error())
		}
		env[key] = rendered
	}
	return nil
}

func (s *Scripted) Environment() (*ChangeMap, error) {
	if s.rc.environment == nil {
		env := castConfigChangeMap(s.d.GetChange("environment"))
		if s.pc.Commands.Environment.IncludeParent {
			for _, line := range os.Environ() {
				split := strings.SplitN(line, "=", 1)
				key := split[0]
				value := ""
				if len(split) > 1 {
					value = split[1]
				}
				if _, ok := env.Old[key]; !ok {
					env.Old[key] = value
				}
				if _, ok := env.New[key]; !ok {
					env.New[key] = value
				}
			}
		}
		if s.old {
			env.Cur = env.Old
		} else {
			env.Cur = env.New
		}
		s.rc.environment = env

		if err := s.renderEnv(true); err != nil {
			s.rc.environment = nil
			return nil, err
		}
		if err := s.renderEnv(false); err != nil {
			s.rc.environment = nil
			return nil, err
		}

		extra := map[string]string{}

		if isSet(s.pc.Commands.Environment.PrefixOld) {
			for k, v := range env.Old {
				key := fmt.Sprintf("%s%s", s.pc.Commands.Environment.PrefixOld, k)
				extra[key] = v
			}
		}

		if isSet(s.pc.Commands.Environment.PrefixNew) {
			for k, v := range env.New {
				key := fmt.Sprintf("%s%s", s.pc.Commands.Environment.PrefixNew, k)
				extra[key] = v
			}
		}
		for k, v := range extra {
			env.Old[k] = v
			env.New[k] = v
		}
	}
	return s.rc.environment, nil
}

func (s *Scripted) setOperation(operation Operation) *Scripted {
	s.op = operation
	return s
}

func (s *Scripted) setOld(old bool) {
	s.old = old
	if old {
		s.rc.Context.Cur = s.rc.Context.Old
		if s.rc.environment != nil {
			s.rc.environment.Cur = s.rc.environment.Old
		}
	} else {
		s.rc.Context.Cur = s.rc.Context.New
		if s.rc.environment != nil {
			s.rc.environment.Cur = s.rc.environment.New
		}
	}
}

func (s *Scripted) addOld(old bool) {
	s.oldLog = append(s.oldLog, old)
	s.setOld(old)
}

func (s *Scripted) removeOld() {
	l := len(s.oldLog)
	s.setOld(s.oldLog[l-1])
	s.oldLog = s.oldLog[:l-1]
}

func (s *Scripted) outputsText(output string, prefix, exceptPrefix string, outputs map[string]string) map[string]string {
	split := strings.Split(output, "\n")
	hasPrefix := isSet(prefix)
	hasExceptPrefix := isSet(exceptPrefix)

	for _, varline := range split {
		s.log(hclog.Trace, "reading output line", "line", varline)

		if varline == "" {
			s.log(hclog.Debug, "skipping empty line")
			continue
		}

		if hasExceptPrefix {
			if strings.HasPrefix(varline, exceptPrefix) {
				s.log(hclog.Debug, "ignoring line with prefix", "prefix", exceptPrefix, "line", varline)
				continue
			}
		}
		if hasPrefix {
			if !strings.HasPrefix(varline, prefix) {
				s.log(hclog.Debug, "ignoring line without prefix", "prefix", prefix, "line", varline)
				continue
			}
			varline = strings.TrimPrefix(varline, prefix)
		}

		pos := strings.Index(varline, "=")
		if pos == -1 {
			s.log(hclog.Debug, "ignoring line without equal sign", "line", varline)
			continue
		}

		key := varline[:pos]
		value := varline[pos+1:]
		s.log(hclog.Info, "read output", "key", key, "value", value)
		outputs[key] = value
	}
	return outputs
}

func (s *Scripted) outputsBase64(output, prefix, exceptPrefix string, outputs map[string]string) map[string]string {
	for key, value := range s.outputsText(output, prefix, exceptPrefix, outputs) {
		decoded, err := base64.StdEncoding.DecodeString(value)
		if err != nil {
			s.log(hclog.Warn, "error decoding base64", "error", err)
			continue
		}
		s.log(hclog.Debug, "read output entry (decoded)", "key", key, key, string(decoded[:]), "base64", value)
		outputs[key] = string(decoded[:])
	}
	return outputs
}

func (s *Scripted) templateExtra(name, tpl string, extraCtx map[string]string) (string, error) {
	s.log(hclog.Trace, "rendering template", "name", name, "template", tpl)
	t := template.New(name)
	t = t.Delims(s.pc.Templates.LeftDelim, s.pc.Templates.RightDelim)
	t = t.Funcs(TemplateFuncs)
	t, err := t.Parse(tpl)
	if err != nil {
		s.log(hclog.Warn, "error when parsing template", "error", err)
		return "", err
	}
	var buf bytes.Buffer
	ctx := &TemplateContext{
		ChangeMap: &ChangeMap{
			Old: s.rc.Context.Old,
			New: s.rc.Context.New,
			Cur: mergeMaps(s.rc.Context.Cur, extraCtx),
		},
		Operation:    s.op,
		EmptyString:  s.pc.EmptyString,
		StatePrefix:  s.pc.StateLinePrefix,
		OutputPrefix: s.pc.OutputLinePrefix,
		Output:       castConfigMap(s.d.Get("output")),
		State:        s.rc.state,
	}
	err = t.Execute(&buf, ctx)
	rendered := buf.String()
	if err != nil {
		s.log(hclog.Warn, "error when executing template", "error", err, "rendered", rendered)
	}
	return rendered, err
}

func (s *Scripted) template(name, tpl string) (string, error) {
	return s.templateExtra(name, tpl, map[string]string{})
}

func (s *Scripted) prefixedTemplate(args ...*TemplateArg) (string, error) {
	var names []string
	var templates []string
	if isFilled(s.pc.Commands.Templates.PrefixFromEnv) {
		names = append(names, "commands_prefix_fromenv")
		templates = append(templates, s.pc.Commands.Templates.PrefixFromEnv)
	}
	if isFilled(s.pc.Commands.Templates.Prefix) {
		names = append(names, "commands_prefix")
		templates = append(templates, s.pc.Commands.Templates.Prefix)
	}
	for _, arg := range args {
		if isFilled(arg.template) {
			names = append(names, arg.name)
			templates = append(templates, arg.template)
		}
	}
	return s.template(strings.Join(names, "+"), s.joinCommands(templates...))
}

func (s *Scripted) getInterpreter(command string) (string, []string, error) {
	var args []string
	hadTemplate := false
	for _, value := range s.pc.Commands.Templates.Interpreter[1:] {
		if strings.Contains(value, s.pc.Templates.LeftDelim) {
			hadTemplate = true
			t := template.New("commands_interpreter")
			t = t.Delims(s.pc.Templates.LeftDelim, s.pc.Templates.RightDelim)
			t, err := t.Parse(value)
			if err != nil {
				return "", nil, err
			}
			var buf bytes.Buffer
			err = t.Execute(&buf, map[string]string{
				"command": command,
			})
			if err != nil {
				return "", nil, err
			}
			value = buf.String()
		}
		args = append(args, value)
	}
	if !hadTemplate {
		args = append(args, command)
	}
	return s.pc.Commands.Templates.Interpreter[0], args, nil
}

func (s *Scripted) executeEnv(env *ChangeMap, commands ...string) (string, error) {
	command := s.joinCommands(commands...)
	interpreter, args, err := s.getInterpreter(command)
	cmd := exec.Command(interpreter, args...)
	if isSet(s.pc.Commands.WorkingDirectory) {
		cmd.Dir = s.pc.Commands.WorkingDirectory
	}
	cmd.Env = mapToEnv(env.Cur)

	output, err := circbuf.NewBuffer(8 * 1024)
	if err != nil {
		return "", fmt.Errorf("failed to initialize redirection buffer: %s", err)
	}

	stdout, _ := circbuf.NewBuffer(s.pc.LoggingBufferSize)

	outLog, err := newLoggedOutput(s, "out")
	errLog, err := newLoggedOutput(s, "err")
	cmd.Stdout = io.MultiWriter(output, outLog.Start(), stdout)
	cmd.Stderr = io.MultiWriter(output, errLog.Start())

	logArgs := []interface{}{
		"interpreter", interpreter,
	}
	for i, v := range args {
		logArgs = append(logArgs, fmt.Sprintf("args[%d]", i), v)
	}
	// Output what we're about to run
	if s.pc.Commands.Output.LogLevel >= hclog.Debug {
		s.log(hclog.Debug, "executing command", "command", command)
	} else {
		s.log(hclog.Trace, "executing", logArgs...)
	}

	// Start the command
	err = cmd.Start()
	if err == nil {
		err = cmd.Wait()
	}

	outLog.Close()
	errLog.Close()

	if err != nil {
		return stdout.String(), fmt.Errorf("error running command '%s': %v. output: %s",
			command, err, output.Bytes())
	}
	return stdout.String(), nil
}

func (s *Scripted) execute(commands ...string) (string, error) {
	env, err := s.Environment()
	if err != nil {
		return "", err
	}
	return s.executeEnv(env, commands...)
}

func (s *Scripted) joinCommands(commands ...string) string {
	out := ""
	for _, cmd := range commands {
		isEmpty := isSet(cmd)
		if out == "" && isEmpty {
			out = cmd
		} else if isEmpty {
			out = fmt.Sprintf(s.pc.Commands.Separator, out, cmd)
		}
	}
	return out
}

func (s *Scripted) log(level hclog.Level, msg string, args ...interface{}) {
	s.logging.Log(level, msg, args...)
}

func (s *Scripted) ensureId() error {
	if isSet(s.pc.Commands.Templates.Id) {
		defer s.logging.PopIf(s.logging.Push("id", true))
		command, err := s.prefixedTemplate(&TemplateArg{"commands_id", s.pc.Commands.Templates.Id})
		if err != nil {
			return err
		}
		stdout, err := s.execute(command)
		if err != nil {
			return err
		}
		s.log(hclog.Debug, "setting resource id", "id", stdout)
		s.d.SetId(stdout)
		return nil
	}
	env := castConfigMap(s.d.Get("environment"))
	var entries []string
	entries = append(entries, getMapHash(s.d.Get("context").(map[string]interface{}))...)
	entries = append(entries, getMapHash(s.d.Get("state").(map[string]interface{}))...)
	for _, entry := range env {
		entries = append(entries, hash(entry))
	}

	value := hash(strings.Join(entries, ""))
	s.log(hclog.Debug, "setting resource id", "id", value)
	s.d.SetId(value)
	return nil
}

func (s *Scripted) getId() string {
	return s.d.Id()
}

func (s *Scripted) setOutput(stdout string) {
	output := s.readLines(stdout, s.pc.OutputLinePrefix, s.pc.OutputFormat, s.pc.StateLinePrefix, map[string]string{})
	s.log(hclog.Debug, "setting output", "output", output)
	s.d.Set("output", output)
}

func (s *Scripted) updateState(stdout string) {
	s.log(hclog.Debug, "updating state")
	s.readLines(stdout, s.pc.StateLinePrefix, s.pc.StateFormat, s.pc.EmptyString, s.rc.state.New)
	for key, value := range s.rc.state.New {
		if value == s.pc.EmptyString {
			delete(s.rc.state.New, key)
		}
	}
	s.syncState()
}

func (s *Scripted) clearState() {
	s.log(hclog.Debug, "clearing state")
	s.rc.state.New = map[string]string{}
	s.syncState()
}

func (s *Scripted) syncState() {
	s.log(hclog.Debug, "setting state", "state", s.rc.state.New)
	s.d.Set("state", s.rc.state.New)
}

func (s *Scripted) clear() {
	s.log(hclog.Debug, "clearing resource")
	s.d.SetId("")
	s.d.Set("output", map[string]string{})
	s.clearState()
}

func (s *Scripted) readLines(data, prefix, format, exceptPrefix string, outputs map[string]string) map[string]string {
	if data == "" {
		return outputs
	}
	switch format {
	case "base64":
		outputs = s.outputsBase64(data, prefix, exceptPrefix, outputs)
	default:
		fallthrough
	case "raw":
		outputs = s.outputsText(data, prefix, exceptPrefix, outputs)
	}
	return outputs
}

func (s *Scripted) checkNeedsUpdate() error {
	defer s.logging.PopIf(s.logging.Push("needs_update", true))
	if !isSet(s.pc.Commands.Templates.NeedsUpdate) {
		s.log(hclog.Debug, `"commands_needs_update" is empty, exiting.`)
		s.setNeedsUpdate(false)
		return nil
	}
	command, err := s.prefixedTemplate(&TemplateArg{"commands_needs_update", s.pc.Commands.Templates.NeedsUpdate})
	if err != nil {
		return err
	}
	s.log(hclog.Debug, "resource needs_update")
	output, err := s.execute(command)
	s.setNeedsUpdate(err == nil && output == s.pc.Commands.NeedsUpdateExpectedOutput)
	return err
}

func (s *Scripted) needsUpdate() bool {
	v, ok := s.d.GetOk("needs_update")
	return !ok || v.(bool)
}

func (s *Scripted) setNeedsUpdate(value bool) {
	s.log(hclog.Debug, "setting `needs_update`", "value", value)
	s.d.Set("needs_update", value)
}

func (s *Scripted) checkDependenciesMet() (bool, error) {
	defer s.logging.PopIf(s.logging.Push("dependencies", true))
	if !isSet(s.pc.Commands.Templates.Dependencies) {
		s.log(hclog.Debug, `"commands_dependencies" is empty, exiting.`)
		s.setDependenciesMet(true)
		return true, nil
	}
	command, err := s.prefixedTemplate(&TemplateArg{"commands_dependencies", s.pc.Commands.Templates.Dependencies})
	if err != nil {
		return false, err
	}
	output, err := s.execute(command)
	s.setDependenciesMet(err == nil && output == s.pc.Commands.DependenciesTriggerOutput)
	return s.dependenciesMet(), err
}

func (s *Scripted) dependenciesMet() bool {
	v, ok := s.d.GetOk("dependencies_met")
	return ok && v.(bool)
}

func (s *Scripted) setDependenciesMet(value bool) {
	s.log(hclog.Debug, "setting `dependencies_met`", "value", value)
	s.d.Set("dependencies_met", value)
}

func (s *Scripted) checkNeedsDelete() (bool, error) {
	defer s.logging.PopIf(s.logging.Push("needs_delete", true))
	if !isSet(s.pc.Commands.Templates.NeedsDelete) {
		s.log(hclog.Debug, `"commands_needs_delete" is empty, exiting.`)
		s.setNeedsDelete(false)
		return false, nil
	}
	command, err := s.prefixedTemplate(&TemplateArg{"commands_needs_delete", s.pc.Commands.Templates.NeedsDelete})
	if err != nil {
		return false, err
	}
	output, err := s.execute(command)
	s.setNeedsDelete(err == nil && output == s.pc.Commands.NeedsDeleteExpectedOutput)
	return s.needsDelete(), err
}

func (s *Scripted) needsDelete() bool {
	v, ok := s.d.GetOk("needs_delete")
	return ok && v.(bool)
}

func (s *Scripted) setNeedsDelete(value bool) {
	s.log(hclog.Debug, "setting `needs_delete`", "value", value)
	s.d.Set("needs_delete", value)
}
