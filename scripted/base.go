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
	"sync"
	"text/template"
	"time"
)

type Operation string

var nextResourceId = 1

const (
	Create        Operation = "create"
	Read          Operation = "read"
	Exists        Operation = "exists"
	Update        Operation = "update"
	Delete        Operation = "delete"
	CustomizeDiff Operation = "customizediff"
)

type Scripted struct {
	pc                  *ProviderConfig
	d                   ResourceInterface
	rc                  *ResourceConfig
	op                  Operation
	logging             *Logging
	oldLog              []bool
	piid                int
	riid                int
	oldId               string
	dependenciesMet     bool
	dependenciesMetOnce sync.Once
}

type ChangeMap struct {
	Old map[string]interface{}
	New map[string]interface{}
	Cur map[string]interface{}
}

type EnvironmentChangeMap struct {
	Old map[string]string
	New map[string]string
	Cur map[string]string
}

type JsonContext struct {
	data string
}

type TemplateContext struct {
	*ChangeMap
	Provider      *ProviderConfig
	Operation     Operation
	EmptyString   string
	TriggerString string
	StatePrefix   string
	OutputPrefix  string
	LinePrefix    string
	Output        map[string]interface{}
	State         *ChangeMap
	TemplateName  string
	TemplateNames []string
}

type ResourceConfig struct {
	Context     *ChangeMap
	state       *ChangeMap
	environment *EnvironmentChangeMap
}

type TemplateArg struct {
	name     string
	template string
}

type KVEntry struct {
	key   string
	value interface{}
	err   error
}

type ResourceInterface interface {
	GetChange(string) (interface{}, interface{})
	Get(string) interface{}
	GetOk(string) (interface{}, bool)
	Set(string, interface{}) error
	Id() string
	SetIdErr(string) error
	GetChangedKeysPrefix(string) []string
}

