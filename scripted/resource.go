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
			"templates_propagate_errors": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Should templates propagate errors?",
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

func resourceScriptedCreate(d *schema.ResourceData, meta interface{}) error {
	s, err := NewState(d, meta, "create", false)
	if err != nil {
		return err
	}
	return resourceScriptedCreateBase(s)
}

func resourceScriptedCreateBase(s *State) error {
	command, err := s.renderTemplate(
		"command_prefix+create_command",
		s.prepareCommands(s.pc.CommandPrefix, s.pc.CreateCommand))
	if err != nil {
		return err
	}
	s.log(hclog.Debug, "creating resource")
	_, err = s.runCommand(command)
	if err != nil {
		return err
	}

	s.ensureId()
	s.log(hclog.Debug, "created resource", "id", s.getId())
	return resourceScriptedReadBase(s)
}

func resourceScriptedRead(d *schema.ResourceData, meta interface{}) error {
	s, err := NewState(d, meta, "read", false)
	if err != nil {
		return err
	}
	return resourceScriptedReadBase(s)
}

func resourceScriptedReadBase(s *State) error {
	command, err := s.renderTemplate(
		"command_prefix+read_command",
		s.prepareCommands(s.pc.CommandPrefix, s.pc.ReadCommand))
	if err != nil {
		return err
	}
	s.log(hclog.Debug, "reading resource")
	stdout, err := s.runCommand(command)
	if err != nil {
		s.log(hclog.Info, "command returned error, marking resource deleted", "error", err, "stdout", stdout)
		if s.pc.DeleteOnReadFailure {
			s.d.SetId("")
			return nil
		}
		return err
	}
	var outputs map[string]string

	switch s.pc.ReadFormat {
	case "base64":
		outputs = s.getOutputsBase64(stdout, s.pc.ReadLinePrefix)
	default:
		fallthrough
	case "raw":
		outputs = s.getOutputsText(stdout, s.pc.ReadLinePrefix)
	}
	s.log(hclog.Debug, "Setting outputs", "output", outputs)
	s.d.Set("output", outputs)

	return nil
}

func resourceScriptedUpdate(d *schema.ResourceData, meta interface{}) error {
	s, err := NewState(d, meta, "update", false)
	if err != nil {
		return err
	}

	if s.pc.DeleteBeforeUpdate {
		if err := resourceScriptedDeleteBase(s); err != nil {
			return err
		}
	}

	if s.pc.CreateBeforeUpdate {
		if err := resourceScriptedCreateBase(s); err != nil {
			return err
		}
	}

	if s.pc.UpdateCommand != "" {
		s.setOld(true)
		deleteCommand, _ := s.renderTemplate(
			"command_prefix+delete_command",
			s.wrapCommands(s.pc.CommandPrefix, s.pc.DeleteCommand))
		s.setOld(false)
		createCommand, _ := s.renderTemplate(
			"command_prefix+create_command",
			s.wrapCommands(s.pc.CommandPrefix, s.pc.CreateCommand))
		command, err := s.renderTemplateExtraCtx(
			"command_prefix+update_command",
			s.prepareCommands(s.pc.CommandPrefix, s.pc.UpdateCommand),
			map[string]string{
				"delete_command": deleteCommand,
				"create_command": createCommand,
			})
		if err != nil {
			return err
		}
		s.log(hclog.Debug, "updating resource", "command", command)
		_, err = s.runCommand(command)
		if err != nil {
			s.log(hclog.Warn, "update command returned error", "error", err)
			return nil
		}
		s.ensureId()
	}

	if s.pc.CreateAfterUpdate {
		if err := resourceScriptedCreateBase(s); err != nil {
			return err
		}
	}

	return resourceScriptedReadBase(s)
}

func resourceScriptedExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	s, err := NewState(d, meta, "exists", false)
	if err != nil {
		return false, err
	}
	if s.pc.ExistsCommand == "" {
		return true, nil
	}
	command, err := s.renderTemplate(
		"command_prefix+exists_command",
		s.prepareCommands(s.pc.CommandPrefix, s.pc.ExistsCommand))
	if err != nil {
		return false, err
	}
	s.log(hclog.Debug, "resource exists")
	_, err = s.runCommand(command)
	if err != nil {
		s.log(hclog.Warn, "command returned error", "error", err)
	}
	exists := getExitStatus(err) == s.pc.ExistsExpectedStatus
	if s.pc.ExistsExpectedStatus == 0 {
		exists = err == nil
	}
	if !exists && s.pc.DeleteOnNotExists {
		s.d.SetId("")
	}
	return exists, nil
}

func resourceScriptedDelete(d *schema.ResourceData, meta interface{}) error {
	s, err := NewState(d, meta, "delete", true)
	if err != nil {
		return err
	}
	return resourceScriptedDeleteBase(s)
}

func resourceScriptedDeleteBase(s *State) error {
	wasOld := s.isOld()
	s.setOld(true)
	command, err := s.renderTemplate(
		"command_prefix+delete_command",
		s.prepareCommands(s.pc.CommandPrefix, s.pc.DeleteCommand))
	if err != nil {
		return err
	}
	s.log(hclog.Debug, "reading resource")
	_, err = s.runCommand(command)
	if err != nil {
		return err
	}

	s.d.SetId("")
	s.setOld(wasOld)
	return nil
}
