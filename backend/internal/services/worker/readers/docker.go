package readers

import (
	"bufio"
	"context"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"io"
	"strings"
)

func DockerStreamLogs(ctx context.Context, containerName string, streamChan chan<- string) error {
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
		if !strings.HasPrefix(text, "ocserv[") {
			continue
		}
		select {
		case <-ctx.Done():
			return nil
		case streamChan <- text:
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