func New(d ResourceInterface, meta interface{}, operation Operation, old bool) (*Scripted, error) {
	s := (&Scripted{
		pc: meta.(*ProviderConfig),
		d:  d,
		rc: &ResourceConfig{
			Context: castConfigChangeMap(d.GetChange("context")),
			state:   castConfigChangeMap(d.GetChange("state")),
		},
		oldId: d.Id(),
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
	s.riid = nextResourceId
	nextResourceId++
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
		rendered, _, err := s.template([]string{fmt.Sprintf("env.%s.%s", prefix, key)}, tpl)
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

func (s *Scripted) Environment() (*EnvironmentChangeMap, error) {
	if s.rc.environment == nil {
		env := castEnvironmentChangeMap(s.d.GetChange("environment"))
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
	s.oldLog = append(s.oldLog, old)
	s.syncOld()
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
		s.log(hclog.Trace, "filtering line", "ctx", "filterLines", "line", line)

		if !isFilled(line) {
			s.log(hclog.Trace, "filtered empty line", "ctx", "filterLines", "line", line)
			continue
		}

		if hasExceptPrefix {
			if strings.HasPrefix(line, exceptPrefix) {
				s.log(hclog.Trace, "filtered line with prefix", "ctx", "filterLines", "prefix", exceptPrefix, "line", line)
				continue
			}
		}
		if hasPrefix {
			if !strings.HasPrefix(line, prefix) {
				s.log(hclog.Trace, "filtered line without prefix", "ctx", "filterLines", "prefix", prefix, "line", line)
				continue
			}
			line = strings.TrimPrefix(line, prefix)
		}
		output <- line
		if Debug {
			s.log(hclog.Trace, "line sent", "ctx", "filterLines", "line", line)
		}
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
		if str, ok := e.value.(string); ok {
			decoded, err := base64.StdEncoding.DecodeString(str)
			if err != nil {
				s.log(hclog.Warn, "error decoding base64", "error", err)
				output <- KVEntry{e.key, "", err}
				continue
			}
			value := string(decoded[:])
			output <- KVEntry{e.key, value, e.err}
		}
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
			output <- KVEntry{key, entry, err}
		}
	}
}

func (s *Scripted) templateExtra(names []string, tpl string, extraCtx map[string]interface{}) (string, *JsonContext, error) {
	name := strings.Join(names, "+")
	t := template.New(name)
	t = t.Delims(s.pc.Templates.LeftDelim, s.pc.Templates.RightDelim)
	t = t.Funcs(TemplateFuncs)
	t, err := t.Parse(tpl)
	if err != nil {
		s.log(hclog.Warn, "error when parsing template", "error", err)
		return "", nil, err
	}
	var buf bytes.Buffer
	ctx := &TemplateContext{
		ChangeMap: &ChangeMap{
			Old: s.rc.Context.Old,
			New: s.rc.Context.New,
			Cur: mergeMaps(s.rc.Context.Cur, extraCtx),
		},
		Provider:      s.pc,
		TemplateName:  name,
		TemplateNames: names,
		Operation:     s.op,
		EmptyString:   s.pc.EmptyString,
		TriggerString: s.pc.Commands.TriggerString,
		StatePrefix:   s.pc.StateLinePrefix,
		LinePrefix:    s.pc.LinePrefix,
		OutputPrefix:  s.pc.OutputLinePrefix,
		Output:        castConfigMap(s.d.Get("output")),
		State:         s.rc.state,
	}
	jsonCtx, err := toJson(ctx)

	if err != nil {
		s.log(hclog.Warn, "error while getting JSON context", "error", err)
		return "", nil, err
	}

	if s.pc.Logging.level == hclog.Trace {
		s.log(hclog.Trace, "rendering template", "name", name, "template", tpl, "context", jsonCtx)
	}
	err = t.Execute(&buf, ctx)
	rendered := buf.String()
	if err != nil {
		s.log(hclog.Warn, "error when executing template", "error", err, "rendered", rendered)
	}
	return rendered, &JsonContext{data: jsonCtx}, err
}

func (s *Scripted) template(names []string, tpl string) (string, *JsonContext, error) {
	return s.templateExtra(names, tpl, map[string]interface{}{})
}

func (s *Scripted) prefixedTemplate(args ...*TemplateArg) (string, *JsonContext, error) {
	var names []string
	var templates []string
	hasAny := false
	for _, arg := range args {
		if isFilled(arg.template) {
			hasAny = true
		}
	}
	if !hasAny {
		return EmptyString, nil, nil
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
	return s.template(names, s.joinCommands(templates...))
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

func (s *Scripted) executeBase(output chan string, env *EnvironmentChangeMap, jsonCtx *JsonContext, commands ...string) error {
	command := s.joinCommands(commands...)
	interpreter, args, err := s.getInterpreter(command)
	cmd := exec.Command(interpreter, args...)
	if isSet(s.pc.Commands.WorkingDirectory) {
		cmd.Dir = s.pc.Commands.WorkingDirectory
	}
	cmd.Env = mapToEnv(env.Cur)
	if s.pc.Commands.Environment.IncludeJsonContext {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", JsonContextEnvKey, jsonCtx.data))
	}

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
		s.log(hclog.Trace, "executing", "interpreter", interpreter, "args", args)
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

func (s *Scripted) executeString(jsonCtx *JsonContext, commands ...string) (string, error) {
	lines := make(chan string)
	output := chToString(lines)
	err := s.execute(lines, jsonCtx, commands...)
	return <-output, err
}

func (s *Scripted) execute(lines chan string, jsonCtx *JsonContext, commands ...string) error {
	env, err := s.Environment()
	if err != nil {
		close(lines)
		return err
	}
	return s.executeBase(lines, env, jsonCtx, commands...)
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
		args = append(args, "id", s.d.Id())
		if s.logging.level <= hclog.Trace {
			args = append(args, "ppid", os.Getppid(), "pid", os.Getpid(), "gid", getGID())
		}
	}
	s.logging.Log(level, msg, args...)
}

func (s *Scripted) ensureId() error {
	if isSet(s.pc.Commands.Templates.Id) {
		defer s.logging.PushDefer("commands", "id")()
		command, jsonCtx, err := s.prefixedTemplate(&TemplateArg{"commands_id", s.pc.Commands.Templates.Id})
		if err != nil {
			return err
		}
		if isFilled(command) {
			s.log(hclog.Debug, "getting resource id")
			stdout, err := s.executeString(jsonCtx, command)
			if err != nil {
				return err
			}
			s.log(hclog.Debug, "setting resource id", "id", stdout)
			s.d.SetIdErr(stdout)
			return nil
		}
	}
	env := castEnvironmentMap(s.d.Get("environment"))
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

func (s *Scripted) outputSetter() (input chan string, doneCh chan bool, saveCh chan bool) {
	input = make(chan string)
	doneCh = make(chan bool)
	saveCh = make(chan bool)

	go func() {
		defer s.logging.PushDefer("ctx", "outputSetter")()
		output := map[string]interface{}{}
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
		save := <-saveCh
		close(saveCh)
		if save {
			setval := terraformify(output)
			s.log(hclog.Debug, "syncing output", "value", output, "setval", fmt.Sprintf("%#v", setval))
			s.d.Set("output", setval)
		}
		doneCh <- true
		close(doneCh)
	}()

	return input, doneCh, saveCh
}

func (s *Scripted) triggerReader() (input chan string, resultCh chan bool) {
	input = make(chan string)
	resultCh = make(chan bool)

	go func() {
		defer s.logging.PushDefer("ctx", "triggerReader")()
		filtered := make(chan string)
		go s.filterLines(input, s.pc.EmptyString, s.pc.EmptyString, filtered)
		wasTriggered := false
		for line := range filtered {
			trigger := s.pc.Commands.TriggerString
			s.log(hclog.Trace, "checking trigger", "line", line, "wasTriggered", wasTriggered, "trigger", trigger, "triggered", line == trigger)
			if wasTriggered {
				continue
			} else if line == trigger {
				wasTriggered = true
				s.log(hclog.Info, "wasTriggered")
			}
		}
		resultCh <- wasTriggered
		close(resultCh)
	}()

	return input, resultCh
}

func (s *Scripted) stateSetter() (input chan string, doneCh chan bool, saveCh chan bool) {
	input = make(chan string)
	doneCh = make(chan bool)
	saveCh = make(chan bool)

	go func() {
		defer s.logging.PushDefer("ctx", "stateSetter")()
		output := make(map[string]interface{})
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
		save := <-saveCh
		close(saveCh)
		if save {
			s.rc.state.New = output
			s.syncState()
		}
		doneCh <- true
		close(doneCh)
	}()

	return input, doneCh, saveCh
}

func (s *Scripted) clearState() {
	s.log(hclog.Trace, "clearing resource.state")
	s.rc.state.New = map[string]interface{}{}
	s.syncState()
}

func (s *Scripted) syncState() {
	s.log(hclog.Debug, "syncing resource.state", "state", s.rc.state.New)
	err := s.d.Set("state", terraformify(s.rc.state.New))
	if err != nil {
		s.log(hclog.Error, "syncing resource.state failed", "error", err)
	}
}

func (s *Scripted) clear() {
	s.log(hclog.Info, "clearing resource")
	s.d.SetIdErr("")
	s.d.Set("output", map[string]interface{}{})
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

func (s *Scripted) checkNeedsUpdate() (bool, error) {
	defer s.logging.PushDefer("commands", "needs_update")()
	onEmpty := func(msg string) (bool, error) {
		s.log(hclog.Trace, msg)
		return false, nil
	}
	if !isSet(s.pc.Commands.Templates.NeedsUpdate) {
		return onEmpty(`"commands_needs_update" is empty, exiting.`)
	}
	command, jsonCtx, err := s.prefixedTemplate(&TemplateArg{"commands_needs_update", s.pc.Commands.Templates.NeedsUpdate})
	if err != nil {
		return false, err
	}
	if !isFilled(command) {
		return onEmpty(`"commands_needs_update" rendered empty, exiting.`)
	}
	s.log(hclog.Info, "checking resource needs update")
	lines, triggered := s.triggerReader()
	err = s.execute(lines, jsonCtx, command)
	return <-triggered, err
}

func (s *Scripted) checkDependenciesMet() (bool, error) {
	var err error
	run := func() {
		defer s.logging.PushDefer("commands", "dependencies")()
		onEmpty := func(msg string) {
			s.log(hclog.Trace, msg)
			s.dependenciesMet = true
			s.log(hclog.Debug, "setting `dependencies_met`", "value", s.dependenciesMet)
		}
		if !isSet(s.pc.Commands.Templates.Dependencies) {
			onEmpty(`"commands_dependencies" is empty, exiting.`)
			return
		}
		command, jsonCtx, e := s.prefixedTemplate(&TemplateArg{"commands_dependencies", s.pc.Commands.Templates.Dependencies})
		if e != nil {
			err = e
			return
		}
		if !isFilled(command) {
			onEmpty(`"commands_dependencies" rendered empty, exiting.`)
			return
		}
		s.log(hclog.Info, "checking resource dependencies met")
		lines, triggered := s.triggerReader()
		err = s.execute(lines, jsonCtx, command)
		s.dependenciesMet = err == nil && <-triggered
		s.log(hclog.Debug, "setting `dependencies_met`", "value", s.dependenciesMet)
	}
	s.dependenciesMetOnce.Do(run)
	return s.dependenciesMet, err
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

func (s *Scripted) rollback() {
	s.log(hclog.Info, "rollback started")
	for _, key := range s.d.GetChangedKeysPrefix("") {
		o, n := s.d.GetChange(key)
		s.log(hclog.Trace, "rolling back value", "key", key, "to", o, "from", n)
		s.d.Set(key, o)
	}
	newId := s.d.Id()
	if s.oldId != newId {
		s.log(hclog.Trace, "rolling back id", "to", s.oldId, "from", newId)
		s.d.SetIdErr(s.oldId)
	}
}

func (s *Scripted) getComputeKeysFromCommand() ([]string, error) {
	var ret []string
	onEmpty := func(msg string) error {
		s.log(hclog.Debug, msg)
		return nil
	}
	defer s.logging.PushDefer("commands", "customizediff_computekeys")()
	if !isSet(s.pc.Commands.Templates.CustomizeDiffComputeKeys) {
		return ret, onEmpty(`"commands_customizediff_computekeys" is empty, exiting.`)
	}
	command, jsonCtx, err := s.prefixedTemplate(&TemplateArg{"commands_customizediff_computekeys", s.pc.Commands.Templates.CustomizeDiffComputeKeys})
	if err != nil {
		return ret, err
	}
	if !isFilled(command) {
		return ret, onEmpty(`"commands_customizediff_computekeys" rendered empty, exiting.`)
	}
	env, err := s.Environment()
	if err != nil {
		return ret, err
	}
	output := make(chan string)
	tokens := make(chan string)

	go func() {
		lines := make(chan string)
		go s.filterLines(output, s.pc.LinePrefix, EmptyString, lines)
		for line := range lines {
			for _, token := range strings.Fields(line) {
				tokens <- token
			}
		}
		close(tokens)
	}()

	s.log(hclog.Info, "getting compute keys", "command", command)
	err = s.executeBase(output, env, jsonCtx, command)
	for token := range tokens {
		ret = append(ret, token)
	}
	return ret, err
}

func (s *Scripted) getRecomputeKeys(prefix string) []string {
	var ret []string
	entries := map[string]bool{}

	for _, key := range s.d.GetChangedKeysPrefix(prefix) {
		entries[key] = true
	}
	for _, key := range s.pc.ComputeOutputKeys {
		key = fmt.Sprintf("output.%s", key)
		if strings.HasPrefix(key, prefix) {
			entries[key] = true
		}
	}
	for _, key := range s.pc.ComputeStateKeys {
		key = fmt.Sprintf("state.%s", key)
		if strings.HasPrefix(key, prefix) {
			entries[key] = true
		}
	}
	for key := range entries {
		ret = append(ret, key)
	}
	return ret
}

func (s *Scripted) getRecomputeKeysExtra(prefix string, extra []string) []string {
	entries := map[string]bool{}
	var ret []string
	for _, key := range s.getRecomputeKeys(prefix) {
		entries[key] = true
	}
	for _, key := range extra {
		if strings.HasPrefix(key, prefix) {
			entries[key] = true
		}
	}
	for key := range entries {
		ret = append(ret, key)
	}
	return ret
}
