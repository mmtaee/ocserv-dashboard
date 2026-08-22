package readers

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

type journalEntry struct {
	Message string `json:"MESSAGE"`
	PID     string `json:"_PID"`
}

func SystemdStreamLogs(ctx context.Context, serviceName string, streamChan chan<- string) error {
	cmd := exec.CommandContext(ctx, "journalctl", "-n", "0", "-fu", serviceName, "--output=json")

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	if err = cmd.Start(); err != nil {
		return err
	}
	defer cmd.Wait()

	scanner := bufio.NewScanner(stdout)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	for scanner.Scan() {
		var entry journalEntry
		if err := json.Unmarshal(scanner.Bytes(), &entry); err != nil {
			continue
		}

		message := strings.TrimSpace(entry.Message)
		if message == "" {
			continue
		}

		if entry.PID != "" && !strings.HasPrefix(message, "ocserv[") {
			message = fmt.Sprintf("ocserv[%s]: %s", entry.PID, message)
		}

		select {
		case <-ctx.Done():
			return nil
		case streamChan <- message:
		}
	}

	return scanner.Err()
}
