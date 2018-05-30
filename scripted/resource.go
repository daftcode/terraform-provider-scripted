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
				// Hack so it doesn't ever force updates
				ForceNew: true,
				StateFunc: func(v interface{}) string {
					return ""
				},
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
			"state": {
				Type:        schema.TypeMap,
				Computed:    true,
				Description: "Output from create/update commands",
			},
		},
	}
	return ret
}

func resourceScriptedCreate(d *schema.ResourceData, meta interface{}) error {
	s, err := NewState(d, meta, Create, false)
	if err != nil {
		return err
	}
	return resourceScriptedCreateBase(s, true)
}

func resourceScriptedCreateBase(s *Scripted, newState bool) error {
	defer s.loggers.PopIf(s.loggers.Push("create", true))
	if s.pc.CreateCommand == "" {
		s.log(hclog.Debug, `"create_command" is empty, exiting.`)
		if err := s.ensureId(); err != nil {
			return err
		}
		s.syncState()
		return resourceScriptedReadBase(s)
	}
	command, err := s.template(
		"command_prefix+create_command",
		s.prepareCommands(s.pc.CommandPrefix, s.pc.CreateCommand))
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
	if newState {
		s.clearState()
	}
	s.updateState(stdout)
	s.log(hclog.Debug, "created resource", "id", s.getId())
	return resourceScriptedReadBase(s)
}

func resourceScriptedRead(d *schema.ResourceData, meta interface{}) error {
	s, err := NewState(d, meta, Read, false)
	if err != nil {
		return err
	}
	return resourceScriptedReadBase(s)
}

func resourceScriptedReadBase(s *Scripted) error {
	defer s.loggers.PopIf(s.loggers.Push("read", true))
	if s.pc.ReadCommand == "" {
		s.log(hclog.Debug, `"read_command" is empty, exiting.`)
		s.d.Set("output", map[string]string{})
		return nil
	}
	command, err := s.template(
		"command_prefix+read_command",
		s.prepareCommands(s.pc.CommandPrefix, s.pc.ReadCommand))
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
		s.log(hclog.Info, "command returned error, marking resource deleted", "error", err, "stdout", stdout)
		if s.pc.DeleteOnReadFailure {
			s.clear()
			return nil
		}
		return err
	}
	s.setOutput(stdout)
	return nil
}

func resourceScriptedUpdate(d *schema.ResourceData, meta interface{}) error {
	s, err := NewState(d, meta, Update, false)
	if err != nil {
		return err
	}

	if s.pc.DeleteBeforeUpdate {
		if err := resourceScriptedDeleteBase(s); err != nil {
			return err
		}
	}

	if s.pc.CreateBeforeUpdate {
		if err := resourceScriptedCreateBase(s, true); err != nil {
			return err
		}
	}

	shouldUpdate := s.pc.UpdateCommand != s.pc.EmptyString
	if shouldUpdate {
		err = resourceScriptedUpdateBase(s)
		if err != nil {
			return err
		}
	} else {
		s.log(hclog.Debug, `"update_command" is empty, skipping`)
	}

	if s.pc.CreateAfterUpdate {
		if err := resourceScriptedCreateBase(s, !shouldUpdate); err != nil {
			return err
		}
	}

	return resourceScriptedReadBase(s)
}

func resourceScriptedUpdateBase(s *Scripted) error {
	defer s.loggers.PopIf(s.loggers.Push("update", true))
	command, err := s.template(
		"command_prefix+update_command",
		s.prepareCommands(s.pc.CommandPrefix, s.pc.UpdateCommand))
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
	s, err := NewState(d, meta, Exists, false)
	if err != nil {
		return false, err
	}
	if s.pc.ExistsCommand == "" {
		s.log(hclog.Debug, `"exists_command" is empty, exiting.`)
		return true, nil
	}
	command, err := s.template(
		"command_prefix+exists_command",
		s.prepareCommands(s.pc.CommandPrefix, s.pc.ExistsCommand))
	if err != nil {
		return false, err
	}
	s.log(hclog.Debug, "resource exists")
	_, err = s.execute(command)
	if err != nil {
		s.log(hclog.Warn, "exists returned error", "error", err)
	}
	exists := getExitStatus(err) == s.pc.ExistsExpectedStatus
	if s.pc.ExistsExpectedStatus == 0 {
		exists = err == nil
	}
	if !exists && s.pc.DeleteOnNotExists {
		s.clear()
	}
	return exists, nil
}

func resourceScriptedDelete(d *schema.ResourceData, meta interface{}) error {
	s, err := NewState(d, meta, Delete, true)
	if err != nil {
		return err
	}
	return resourceScriptedDeleteBase(s)
}

func resourceScriptedDeleteBase(s *Scripted) error {
	defer s.loggers.PopIf(s.loggers.Push("update", true))
	if s.pc.DeleteCommand == "" {
		s.log(hclog.Debug, `"delete_command" is empty, exiting.`)
		s.clear()
		return nil
	}
	s.addOld(true)
	defer s.removeOld()
	command, err := s.template(
		"command_prefix+delete_command",
		s.prepareCommands(s.pc.CommandPrefix, s.pc.DeleteCommand))
	if err != nil {
		return err
	}
	s.log(hclog.Debug, "reading resource")
	_, err = s.execute(command)
	if err != nil {
		return err
	}

	s.clear()
	return nil
}
