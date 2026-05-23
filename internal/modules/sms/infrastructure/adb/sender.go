package adb

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"

	"project/internal/app/logger"
	"project/internal/modules/sms/domain"
)

var adbLog = logger.New("sms-adb")

type Config struct {
	ADBPath string
	Device  string
}

type Sender struct {
	cfg Config
}

func NewSender(cfg Config) *Sender {
	adbPath := strings.TrimSpace(cfg.ADBPath)
	if adbPath == "" {
		adbPath = "adb"
	}

	return &Sender{
		cfg: Config{
			ADBPath: adbPath,
			Device:  strings.TrimSpace(cfg.Device),
		},
	}
}

func (s *Sender) Send(ctx context.Context, message domain.Message) error {
	if strings.TrimSpace(message.Address) == "" || strings.TrimSpace(message.Body) == "" {
		return domain.ErrInvalidMessage
	}

	if err := s.run(ctx, "shell", "appops", "set", "com.android.shell", "WRITE_SMS", "allow"); err != nil {
		return fmt.Errorf("grant write sms permission: %w", err)
	}

	body := escapeAndroidShellDoubleQuoted(message.Body)
	script := fmt.Sprintf(
		`NOW=$(date +%%s); content insert --uri content://sms/inbox --bind address:s:%s --bind body:s:"%s" --bind read:i:0 --bind seen:i:0 --bind type:i:1 --bind date:l:${NOW}000`,
		message.Address,
		body,
	)
	if err := s.run(ctx, "shell", script); err != nil {
		return fmt.Errorf("insert sms: %w", err)
	}

	verifyOutput, err := s.runOutput(ctx, "shell",
		"content", "query",
		"--uri", "content://sms/inbox",
		"--projection", "address,body,date",
		"--where", fmt.Sprintf("address='%s'", message.Address),
		"--sort", "date DESC",
	)
	if err != nil {
		adbLog.Warnf("verify inserted sms failed: err=%v", err)
	} else {
		adbLog.Infof("sms inbox verify: %s", trimOutputLines(verifyOutput, 3))
	}

	adbLog.Successf("sms sent via adb: address=%s body=%q", message.Address, message.Body)
	return nil
}

func (s *Sender) run(ctx context.Context, args ...string) error {
	_, err := s.runOutput(ctx, args...)
	return err
}

func (s *Sender) runOutput(ctx context.Context, args ...string) (string, error) {
	cmdArgs := s.adbArgs(args...)
	cmd := exec.CommandContext(ctx, s.cfg.ADBPath, cmdArgs...)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("%w: %s", err, strings.TrimSpace(stderr.String()))
	}

	output := strings.TrimSpace(stdout.String())
	if stderrText := strings.TrimSpace(stderr.String()); stderrText != "" {
		if output == "" {
			output = stderrText
		} else {
			output += "\n" + stderrText
		}
	}

	return output, nil
}

func (s *Sender) adbArgs(args ...string) []string {
	result := make([]string, 0, len(args)+2)
	if s.cfg.Device != "" {
		result = append(result, "-s", s.cfg.Device)
	}

	return append(result, args...)
}

func escapeAndroidShellDoubleQuoted(value string) string {
	value = strings.ReplaceAll(value, `\`, `\\`)
	value = strings.ReplaceAll(value, `"`, `\"`)
	value = strings.ReplaceAll(value, `$`, `\$`)
	value = strings.ReplaceAll(value, "`", "\\`")

	return value
}

func trimOutputLines(output string, limit int) string {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) <= limit {
		return strings.Join(lines, " | ")
	}

	return strings.Join(lines[:limit], " | ")
}
