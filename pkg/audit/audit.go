package audit

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/elastic/go-libaudit/v2"
	"github.com/elastic/go-libaudit/v2/auparse"
	"github.com/elastic/go-libaudit/v2/rule"
	"github.com/elastic/go-libaudit/v2/rule/flags"
	"github.com/pkg/errors"

	"github.com/juan-leon/fetter/pkg/cgroups"
	"github.com/juan-leon/fetter/pkg/log"
	"github.com/juan-leon/fetter/pkg/settings"
	"github.com/juan-leon/fetter/pkg/triggers"
)

const (
	auditLocked = 2
	cgPrefix    = "fetter_"
)

const (
	MODE_REUSE    string = "reuse"
	MODE_OVERRIDE string = "override"
	MODE_PRESERVE string = "preserve"
)

const (
	SYSCALL_READ    string = "read"
	SYSCALL_EXECUTE string = "execute"
	SYSCALL_WRITE   string = "write"
)

type SysCallListener struct {
	config     *settings.Settings
	client     *libaudit.AuditClient
	procMover  cgroups.ProcessMover
	procRunner triggers.ProcessRunner
}

func NewSysCallListener(config *settings.Settings, procMover cgroups.ProcessMover, procRunner triggers.ProcessRunner) *SysCallListener {
	if !assertAuditMode(config.Audit.Mode) {
		log.Logger.Errorf("unknown config for audit.mode: %s", config.Audit.Mode)
		return nil
	}

	client, err := libaudit.NewMulticastAuditClient(nil)
	if err != nil {
		log.Logger.Errorf("failed to create audit client euid=%v", os.Geteuid())
		return nil
	}
	return &SysCallListener{
		client:     client,
		config:     config,
		procMover:  procMover,
		procRunner: procRunner,
	}
}

func (scl *SysCallListener) Loop() {
	defer closeAuditClient(scl.client)
	scl.configure()
	log.Logger.Debugw("Forever snooping syscalls")
	scl.loop()
}

func (scl *SysCallListener) configure() {
	if scl.config.Audit.Mode != MODE_REUSE {
		scl.addRules()
	} else {
		log.Logger.Infof("Reusing existing audit rules")
	}
	scl.client.SetEnabled(true, libaudit.NoWait)
}

func closeAuditClient(client *libaudit.AuditClient) error {
	discard := func(bytes []byte) ([]syscall.NetlinkMessage, error) {
		return nil, nil
	}
	// Drain the netlink channel in parallel to Close() to prevent a deadlock.
	// Code copied from auditd module form auditbeat project
	go func() {
		for {
			_, err := client.Netlink.Receive(true, discard)
			switch err {
			case nil, syscall.EINTR:
			case syscall.EAGAIN:
				time.Sleep(50 * time.Millisecond)
			default:
				return
			}
		}
	}()
	return client.Close()
}

func (scl *SysCallListener) addRules() error {
	client, err := libaudit.NewAuditClient(nil)
	if err != nil {
		log.Logger.Errorf("failed to create audit client: %s", err)
		return err
	}
	defer closeAuditClient(client)

	status, err := client.GetStatus()
	if err != nil {
		log.Logger.Errorf("failed to get status from audit client: %s", err)
		return err
	}
	if scl.config.Audit.Mode != MODE_REUSE {
		if status.Enabled == auditLocked {
			log.Logger.Fatalf("Audit rules are locked :-(")
		}
	}

	if scl.config.Audit.Mode == MODE_OVERRIDE {
		n, err := client.DeleteRules()
		if err != nil {
			log.Logger.Errorf("Failed to delete existing rules: %s", err)
			return err
		}
		log.Logger.Infof("Deleted %d pre-existing audit rules.", n)
	}

	for name, r := range scl.config.Rules {
		scl.addRule(name, r, client)
	}
	return nil
}

