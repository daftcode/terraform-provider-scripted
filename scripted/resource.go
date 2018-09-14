package scripted

import (
	"fmt"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
	"reflect"
)

var resourceSchema = getResourceSchema()

func getResourceSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"context": {
			Type:        schema.TypeMap,
			Optional:    true,
			Description: "Template context for rendering commands",
			Sensitive:   true,
		},
		"environment": {
			Type:        schema.TypeMap,
			Optional:    true,
			Description: "Environment to run commands in",
			Sensitive:   true,
		},
		"output": {
			Type:        schema.TypeMap,
			Computed:    true,
			Description: "Output from the read command",
			Sensitive:   true,
		},
		"state": {
			Type:        schema.TypeMap,
			Computed:    true,
			Description: "Output from create/update commands. Set key: `echo '{{ .StatePrefix }}key=value'`. Delete key: `echo '{{ .StatePrefix }}key={{ .EmptyString }}'`",
			Sensitive:   true,
		},
		"revision": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Resource's revision",
		},
	}
}

func getScriptedResource() *schema.Resource {
	ret := &schema.Resource{
		SchemaVersion: 2,
		MigrateState:  stateMigrateFunc,

		Importer: &schema.ResourceImporter{State: schema.ImportStatePassthrough},
		Create:   resourceScriptedCreate,
		Read:     resourceScriptedRead,
		Update:   resourceScriptedUpdate,
		Delete:   resourceScriptedDelete,
		Exists:   resourceScriptedExists,

		Schema: getResourceSchema(),

		CustomizeDiff: resourceScriptedCustomizeDiff,
	}
	return ret
}

func hasChanged(a, b interface{}) bool {
	if eq, ok := a.(schema.Equal); ok {
		return !eq.Equal(b)
	}

	return !reflect.DeepEqual(a, b)
}

func stateMigrateFunc(version int, state *terraform.InstanceState, i interface{}) (*terraform.InstanceState, error) {
	if _, ok := state.Attributes["revision"]; !ok {
		state.Attributes["revision"] = "0"
	}
	if _, ok := state.Attributes["update_trigger"]; ok {
		delete(state.Attributes, "update_trigger")
	}
	return state, nil
}

func resourceScriptedCustomizeDiff(diff *schema.ResourceDiff, i interface{}) error {
	s, err := New(WrapResourceDiff(diff), i, OperationCustomizeDiff, false)
	if err != nil {
		return err
	}
	var computedKeys []string
	if computedKeys, err = s.getComputeKeysFromCommand(); err != nil {
		return err
	}
	computedKeys = append(computedKeys, "revision")
	computedKeys = s.getRecomputeKeysExtra(computedKeys, "output", "state")

	changed := s.d.IsNew()

	if !s.d.IsNew() {
		var allDiffKeys []string
		allDiffKeys = append(allDiffKeys, mergeStringSlices(
			diff.GetChangedKeysPrefix("context"),
			diff.GetChangedKeysPrefix("environment"),
			computedKeys,
		)...)
		shouldLog := s.logging.level <= hclog.Debug

		vDiff := make(map[string]map[string]interface{})

		for _, key := range allDiffKeys {
			o, n := s.d.GetChange(key)
			if !diff.NewValueKnown(key) {
				changed = true
				if shouldLog {
					vDiff[key] = map[string]interface{}{"old": o, "new": n, "newKnown": false}
				}
			} else if hasChanged(o, n) {
				changed = true
				if shouldLog {
					vDiff[key] = map[string]interface{}{"old": o, "new": n, "newKnown": true}
				}
			}
		}

		if shouldLog {
			s.log(hclog.Debug, "customize diff", "diff", toJsonMust(vDiff))
		}
	}

	if hasChanged(s.d.GetChange("revision")) {
		changed = true
	}

	if !changed {
		if needsUpdate, err := s.checkNeedsUpdate(); err != nil {
			return err
		} else if needsUpdate {
			changed = true
		}
	}

	if changed {
		s.log(hclog.Info, "update triggered")
		s.bumpRevision()
		for _, key := range computedKeys {
			s.log(hclog.Trace, "setting key as computed", "key", key)
			diff.SetNewComputed(key)
		}
	} else {
		for _, key := range s.getRecomputeKeysExtra(computedKeys, "output") {
			s.log(hclog.Trace, "setting key as computed", "key", key)
			diff.SetNewComputed(key)
		}
	}

	return nil
}

