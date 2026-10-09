package readers

import (
	"bufio"
	"context"
	"io"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/mmtaee/ocserv-dashboard/backend/internal/platform/logging"
)

func DockerStreamLogs(ctx context.Context, containerName string, streamChan chan<- logger.StreamEntry) error {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}
	defer cli.Close()

	options := container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
		Tail:       "0",
		Timestamps: true,
	}

	logReader, err := cli.ContainerLogs(ctx, containerName, options)
	if err != nil {
		return err
	}
	defer logReader.Close()

	pr, pw := io.Pipe()
	copyDone := make(chan error, 1)
	go func() {
		defer pw.Close()
		_, copyErr := stdcopy.StdCopy(pw, pw, logReader)
		copyDone <- copyErr
	}()

	scanner := bufio.NewScanner(pr)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		text := strings.TrimSpace(scanner.Text())
		timestamp := time.Now().UTC()
		parts := strings.SplitN(text, " ", 2)
		if len(parts) == 2 {
			if parsed, err := time.Parse(time.RFC3339Nano, parts[0]); err == nil {
				timestamp = parsed.UTC()
				text = strings.TrimSpace(parts[1])
			}
		}
		if !strings.HasPrefix(text, "ocserv[") {
			continue
		}
		select {
		case <-ctx.Done():
			return nil
		case streamChan <- logger.StreamEntry{Message: text, Timestamp: timestamp}:
		}
	}
	scanErr := scanner.Err()
	_ = pr.Close()
	copyErr := <-copyDone
	if ctx.Err() != nil {
		return nil
	}
	if scanErr != nil {
		return scanErr
	}
	return copyErr
}