func (scl *SysCallListener) addRule(name string, r settings.Rule, client *libaudit.AuditClient) {
	if err := validateRule(r); err != nil {
		log.Logger.Errorw("Failed to validate rule", "rule", r, "error", err.Error())
		return
	}
	for _, path := range r.Paths {
		asString := asAuditFmt(name, path, r.Action)
		parsedRule, err := flags.Parse(asString)
		if err != nil {
			log.Logger.Errorw("Failed to parse rule", "rule", r, "error", err.Error())
			return
		}

		ruleData, err := rule.Build(parsedRule)
		if err != nil {
			log.Logger.Errorw("Failed to build rule", "rule", r, "error", err.Error())
			return
		}

		err = client.AddRule([]byte(ruleData))
		if err != nil {
			log.Logger.Errorw("Failed to add rule", "rule", r, "error", err.Error())
			return
		}
		log.Logger.Debugw("Added path for rule", "path", path, "rule", r)
	}
}

func (scl *SysCallListener) loop() {
	for {
		auditMsg, err := scl.client.Receive(false)
		if err != nil {
			if errors.Cause(err) == syscall.EBADF {
				log.Logger.Warn("Audit client has been closed")
				break
			}
			log.Logger.Warn("Error listening kernel events: %s", err)
			continue
		}
		if auditMsg.Type != auparse.AUDIT_SYSCALL {
			// We are interested in SYSCALL events only (as those include
			// execution, read and write, and are triggered bebore process is
			// ended).
			continue
		}
		log.Logger.Debugw("Received syscall event", "raw-syscall", string(auditMsg.Data))

		msg, err := auparse.Parse(auditMsg.Type, string(auditMsg.Data))
		if err != nil {
			log.Logger.Errorw("Error parsing msg", "raw-syscall", string(auditMsg.Data))
			continue
		}
		scl.processMessage(msg)
	}
}

func (scl *SysCallListener) processMessage(msg *auparse.AuditMessage) {
	// For auparse, key is a tag and is not present in 'Data'
	tags, err := msg.Tags()
	if err != nil {
		log.Logger.Errorf("Could not parse tags from message: %s", err)
		return
	}
	for _, tagValue := range tags {
		if strings.HasPrefix(tagValue, cgPrefix) {
			data, err := msg.Data()
			if err != nil {
				log.Logger.Errorf("Could not extract data from message: %s", err)
				return
			}
			pid, err := strconv.Atoi(data["pid"])
			if err != nil {
				log.Logger.Fatalf("Got a non-numeric pid %s: %s", data["pid"], err)
			}
			scl.processMatch(pid, tagValue[len(cgPrefix):], &data)
			return
		}
	}
}

func (scl *SysCallListener) processMatch(pid int, rule string, data *map[string]string) {
	log.Logger.Infof("Match for rule %s in pid %d", rule, pid)
	if group := scl.config.GetGroup(rule); group != "" {
		scl.procMover.Move(pid, group)
	}
	if trigger := scl.config.GetTrigger(rule); trigger != "" {
		scl.procRunner.Run(trigger, data)
	}
}

func assertAuditMode(mode string) bool {
	switch mode {
	case
		MODE_REUSE, MODE_OVERRIDE, MODE_PRESERVE:
		return true
	default:
		return false
	}
}

func asAuditFmt(name, path, ruleAction string) string {
	action := "x"
	switch ruleAction {
	case SYSCALL_READ:
		action = "r"
	case SYSCALL_WRITE:
		action = "w"
	}
	return fmt.Sprintf("-w %s -p %s -k %s%s", path, action, cgPrefix, name)
}

func validateRule(r settings.Rule) error {
	if len(r.Paths) < 1 {
		return fmt.Errorf("path cannot be empty")
	}
	if r.Group == "" && r.Trigger == "" {
		return fmt.Errorf("both group and trigger cannot be empty")
	}
	if r.Action == "" {
		return fmt.Errorf("action cannot be empty")
	}
	switch r.Action {
	case
		SYSCALL_READ, SYSCALL_EXECUTE, SYSCALL_WRITE:
	default:
		return fmt.Errorf("unknown action %s for rule", r.Action)
	}
	return nil
}