func resourceScriptedCreate(d *schema.ResourceData, meta interface{}) error {
	s, err := New(WrapResourceData(d), meta, OperationCreate, false)
	if err != nil {
		return err
	}
	if met, err := s.checkDependenciesMet(); err != nil {
		return err
	} else if !met {
		s.rollback()
		return fmt.Errorf("Dependency not met, create failed.")
	}
	err = resourceScriptedCreateBase(s)
	if err != nil {
		return err
	}
	if err := resourceScriptedReadBase(s); err != nil {
		s.rollback()
		return err
	}
	s.bumpRevision()
	return nil
}

func resourceScriptedRead(d *schema.ResourceData, meta interface{}) error {
	s, err := New(WrapResourceData(d), meta, OperationRead, false)
	if err != nil {
		return err
	}
	defer s.runningMessages()()
	if met, err := s.checkDependenciesMet(); err != nil {
		return err
	} else if !met {
		s.rollback()
		return nil
	}

	if err := resourceScriptedReadBase(s); err != nil {
		s.rollback()
		return err
	}
	return nil
}

func resourceScriptedUpdate(d *schema.ResourceData, meta interface{}) error {
	s, err := New(WrapResourceData(d), meta, OperationUpdate, false)
	if err != nil {
		return err
	}
	err = func() error {
		defer s.runningMessages()()
		if met, err := s.checkDependenciesMet(); err != nil {
			return err
		} else if !met {
			s.rollback()
			return fmt.Errorf("Dependency not met, update failed.")
		}
		if isSet(s.pc.Commands.Templates.Update) {
			err = resourceScriptedUpdateBase(s)
			if err != nil {
				return err
			}
		} else {
			if err := resourceScriptedDeleteBase(s); err != nil {
				return err
			}
			if err := resourceScriptedCreateBase(s); err != nil {
				return err
			}
		}

		return resourceScriptedReadBase(s)
	}()
	if err != nil {
		s.rollback()
	}
	s.bumpRevision()
	return err
}

func resourceScriptedExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	s, err := New(WrapResourceData(d), meta, OperationExists, false)
	if err != nil {
		return true, err
	}
	defer s.runningMessages()()
	if met, err := s.checkDependenciesMet(); err != nil {
		return true, err
	} else if !met {
		return true, nil
	}
	defer s.logging.PushDefer("commands", "exists")()

	if !isSet(s.pc.Commands.Templates.Exists) {
		s.log(hclog.Debug, fmt.Sprintf(`"%s" is empty, exiting.`, CommandExists))
		return true, nil
	}
	command, jsonCtx, err := s.prefixedTemplate(&TemplateArg{CommandExists, s.pc.Commands.Templates.Exists})
	if err != nil {
		return false, err
	}
	if !isFilled(command) {
		return true, nil
	}
	s.log(hclog.Info, "checking resource exists")
	lines, triggerCh := s.triggerReader()
	err = s.execute(lines, jsonCtx, command)
	triggered := <-triggerCh
	missing := triggered
	if err != nil {
		s.log(hclog.Warn, "exists returned error", "error", err)
		missing = true
	} else if missing && s.pc.Commands.DeleteOnNotExists {
		s.clear()
	}

	s.log(hclog.Debug, "resource exists result", "exists", !missing, "triggered", triggered, "err", err)
	return !missing, err
}

func resourceScriptedDelete(d *schema.ResourceData, meta interface{}) error {
	s, err := New(WrapResourceData(d), meta, OperationDelete, true)
	if err != nil {
		return err
	}
	defer s.runningMessages()()
	if met, err := s.checkDependenciesMet(); err != nil {
		return err
	} else if !met {
		s.rollback()
		return nil
	}
	if err := resourceScriptedDeleteBase(s); err != nil {
		s.rollback()
		return err
	}
	return nil
}

