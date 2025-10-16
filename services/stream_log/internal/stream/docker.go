package stream

import (
	"bufio"
	"context"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"io"
	"log"
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
		Tail:       "100",
	}

	logReader, err := cli.ContainerLogs(ctx, containerName, options)
	if err != nil {
		return err
	}
	defer logReader.Close()

	pr, pw := io.Pipe()
	go func() {
		defer pw.Close()
		_, _ = stdcopy.StdCopy(pw, pw, logReader)
	}()

	scanner := bufio.NewScanner(pr)
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return nil
		default:
			text := strings.TrimSpace(scanner.Text())
			log.Println("docker proccessor log: ", text)
			if strings.HasPrefix(text, "ocserv[") {
				streamChan <- text
			}
			continue
		}
	}
	return scanner.Err()
}
