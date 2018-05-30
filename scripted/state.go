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
	"sort"
	"strings"
	"text/template"
)

type Operation string

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
	loggers *Loggers
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
	StatePrefix  string
	OutputPrefix string
	Output       map[string]string
	State        *ChangeMap
}

type ResourceConfig struct {
	LogName              string
	EnvironmentTemplates []string
	Context              *ChangeMap
	state                *ChangeMap
	environment          *ChangeMap
}

func NewState(d *schema.ResourceData, meta interface{}, operation Operation, old bool) (*Scripted, error) {
	s := (&Scripted{
		pc: meta.(*ProviderConfig),
		d:  d,
		rc: &ResourceConfig{
			LogName: d.Get("log_name").(string),
			Context: castConfigChangeMap(d.GetChange("context")),
			state:   castConfigChangeMap(d.GetChange("state")),
		},
	}).setOperation(operation)
	s.ensureLoggers()
	s.setOld(old)
	s.log(hclog.Trace, "resource initialized")
	s.log(hclog.Trace, "initialized state", "old", s.rc.state.Old, "new", s.rc.state.New)
	return s, nil
}

func (s *Scripted) ensureLoggers() *Scripted {
	args := []interface{}{
		"operation", s.op,
	}
	if s.rc.LogName != "" {
		args = append(args, "resource", s.rc.LogName)
	}
	s.loggers = initLoggers(s, args...)
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
		if !strings.Contains(tpl, s.pc.TemplatesLeftDelim) {
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
		if s.pc.IncludeParentEnvironment {
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

		if s.pc.OldEnvironmentPrefix != "" {
			for k, v := range env.Old {
				key := fmt.Sprintf("%s%s", s.pc.OldEnvironmentPrefix, k)
				extra[key] = v
			}
		}

		if s.pc.NewEnvironmentPrefix != "" {
			for k, v := range env.New {
				key := fmt.Sprintf("%s%s", s.pc.NewEnvironmentPrefix, k)
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

func (s *Scripted) outputsText(output string, prefix string, outputs map[string]string) map[string]string {
	split := strings.Split(output, "\n")
	for _, varline := range split {
		s.log(hclog.Trace, "reading output line", "line", varline)

		if varline == "" {
			s.log(hclog.Debug, "skipping empty line")
			continue
		}

		if prefix != "" {
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

func (s *Scripted) outputsBase64(output, prefix string, outputs map[string]string) map[string]string {
	for key, value := range s.outputsText(output, prefix, outputs) {
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
	t = t.Delims(s.pc.TemplatesLeftDelim, s.pc.TemplatesRightDelim)
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

func (s *Scripted) executeEnv(env *ChangeMap, commands ...string) (string, error) {
	interpreter := s.pc.Interpreter[0]
	command := s.prepareCommands(commands...)
	args := append(s.pc.Interpreter[1:], command)
	cmd := exec.Command(interpreter, args...)
	cmd.Dir = s.pc.WorkingDirectory
	cmd.Env = mapToEnv(env.Cur)

	stdout, _ := circbuf.NewBuffer(s.pc.BufferSize)
	output, _ := circbuf.NewBuffer(8 * 1024)

	outLog, err := newLoggedOutput(s, "out")
	errLog, err := newLoggedOutput(s, "err")
	cmd.Stdout = io.MultiWriter(outLog.Start(), stdout)
	cmd.Stderr = errLog.Start()

	logArgs := []interface{}{
		"interpreter", interpreter,
	}
	for i, v := range args {
		logArgs = append(logArgs, fmt.Sprintf("args[%d]", i), v)
	}
	// Output what we're about to run
	s.log(hclog.Debug, "executing", logArgs...)

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

func (s *Scripted) prepareCommands(commands ...string) string {
	out := ""
	for _, cmd := range commands {
		if out == "" {
			out = cmd
		} else if cmd != "" {
			out = fmt.Sprintf(s.pc.CommandJoiner, out, cmd)
		}
	}
	return out
}

func (s *Scripted) log(level hclog.Level, msg string, args ...interface{}) {
	resourceName := s.d.Get("log_name").(string)
	if resourceName != "" {
		args = append(args, "resource", resourceName)
	}
	s.loggers.Log(level, msg, args...)
}

func (s *Scripted) ensureId() error {
	if s.pc.IdCommand != "" {
		defer s.loggers.PopIf(s.loggers.Push("id", true))
		command, err := s.template(
			"command_prefix+id_command",
			s.prepareCommands(s.pc.CommandPrefix, s.pc.IdCommand),
		)
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

	var keys []string
	ctx := s.d.Get("context").(map[string]interface{})
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

	value := hash(strings.Join(entries, ""))
	s.log(hclog.Debug, "setting resource id", "id", value)
	s.d.SetId(value)
	return nil
}

func (s *Scripted) getId() string {
	return s.d.Id()
}

func (s *Scripted) setOutput(stdout string) {
	output := s.readLines(stdout, s.pc.OutputLinePrefix, map[string]string{})
	s.log(hclog.Debug, "setting output", "output", output)
	s.d.Set("output", output)
}

func (s *Scripted) updateState(stdout string) {
	s.log(hclog.Debug, "updating state")
	s.readLines(stdout, s.pc.StateLinePrefix, s.rc.state.New)
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

func (s *Scripted) readLines(data, prefix string, outputs map[string]string) map[string]string {
	if data == "" {
		return outputs
	}
	switch s.pc.ReadFormat {
	case "base64":
		outputs = s.outputsBase64(data, prefix, outputs)
	default:
		fallthrough
	case "raw":
		outputs = s.outputsText(data, prefix, outputs)
	}
	return outputs
}
