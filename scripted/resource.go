package scripted

import (
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/terraform/helper/schema"
)

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
			},
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
			"needs_update": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Helper indicating whether resource should be updated, ignore this.",
			},
			"dependencies_met": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Helper indicating whether resource dependencies are met, ignore this.",
			},
		},

		CustomizeDiff: func(diff *schema.ResourceDiff, i interface{}) error {
			diff.Clear("log_name")
			if met, ok := diff.GetOk("dependencies_met"); ok && !met.(bool) {
				for _, key := range diff.UpdatedKeys() {
					diff.Clear(key)
				}
				return nil
			}
			if diff.Get("needs_update").(bool) {
				diff.SetNewComputed("needs_update")
			} else {

				diff.Clear("needs_update")
			}
			return nil
		},
	}
	return ret
}

func resourceScriptedCreate(d *schema.ResourceData, meta interface{}) error {
	s, err := New(d, meta, Create, false)
	if err != nil {
		return err
	}
	if err := s.checkDependenciesMet(); err != nil {
		return err
	}
	if !s.dependenciesMet() {
		s.log(hclog.Warn, "dependencies not met, exiting")
		s.d.SetId("")
		return nil
	}
	return resourceScriptedCreateBase(s)
}

func resourceScriptedCreateBase(s *Scripted) error {
	defer s.logging.PopIf(s.logging.Push("create", true))
	if !isSet(s.pc.Commands.Templates.Create) {
		s.log(hclog.Debug, `"commands_create" is empty, exiting.`)
		if err := s.ensureId(); err != nil {
			return err
		}
		s.syncState()
		return resourceScriptedReadBase(s)
	}
	command, err := s.template(
		"commands_prefix_fromenv+commands_prefix+commands_modify_prefix+commands_create",
		s.joinCommands(s.pc.Commands.Templates.PrefixFromEnv, s.pc.Commands.Templates.Prefix, s.pc.Commands.Templates.ModifyPrefix, s.pc.Commands.Templates.Create))
	if err != nil {
		return err
	}
	s.log(hclog.Debug, "creating resource")
	stdout, err := s.execute(command)
	if err != nil {
		return err
	}

	if err := s.ensureId(); err != nil {
		return err
	}
	s.clearState()
	s.updateState(stdout)
	s.log(hclog.Debug, "created resource", "id", s.getId())
	return resourceScriptedReadBase(s)
}

func resourceScriptedRead(d *schema.ResourceData, meta interface{}) error {
	s, err := New(d, meta, Read, false)
	if err != nil {
		return err
	}
	return resourceScriptedReadBase(s)
}

func resourceScriptedReadBase(s *Scripted) error {
	if err := s.checkDependenciesMet(); err != nil {
		return err
	}
	if err := s.checkNeedsUpdate(); err != nil {
		return err
	}
	defer s.logging.PopIf(s.logging.Push("read", true))
	if !isSet(s.pc.Commands.Templates.Read) {
		s.log(hclog.Debug, `"commands_read" is not set, exiting.`)
		s.d.Set("output", map[string]string{})
		return nil
	}
	command, err := s.template(
		"commands_prefix_fromenv+commands_prefix+commands_read",
		s.joinCommands(s.pc.Commands.Templates.PrefixFromEnv, s.pc.Commands.Templates.Prefix, s.pc.Commands.Templates.Read))
	if err != nil {
		return err
	}
	s.log(hclog.Debug, "reading resource")
	env, err := s.Environment()
	if err != nil {
		if s.op == Read {
			// Return immediately in read so Context.Refresh() passes
			return nil
		}
		return err
	}
	stdout, err := s.executeEnv(env, command)
	if err != nil {
		if s.pc.Commands.DeleteOnReadFailure {
			s.log(hclog.Info, "command returned error, marking resource deleted", "error", err, "stdout", stdout)
			s.clear()
			return nil
		}
		return err
	}
	s.setOutput(stdout)
	return nil
}

func resourceScriptedUpdate(d *schema.ResourceData, meta interface{}) error {
	s, err := New(d, meta, Update, false)
	if err != nil {
		return err
	}
	shouldUpdate := isSet(s.pc.Commands.Templates.Update)

	if !shouldUpdate {
		if err := resourceScriptedDeleteBase(s); err != nil {
			return err
		}
	}

	if shouldUpdate {
		err = resourceScriptedUpdateBase(s)
		if err != nil {
			return err
		}
	} else {
		s.log(hclog.Debug, `"commands_update" is empty, skipping`)
	}

	if !shouldUpdate {
		if err := resourceScriptedCreateBase(s); err != nil {
			return err
		}
	}

	return resourceScriptedReadBase(s)
}

func resourceScriptedUpdateBase(s *Scripted) error {
	defer s.logging.PopIf(s.logging.Push("update", true))
	command, err := s.template(
		"commands_prefix_fromenv+commands_prefix+commands_modify_prefix+commands_update",
		s.joinCommands(s.pc.Commands.Templates.PrefixFromEnv, s.pc.Commands.Templates.Prefix, s.pc.Commands.Templates.ModifyPrefix, s.pc.Commands.Templates.Update))
	if err != nil {
		return err
	}
	s.log(hclog.Debug, "updating resource", "command", command)
	stdout, err := s.execute(command)
	if err != nil {
		s.log(hclog.Warn, "update command returned error", "error", err)
		return err
	}
	s.ensureId()
	s.updateState(stdout)
	return nil
}

func resourceScriptedExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	s, err := New(d, meta, Exists, false)
	if err != nil {
		return true, err
	}
	defer s.logging.PopIf(s.logging.Push("exists", true))
	if !isSet(s.pc.Commands.Templates.Exists) {
		s.log(hclog.Debug, `"commands_exists" is empty, exiting.`)
		return true, nil
	}
	command, err := s.template(
		"commands_prefix_fromenv+commands_prefix+commands_exists",
		s.joinCommands(s.pc.Commands.Templates.PrefixFromEnv, s.pc.Commands.Templates.Prefix, s.pc.Commands.Templates.Exists))
	if err != nil {
		return false, err
	}
	s.log(hclog.Debug, "resource exists")
	output, err := s.execute(command)
	if err != nil {
		s.log(hclog.Warn, "exists returned error", "error", err)
	}
	exists := err == nil && output != s.pc.Commands.ExistsExpectedOutput
	if !exists && s.pc.Commands.DeleteOnNotExists {
		s.clear()
	}
	return exists, err
}

func resourceScriptedDelete(d *schema.ResourceData, meta interface{}) error {
	s, err := New(d, meta, Delete, true)
	if err != nil {
		return err
	}
	return resourceScriptedDeleteBase(s)
}

func resourceScriptedDeleteBase(s *Scripted) error {
	defer s.logging.PopIf(s.logging.Push("delete", true))
	if !isSet(s.pc.Commands.Templates.Delete) {
		s.log(hclog.Debug, `"commands_delete" is empty, exiting.`)
		s.clear()
		return nil
	}
	s.addOld(true)
	defer s.removeOld()
	command, err := s.template(
		"commands_prefix_fromenv+commands_prefix+commands_delete",
		s.joinCommands(s.pc.Commands.Templates.PrefixFromEnv, s.pc.Commands.Templates.Prefix, s.pc.Commands.Templates.Delete))
	if err != nil {
		return err
	}
	s.log(hclog.Debug, "deleting resource")
	_, err = s.execute(command)
	if err != nil {
		return err
	}

	s.clear()
	return nil
}
