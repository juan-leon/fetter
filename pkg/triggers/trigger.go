package triggers

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"strconv"
	"syscall"

	"github.com/juan-leon/fetter/pkg/log"
	"github.com/juan-leon/fetter/pkg/settings"
)

type TriggerRunner struct {
	triggers map[string]settings.Trigger
}

func NewTriggerRunner(config *settings.Settings) *TriggerRunner {
	return &TriggerRunner{
		triggers: config.Triggers,
	}
}

func (tr *TriggerRunner) Run(name string, data *map[string]string) error {
	if trigger, ok := tr.triggers[name]; ok {
		go func() { run(&trigger, name, data) }()
		return nil
	}
	return fmt.Errorf("could not find trigger named: %s", name)
}

func run(trigger *settings.Trigger, name string, data *map[string]string) error {
	cmd := exec.Command(trigger.Run, trigger.Args...)
	user, err := getUser(trigger.User)
	if err != nil {
		log.Logger.Errorf("Could not find user for trigger %s: %s", name, err)
		return err
	}
	if sysProcAttr, err := getSysProcAttr(user); err == nil {
		if os.Geteuid() == 0 {
			cmd.SysProcAttr = sysProcAttr
		}
	}
	cmd.Env = getEnv(user, data)
	log.Logger.Infof("Running %s for trigger '%s'", trigger.Run, name)
	out, err := cmd.CombinedOutput()
	if err != nil {
		s := string(out)
		if len(s) > 128 {
			s = s[:128]
		}
		log.Logger.Errorw("Trigger execution failed", "name", name, "error", err, "out", s)
		return err
	}
	log.Logger.Infof("Trigger '%s' ended with no error", trigger.Run)
	return nil
}

func getSysProcAttr(user *user.User) (*syscall.SysProcAttr, error) {
	sysProcAttr := &syscall.SysProcAttr{}
	uid, err := strconv.Atoi(user.Uid)
	if err != nil {
		log.Logger.Errorf("Could not find uid for user '%s': %s", user.Username, err)
		return sysProcAttr, err
	}
	sysProcAttr.Credential = &syscall.Credential{Uid: uint32(uid)}
	return sysProcAttr, nil
}

func getUser(userName string) (*user.User, error) {
	if userName == "" {
		userName = "nobody"
	}
	return user.Lookup(userName)
}

func getEnv(user *user.User, data *map[string]string) []string {
	result := os.Environ()
	result = append(result, fmt.Sprintf("HOME=%s", user.HomeDir))
	if data != nil {
		for k, v := range *data {
			result = append(result, fmt.Sprintf("%s=%s", k, v))
		}
	}
	return result
}