func resourceScriptedCreateBase(s *Scripted) error {
	defer s.logging.PushDefer("commands", "create")()
	onEmpty := func(msg string) error {
		if isSet(s.pc.Commands.Templates.Update) {
			s.log(hclog.Debug, fmt.Sprintf(`"%s" is empty, running "%s" instead.`, CommandCreate, CommandUpdate))
			return resourceScriptedUpdateBase(s)
		}
		s.log(hclog.Debug, msg)
		if err := s.ensureId(); err != nil {
			return err
		}
		s.clearState()
		return nil
	}

	if !isSet(s.pc.Commands.Templates.Create) {
		return onEmpty(fmt.Sprintf(`"%s" is empty, exiting.`, CommandCreate))
	}
	command, jsonCtx, err := s.prefixedTemplate(
		&TemplateArg{"commands_modify_prefix", s.pc.Commands.Templates.ModifyPrefix},
		&TemplateArg{CommandCreate, s.pc.Commands.Templates.Create},
	)
	if err != nil {
		return err
	}

	if !isFilled(command) {
		return onEmpty(fmt.Sprintf(`"%s" rendered empty, exiting.`, CommandCreate))
	}

	s.log(hclog.Info, "creating resource")
	lines, done, save := s.stateSetter()
	err = s.execute(lines, jsonCtx, command)
	save <- err == nil
	<-done
	if err != nil {
		return err
	}

	if err := s.ensureId(); err != nil {
		return err
	}
	s.log(hclog.Debug, "created resource", "id", s.getId())
	return nil
}

func resourceScriptedReadBase(s *Scripted) error {
	onEmpty := func(msg string) error {
		s.log(hclog.Debug, msg)
		s.d.Set("output", map[string]string{})
		return nil
	}
	defer s.logging.PushDefer("commands", "read")()
	if !isSet(s.pc.Commands.Templates.Read) {
		return onEmpty(fmt.Sprintf(`"%s" is empty, exiting.`, CommandRead))
	}
	command, jsonCtx, err := s.prefixedTemplate(&TemplateArg{CommandRead, s.pc.Commands.Templates.Read})
	if err != nil {
		return err
	}
	if !isFilled(command) {
		return onEmpty(fmt.Sprintf(`"%s" rendered empty, exiting.`, CommandRead))
	}
	env, err := s.Environment()
	if err != nil {
		if s.op == OperationRead {
			// Return immediately in read so Context.Refresh() passes
			return nil
		}
		return err
	}
	s.log(hclog.Info, "reading resource", "command", command)
	output, doneCh, saveCh := s.outputSetter()
	err = s.executeBase(output, env, jsonCtx, command)
	saveCh <- err == nil
	<-doneCh
	if err != nil {
		if s.pc.Commands.DeleteOnReadFailure {
			s.log(hclog.Info, "command returned error, marking resource deleted", "error", err, "output", output)
			s.clear()
			return nil
		}
		return err
	}
	return nil
}

func resourceScriptedUpdateBase(s *Scripted) error {
	defer s.logging.PushDefer("commands", "update")()
	command, jsonCtx, err := s.prefixedTemplate(
		&TemplateArg{"commands_modify_prefix", s.pc.Commands.Templates.ModifyPrefix},
		&TemplateArg{CommandUpdate, s.pc.Commands.Templates.Update},
	)
	if err != nil {
		return err
	}
	if !isFilled(command) {
		s.log(hclog.Warn, fmt.Sprintf(`"%s" rendered empty, exiting.`, CommandUpdate))
		if err := s.ensureId(); err != nil {
			return err
		}
		return nil
	}
	s.log(hclog.Info, "updating resource", "command", command)
	lines, done, save := s.stateSetter()
	err = s.execute(lines, jsonCtx, command)
	save <- err == nil
	<-done
	if err != nil {
		s.log(hclog.Warn, "update command returned error", "error", err)
		return err
	}
	if err := s.ensureId(); err != nil {
		return err
	}
	return nil
}

func resourceScriptedDeleteBase(s *Scripted) error {
	defer s.logging.PushDefer("commands", "delete")()
	onEmpty := func(msg string) error {
		s.log(hclog.Debug, msg)
		s.clear()
		return nil
	}
	if !isSet(s.pc.Commands.Templates.Delete) {
		return onEmpty(fmt.Sprintf(`"%s" is empty, exiting.`, CommandDelete))
	}
	s.addOld(true)
	defer s.removeOld()
	command, jsonCtx, err := s.prefixedTemplate(&TemplateArg{CommandDelete, s.pc.Commands.Templates.Delete})
	if err != nil {
		return err
	}
	if !isFilled(command) {
		return onEmpty(fmt.Sprintf(`"%s" rendered empty, exiting.`, CommandDelete))
	}
	s.log(hclog.Info, "deleting resource")
	_, err = s.executeString(jsonCtx, command)
	if err != nil {
		return err
	}

	s.clear()
	return nil
}
