package system

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var ErrUpdateRunning = errors.New("an update is already running")

const updateScript = `set -o pipefail
umask 077
cd "$1"
finish() {
    code=$?
    trap - EXIT
    if ((code == 0)); then state=succeeded; else state=failed; fi
    printf '%s\n' "$state" >"$2/.update-state.tmp"
    mv "$2/.update-state.tmp" "$2/.update-state"
    rmdir "$2/.update-lock" 2>/dev/null || true
    exit "$code"
}
trap finish EXIT
bash ./setup.sh --update --yes 2>&1 | tee "$2/.update.log"
exit "${PIPESTATUS[0]}"`

func (u *Usecase) StartUpdate(ctx context.Context) (*UpdateStatus, error) {
	release, err := u.Release(ctx)
	if err != nil {
		return nil, err
	}
	if !release.Available {
		return nil, errors.New("dashboard is already up to date")
	}
	lock := filepath.Join(u.installDir, ".update-lock")
	if err := os.Mkdir(lock, 0700); err != nil {
		if errors.Is(err, os.ErrExist) {
			return nil, ErrUpdateRunning
		}
		return nil, fmt.Errorf("could not lock dashboard update: %w", err)
	}
	fail := func(err error) (*UpdateStatus, error) {
		_ = os.Remove(lock)
		_ = u.writeUpdateFile(".update-state", "failed")
		return nil, err
	}
	if err := u.writeUpdateFile(".update-target", release.Latest); err != nil {
		return fail(err)
	}
	if err := u.writeUpdateFile(".update-state", "running"); err != nil {
		return fail(err)
	}
	_ = os.Remove(filepath.Join(u.installDir, ".update.log"))
	starter := u.updateStarter
	if starter == nil {
		starter = u.startUpdateJob
	}
	if err := starter(release.Latest); err != nil {
		return fail(fmt.Errorf("could not start dashboard update: %w", err))
	}
	return u.UpdateStatus(), nil
}

func (u *Usecase) UpdateStatus() *UpdateStatus {
	state := u.readUpdateFile(".update-state")
	if state != "running" && state != "succeeded" && state != "failed" {
		state = "idle"
	}
	return &UpdateStatus{
		State:   state,
		Target:  u.readUpdateFile(".update-target"),
		Message: lastUpdateLogLine(filepath.Join(u.installDir, ".update.log")),
	}
}

func (u *Usecase) startUpdateJob(_ string) error {
	if u.dockerMode {
		image := "ocserv-dashboard:master"
		if u.agentNode {
			image = "ocserv-dashboard:agent"
		}
		return exec.Command("docker", "run", "--detach", "--rm", "--name", "ocserv-dashboard-update",
			"--entrypoint", "/bin/bash", "--device", "/dev/net/tun", "--volume", "/var/run/docker.sock:/var/run/docker.sock",
			"--volume", u.installDir+":"+u.installDir, "--workdir", u.sourceDir, image, "-c", updateScript, "updater", u.sourceDir, u.installDir).Run()
	}
	return exec.Command("systemd-run", "--collect", "--unit=ocserv-dashboard-update", "/bin/bash", "-c", updateScript, "updater", u.sourceDir, u.installDir).Run()
}

func (u *Usecase) writeUpdateFile(name, value string) error {
	temporary := filepath.Join(u.installDir, name+".tmp")
	if err := os.WriteFile(temporary, []byte(value+"\n"), 0600); err != nil {
		return err
	}
	return os.Rename(temporary, filepath.Join(u.installDir, name))
}

func (u *Usecase) readUpdateFile(name string) string {
	value, err := os.ReadFile(filepath.Join(u.installDir, name))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(value))
}

func lastUpdateLogLine(path string) string {
	file, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer file.Close()
	if info, statErr := file.Stat(); statErr == nil && info.Size() > 64*1024 {
		_, _ = file.Seek(-64*1024, io.SeekEnd)
	}
	line := ""
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if value := strings.TrimSpace(scanner.Text()); value != "" {
			lower := strings.ToLower(value)
			if !strings.Contains(lower, "password") && !strings.Contains(lower, "secret") && !strings.Contains(lower, "token") {
				line = value
			}
		}
	}
	return line
}
