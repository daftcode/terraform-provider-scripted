package scripted

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"fmt"
	"github.com/armon/circbuf"
	"github.com/hashicorp/go-hclog"
	"io"
	"os"
	"os/exec"
	"strings"
	"text/template"
	"time"
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
	d       ResourceInterface
	rc      *ResourceConfig
	op      Operation
	logging *Logging
	oldLog  []bool
	piid    int
	riid    int
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

type KVEntry struct {
	key   string
	value string
	err   error
}

type ResourceInterface interface {
	GetChange(string) (interface{}, interface{})
	Get(string) interface{}
	GetOk(string) (interface{}, bool)
	Id() string
	Set(string, interface{}) error
	SetIdErr(string) error
}

func New(d ResourceInterface, meta interface{}, operation Operation, old bool) (*Scripted, error) {
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
	s.syncOld()
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
	s.riid = nextResourceId
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
	if s.old() {
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
			if !s.old() {
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
		} else if len(s.pc.Commands.Environment.InheritVariables) > 0 {
			for _, key := range s.pc.Commands.Environment.InheritVariables {
				value := os.Getenv(key)
				if _, ok := env.Old[key]; !ok {
					env.Old[key] = value
				}
				if _, ok := env.New[key]; !ok {
					env.New[key] = value
				}
			}
		}
		if s.old() {
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
}

func (s *Scripted) old() bool {
	l := len(s.oldLog)
	if l == 0 {
		return false
	}
	return s.oldLog[l-1]
}

func (s *Scripted) syncOld() {
	if s.old() {
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
	s.syncOld()
}

func (s *Scripted) removeOld() {
	l := len(s.oldLog)
	s.oldLog = s.oldLog[:l-1]
	s.syncOld()
}

func (s *Scripted) scanLines(lines chan string, reader io.Reader) {
	defer close(lines)
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		lines <- scanner.Text()
	}
}

func (s *Scripted) filterLines(input chan string, prefix, exceptPrefix string, output chan string) {
	defer close(output)
	hasPrefix := isSet(prefix)
	hasExceptPrefix := isSet(exceptPrefix)
	for line := range input {
		s.log(hclog.Trace, "filtering line", "line", line)

		if !isFilled(line) {
			s.log(hclog.Trace, "filtered empty line", "line", line)
			continue
		}

		if hasExceptPrefix {
			if strings.HasPrefix(line, exceptPrefix) {
				s.log(hclog.Trace, "filtered line with prefix", "prefix", exceptPrefix, "line", line)
				continue
			}
		}
		if hasPrefix {
			if !strings.HasPrefix(line, prefix) {
				s.log(hclog.Trace, "filtered line without prefix", "prefix", prefix, "line", line)
				continue
			}
			line = strings.TrimPrefix(line, prefix)
		}
		// s.log(hclog.Trace, "filter passed", "line", line)
		output <- line
		// s.log(hclog.Trace, "filter sent  ", "line", line)
	}
}

func (s *Scripted) scanText(input chan string, output chan KVEntry) {
	defer close(output)

	for line := range input {
		pos := strings.Index(line, "=")
		if pos == -1 {
			s.log(hclog.Debug, "ignoring line without equal sign", "line", line)
			continue
		}

		key := line[:pos]
		value := line[pos+1:]
		s.log(hclog.Trace, "scanned text", "key", key, "value", value)
		output <- KVEntry{key, value, nil}
	}
}

func (s *Scripted) scanBase64(input chan string, output chan KVEntry) {
	defer close(output)
	textEntries := make(chan KVEntry)
	go s.scanText(input, textEntries)

	for e := range textEntries {
		decoded, err := base64.StdEncoding.DecodeString(e.value)
		if err != nil {
			s.log(hclog.Warn, "error decoding base64", "error", err)
			output <- KVEntry{e.key, "", err}
			continue
		}
		value := string(decoded[:])
		output <- KVEntry{e.key, value, e.err}
	}
}

func (s *Scripted) scanJson(input chan string, output chan KVEntry) {
	defer close(output)

	for line := range input {
		if !strings.HasPrefix(line, "{") {
			s.log(hclog.Trace, "not a json line, skipping", "line", line)
			continue
		}
		data, err := fromJson(line)
		if err != nil {
			s.log(hclog.Warn, "invalid json line", "line", line, "error", err)
			continue
		}
		for key, entry := range data.(map[string]interface{}) {
			value, ok := entry.(string)
			err = nil
			if !ok {
				err = fmt.Errorf(`failed to convert %v to string`, entry)
			}
			output <- KVEntry{key, value, err}
		}
	}
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
	hasAny := false
	for _, arg := range args {
		if isFilled(arg.template) {
			hasAny = true
		}
	}
	if !hasAny {
		return EmptyString, nil
	}
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

func (s *Scripted) executeBase(output chan string, env *ChangeMap, commands ...string) error {
	command := s.joinCommands(commands...)
	interpreter, args, err := s.getInterpreter(command)
	cmd := exec.Command(interpreter, args...)
	if isSet(s.pc.Commands.WorkingDirectory) {
		cmd.Dir = s.pc.Commands.WorkingDirectory
	}
	cmd.Env = mapToEnv(env.Cur)

	outBuf, err := circbuf.NewBuffer(s.pc.LoggingBufferSize)
	if err != nil {
		return fmt.Errorf("failed to initialize redirection buffer: %s", err)
	}

	pr, pw := io.Pipe()
	defer pw.Close()
	go s.scanLines(output, pr)

	outLog := newLoggedOutput(s, "out")
	cmd.Stdout = io.MultiWriter(outBuf, outLog.Start(), pw)
	defer outLog.Close()

	errLog := newLoggedOutput(s, "err")
	cmd.Stderr = io.MultiWriter(outBuf, errLog.Start())
	defer errLog.Close()

	// Output what we're about to run
	if s.pc.Logging.level >= hclog.Debug {
		s.log(hclog.Debug, "executing command", "command", command)
	} else {
		logArgs := []interface{}{
			"interpreter", interpreter,
		}
		for i, v := range args {
			logArgs = append(logArgs, fmt.Sprintf("args[%d]", i), v)
		}
		s.log(hclog.Trace, "executing", logArgs...)
	}

	// Start the command
	err = cmd.Start()
	s.log(hclog.Trace, "command started")
	if err == nil {
		s.log(hclog.Trace, "command wait")
		err = cmd.Wait()
		s.log(hclog.Trace, "command waited", "err", err)
	}
	s.log(hclog.Trace, "command finished", "err", err)

	if err != nil {
		return fmt.Errorf("error running command '%s': %v. outBuf: %s",
			command, err, outBuf.Bytes())
	}
	return nil
}

func (s *Scripted) executeString(commands ...string) (string, error) {
	lines := make(chan string)
	output := chToString(lines)
	err := s.execute(lines, commands...)
	return <-output, err
}

func (s *Scripted) execute(lines chan string, commands ...string) error {
	env, err := s.Environment()
	if err != nil {
		close(lines)
		return err
	}
	return s.executeBase(lines, env, commands...)
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
	if s.pc.Commands.Output.LogIids {
		args = append(args, "resource_id", s.d.Id())
	}
	s.logging.Log(level, msg, args...)
}

func (s *Scripted) ensureId() error {
	if isSet(s.pc.Commands.Templates.Id) {
		defer s.logging.PopIf(s.logging.Push("commands", "id"))
		command, err := s.prefixedTemplate(&TemplateArg{"commands_id", s.pc.Commands.Templates.Id})
		if err != nil {
			return err
		}
		if isFilled(command) {
			s.log(hclog.Debug, "getting resource id")
			stdout, err := s.executeString(command)
			if err != nil {
				return err
			}
			s.log(hclog.Debug, "setting resource id", "id", stdout)
			s.d.SetIdErr(stdout)
			return nil
		}
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
	s.d.SetIdErr(value)
	return nil
}

func (s *Scripted) getId() string {
	return s.d.Id()
}

func (s *Scripted) outputSetter() (input chan string, doneCh chan bool) {
	input = make(chan string)
	doneCh = make(chan bool)

	go func() {
		s.log(hclog.Trace, "outputSetter", "input", ToString(input))
		output := map[string]string{}
		filtered := make(chan string)
		go s.filterLines(input, s.pc.OutputLinePrefix, s.pc.StateLinePrefix, filtered)
		entries := make(chan KVEntry)
		go s.scanOutput(filtered, s.pc.OutputFormat, entries)
		for e := range entries {
			if e.err != nil {
				s.log(hclog.Error, "failed getting output", "key", e.key, "value", e.value, "err", e.err)
				continue
			}
			if isSet(e.value) {
				s.log(hclog.Trace, "setting output", "key", e.key, "value", e.value)
				output[e.key] = e.value
			} else {
				s.log(hclog.Trace, "deleting output", "key", e.key)
				delete(output, e.key)
			}
		}
		s.d.Set("output", output)
		doneCh <- true
		close(doneCh)
	}()

	return input, doneCh
}

func (s *Scripted) stateSetter() (input chan string, doneCh chan bool) {
	input = make(chan string)
	doneCh = make(chan bool)

	s.clearState()

	go func() {
		s.log(hclog.Trace, "stateSetter", "input", ToString(input))
		output := s.rc.state.New
		filtered := make(chan string)
		go s.filterLines(input, s.pc.StateLinePrefix, s.pc.EmptyString, filtered)
		entries := make(chan KVEntry)
		go s.scanOutput(filtered, s.pc.StateFormat, entries)
		for e := range entries {
			if e.err != nil {
				s.log(hclog.Error, "failed getting state", "key", e.key, "value", e.value, "err", e.err)
				continue
			}
			if isSet(e.value) {
				s.log(hclog.Trace, "setting state", "key", e.key, "value", e.value)
				output[e.key] = e.value
			} else {
				s.log(hclog.Trace, "deleting state", "key", e.key)
				delete(output, e.key)
			}
		}
		s.syncState()
		doneCh <- true
		close(doneCh)
	}()

	return input, doneCh
}

func (s *Scripted) clearState() {
	s.log(hclog.Trace, "clearing resource.state")
	s.rc.state.New = map[string]string{}
	s.syncState()
}

func (s *Scripted) syncState() {
	s.log(hclog.Debug, "setting resource.state", "state", s.rc.state.New)
	s.d.Set("state", s.rc.state.New)
}

func (s *Scripted) clear() {
	s.log(hclog.Info, "clearing resource")
	s.d.SetIdErr("")
	s.d.Set("output", map[string]string{})
	s.clearState()
}

func (s *Scripted) scanOutput(input chan string, format string, output chan KVEntry) {
	switch format {
	case "json":
		go s.scanJson(input, output)
	case "base64":
		go s.scanBase64(input, output)
	default:
		fallthrough
	case "raw":
		go s.scanText(input, output)
	}
}

func (s *Scripted) checkNeedsUpdate() error {
	defer s.logging.PopIf(s.logging.Push("commands", "needs_update"))
	onEmpty := func(msg string) error {
		s.log(hclog.Trace, msg)
		s.setNeedsUpdate(false)
		return nil
	}
	if !isSet(s.pc.Commands.Templates.NeedsUpdate) {
		return onEmpty(`"commands_needs_update" is empty, exiting.`)
	}
	command, err := s.prefixedTemplate(&TemplateArg{"commands_needs_update", s.pc.Commands.Templates.NeedsUpdate})
	if err != nil {
		return err
	}
	if !isFilled(command) {
		return onEmpty(`"commands_needs_update" rendered empty, exiting.`)
	}
	s.log(hclog.Info, "checking resource needs update")
	output, err := s.executeString(command)
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
	defer s.logging.PopIf(s.logging.Push("commands", "dependencies"))
	onEmpty := func(msg string) (bool, error) {
		s.log(hclog.Trace, msg)
		s.setDependenciesMet(true)
		return true, nil
	}
	if !isSet(s.pc.Commands.Templates.Dependencies) {
		return onEmpty(`"commands_dependencies" is empty, exiting.`)
	}
	command, err := s.prefixedTemplate(&TemplateArg{"commands_dependencies", s.pc.Commands.Templates.Dependencies})
	if err != nil {
		return false, err
	}
	if !isFilled(command) {
		return onEmpty(`"commands_dependencies" rendered empty, exiting.`)
	}
	s.log(hclog.Info, "checking resource dependencies met")
	output, err := s.executeString(command)
	s.setDependenciesMet(err == nil && output == s.pc.Commands.DependenciesTriggerOutput)
	return s.dependenciesMet(), err
}

func (s *Scripted) dependenciesMet() bool {
	v, ok := s.d.GetOk("dependencies_met")
	return ok && v.(bool)
}

func (s *Scripted) setDependenciesMet(value bool) {
	s.log(hclog.Debug, "setting `dependencies_met`", "value", value)
	if !value {
		s.log(hclog.Error, "dependencies not met!")
	}
	s.d.Set("dependencies_met", value)
}

func (s *Scripted) runningMessages() func() {
	if s.pc.RunningMessageInterval <= 0 {
		return func() {}
	}
	interval := time.Duration(s.pc.RunningMessageInterval * float64(time.Second))
	ticker := time.NewTicker(interval)
	go func() {
		start := time.Now()
		for range ticker.C {
			since := time.Since(start)
			if since > 3*interval {
				repr := since.Round(time.Second / 10).String()
				s.log(hclog.Error, fmt.Sprintf("still runnning after %s...", repr), "duration", repr)
			}
		}
		repr := time.Since(start).Round(time.Second / 10).String()
		s.log(hclog.Error, fmt.Sprintf("finished after %s", repr), "duration", repr)
	}()
	return ticker.Stop
}
