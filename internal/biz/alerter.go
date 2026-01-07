package biz

import (
	"context"
	"os/exec"
	"strings"

	"github.com/go-kratos/kratos/v2/log"
)

// defaultAlerter is the default implementation of Alerter
type defaultAlerter struct {
	log *log.Helper
}

// NewAlerter creates a new alert component
func NewAlerter(logger log.Logger) Alerter {
	return &defaultAlerter{
		log: log.NewHelper(logger),
	}
}

// Alert sends an alert message
func (a *defaultAlerter) Alert(ctx context.Context, msg string) error {
	a.log.Errorf("ALERT TRIGGERED: %s", msg)

	// TODO: Integrate with DingTalk, Lark, Email, etc.
	// Or write to Prometheus metrics
	// Demo: execute a specific script if it exists in the environment
	// a.executeCommand("echo", msg)

	return nil
}

// executeCommand executes an external command (used for extending custom alert actions)
func (a *defaultAlerter) executeCommand(name string, args ...string) {
	cmd := exec.Command(name, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		a.log.Errorf("Failed to execute command: %v, output: %s", err, string(output))
		return
	}
	a.log.Infof("Command executed successfully: %s", strings.TrimSpace(string(output)))
}
