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
			// "log_name": {
			// 	Type:        schema.TypeString,
			// 	Computed: true,
			// 	Description: "Resource name to display in log messages",
			// },
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
			"needs_delete": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Helper indicating whether resource should be deleted, ignore this.",
			},
			"needs_update": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Helper indicating whether resource should be updated, ignore this.",
			},
			"dependencies_met": {
				Type:        schema.TypeBool,
				Computed:    true,
				Optional:    true,
				Description: "Helper indicating whether resource dependencies are met, ignore this.",
			},
		},

		CustomizeDiff: resourceScriptedCustomizeDiff,
	}
	return ret
}

func resourceScriptedCustomizeDiff(diff *schema.ResourceDiff, i interface{}) error {
	diff.Clear("needs_delete")
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
}

func resourceScriptedCreate(d *schema.ResourceData, meta interface{}) error {
	s, err := New(d, meta, Create, false)
	if err != nil {
		return err
	}
	met, err := s.checkDependenciesMet()
	if err != nil {
		return err
	}
	if !met {
		s.log(hclog.Warn, "create dependencies not met, clearing state then exiting")
		s.clearState()
		return nil
	}
	if needsDelete, err := s.checkNeedsDelete(); err != nil || needsDelete {
		s.d.SetId("")
		return err
	}
	return resourceScriptedCreateBase(s)
}

func resourceScriptedRead(d *schema.ResourceData, meta interface{}) error {
	s, err := New(d, meta, Read, false)
	if err != nil {
		return err
	}
	defer s.runningMessages()()
	met, err := s.checkDependenciesMet()
	if err != nil {
		return err
	}
	if !met {
		return nil
	}

	return resourceScriptedReadBase(s)
}

func resourceScriptedUpdate(d *schema.ResourceData, meta interface{}) error {
	s, err := New(d, meta, Update, false)
	if err != nil {
		return err
	}
	defer s.runningMessages()()
	if s.needsDelete() {
		s.log(hclog.Debug, `needsDelete == true, exiting update.`)
		s.clear()
		return nil
	}
	met, err := s.checkDependenciesMet()
	if err != nil {
		return err
	}
	if !met {
		return nil
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

func resourceScriptedExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	s, err := New(d, meta, Exists, false)
	if err != nil {
		return true, err
	}
	defer s.runningMessages()()
	met, err := s.checkDependenciesMet()
	if err != nil {
		return true, err
	}
	if !met {
		return true, nil
	}
	defer s.logging.PopIf(s.logging.Push("exists", true))
	if needsDelete, err := s.checkNeedsDelete(); err != nil || needsDelete {
		s.d.SetId("")
		return false, err
	}
	if !isSet(s.pc.Commands.Templates.Exists) {
		s.log(hclog.Debug, `"commands_exists" is empty, exiting.`)
		return true, nil
	}
	command, err := s.prefixedTemplate(&TemplateArg{"commands_exists", s.pc.Commands.Templates.Exists})
	if err != nil {
		return false, err
	}
	if !isFilled(command) {
		return true, nil
	}
	s.log(hclog.Debug, "resource exists")
	output, err := s.executeString(command)
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
	defer s.runningMessages()()
	met, err := s.checkDependenciesMet()
	if err != nil {
		return err
	}
	if !met {
		return nil
	}
	return resourceScriptedDeleteBase(s)
}

func resourceScriptedCreateBase(s *Scripted) error {
	defer s.logging.PopIf(s.logging.Push("create", true))
	onEmpty := func(msg string) error {
		if isSet(s.pc.Commands.Templates.Update) {
			s.log(hclog.Debug, `"commands_create" is empty, running "commands_update".`)
			return resourceScriptedUpdateBase(s)
		}
		s.log(hclog.Debug, msg)
		if err := s.ensureId(); err != nil {
			return err
		}
		s.clearState()
		s.syncState()
		return resourceScriptedReadBase(s)
	}

	if !isSet(s.pc.Commands.Templates.Create) {
		return onEmpty(`"commands_create" is empty, exiting.`)
	}
	command, err := s.prefixedTemplate(
		&TemplateArg{"commands_modify_prefix", s.pc.Commands.Templates.ModifyPrefix},
		&TemplateArg{"commands_create", s.pc.Commands.Templates.Create},
	)
	if err != nil {
		return err
	}

	if !isFilled(command) {
		return onEmpty(`"commands_create" rendered empty, exiting.`)
	}

	s.log(hclog.Debug, "creating resource")
	s.clearState()
	lines, done := s.stateUpdater()
	err = s.execute(lines, command)
	<-done
	if err != nil {
		return err
	}

	if err := s.ensureId(); err != nil {
		return err
	}
	s.log(hclog.Debug, "created resource", "id", s.getId())
	return resourceScriptedReadBase(s)
}

func resourceScriptedReadBase(s *Scripted) error {
	if err := s.checkNeedsUpdate(); err != nil {
		return err
	}
	onEmpty := func(msg string) error {
		s.log(hclog.Debug, msg)
		s.d.Set("output", map[string]string{})
		return nil
	}
	defer s.logging.PopIf(s.logging.Push("read", true))
	if !isSet(s.pc.Commands.Templates.Read) {
		return onEmpty(`"commands_read" is not msg, exiting.`)
	}
	command, err := s.prefixedTemplate(&TemplateArg{"commands_read", s.pc.Commands.Templates.Read})
	if err != nil {
		return err
	}
	if !isFilled(command) {
		return onEmpty(`"commands_read" rendered empty, exiting.`)
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
	s.log(hclog.Trace, "executing read", "command", command)
	output, doneCh := s.outputSetter()
	err = s.executeBase(output, env, command)
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
	defer s.logging.PopIf(s.logging.Push("update", true))
	command, err := s.prefixedTemplate(
		&TemplateArg{"commands_modify_prefix", s.pc.Commands.Templates.ModifyPrefix},
		&TemplateArg{"commands_update", s.pc.Commands.Templates.Update},
	)
	if err != nil {
		return err
	}
	if !isFilled(command) {
		s.log(hclog.Warn, `"commands_update" rendered empty, exiting.`)
		if err := s.ensureId(); err != nil {
			return err
		}
		return nil
	}
	s.log(hclog.Debug, "updating resource", "command", command)
	lines, done := s.stateUpdater()
	err = s.execute(lines, command)
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
	defer s.logging.PopIf(s.logging.Push("delete", true))
	onEmpty := func(msg string) error {
		s.log(hclog.Debug, msg)
		s.clear()
		return nil
	}
	if !isSet(s.pc.Commands.Templates.Delete) {
		return onEmpty(`"commands_delete" is empty, exiting.`)
	}
	s.addOld(true)
	defer s.removeOld()
	command, err := s.prefixedTemplate(&TemplateArg{"commands_delete", s.pc.Commands.Templates.Delete})
	if err != nil {
		return err
	}
	if !isFilled(command) {
		return onEmpty(`"commands_delete" rendered empty, exiting.`)
	}
	s.log(hclog.Debug, "deleting resource")
	_, err = s.executeString(command)
	if err != nil {
		return err
	}

	s.clear()
	return nil
}
