package settings

import (
	"context"
	"fmt"
	"os"

	"github.com/heetch/confita"
	"github.com/heetch/confita/backend/file"
)

// Load configuration into settings variable
func Load(path string) (settings *Settings, err error) {
	settings = &Settings{
		Name: "fetter",
		Mode: RunModeAudit,
		Logging: Logging{
			File:  "/tmp/fetter.log",
			Level: "info",
		},
		Audit: Audit{Mode: "override"},
	}
	if _, err = os.Stat(path); err != nil {
		return nil, err
	}
	loader := confita.NewLoader(file.NewBackend(path))
	err = loader.Load(context.Background(), settings)
	if err == nil {
		err = assertConfigOk(settings)
	}
	return
}

func assertConfigOk(settings *Settings) error {
	for name, rule := range settings.Rules {
		if rule.Trigger != "" {
			if rule.Trigger == "KILL" {
				continue
			}
			if _, ok := settings.Triggers[rule.Trigger]; !ok {
				return fmt.Errorf("missing trigger '%s' defined for rule '%s'", rule.Trigger, name)
			}
		}
		if rule.Group != "" {
			if _, ok := settings.Groups[rule.Group]; !ok {
				return fmt.Errorf("missing group '%s' defined for rule '%s'", rule.Trigger, name)
			}
		}
	}
	switch settings.Mode {
	case
		RunModeAudit, RunModeScanner:
		return nil
	default:
		return fmt.Errorf("run mode not supported: %s", settings.Mode)
	}
}
