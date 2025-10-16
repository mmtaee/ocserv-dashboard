package stream

import (
	"bufio"
	"context"
	"os/exec"
)

func SystemdStreamLogs(ctx context.Context, serviceName string, streamChan chan<- string) error {
	cmd := exec.Command("journalctl", "-n", "100", "-fu", serviceName, "--output=cat")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	if err = cmd.Start(); err != nil {
		return err
	}
	defer cmd.Wait()

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return nil
		case streamChan <- scanner.Text():
		}
	}
	return scanner.Err()
}
