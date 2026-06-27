package user

import (
	"bufio"
	"bytes"
	"context"
	"github.com/mmtaee/ocserv-dashboard/core/models"
	"github.com/mmtaee/ocserv-dashboard/core/pkg/utils"
	"os"
	"os/exec"
	"strings"
)

type OcservUser struct{}

type OcservUserManagement interface {
	Create(username, group, password string, config *models.OcservUserConfig) error
	Lock(username string) (string, error)
	UnLock(username string) (string, error)
	Delete(username string) (string, error)
}

type OcservUserConfigManagement interface {
	SyncConfig(username, group string, config *models.OcservUserConfig) error
	CreateConfig(username string, config *models.OcservUserConfig) error
	DeleteConfig(username string) error
}
type OcservUserPasswords interface {
	Ocpasswd(ctx context.Context) ([]Ocpasswd, int, error)
}

type OcservUserCertificateManagement interface {
	CreateCertificate(username, password string) error
	RevokeCertificate(username string) error
	SuspendCertificate(username string) error
	UnsuspendCertificate(username string) error
	CertificateStatus(username string) CertificateStatus
	CertificatePath(username string) (string, error)
	CertificateBackup(username string) (*models.OcservUserCertificateBackup, error)
	RestoreCertificateBackup(username string, cert *models.OcservUserCertificateBackup) error
}

type OcservUserInterface interface {
	OcservUserManagement
	OcservUserConfigManagement
	OcservUserPasswords
	OcservUserCertificateManagement
}

func NewOcservUser() *OcservUser {
	return &OcservUser{}
}

// Create creates a new ocserv user with the given username, group, and password.
// It runs the ocpasswd command to register the user. If a config is provided,
// a per-user configuration file is also written into ocserv.ConfigUserBaseDir
// with permission 0640. Returns an error if user creation fails.
func (u *OcservUser) Create(group, username, password string, config *models.OcservUserConfig) error {
	args := []string{"-c", utils.OcpasswdPath, username}
	if group != "" && group != "defaults" {
		args = append([]string{"-g", group}, args...)
	}
	cmd := exec.Command(utils.OcpasswdExec, args...)

	cmd.Stdin = bytes.NewBufferString(password + "\n" + password + "\n")
	_, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}

	if err = u.SyncConfig(username, group, config); err != nil {
		return err
	}
	return nil
}

// Lock disables a user account by running ocpasswd with the -l flag.
// Returns the command output or an error.
func (u *OcservUser) Lock(username string) (string, error) {
	output, err := utils.RunOcpasswd("-l", "-c", utils.OcpasswdPath, username)
	if err != nil {
		return "", err
	}

	if err = u.SuspendCertificate(username); err != nil {
		return output, err
	}

	return output, nil
}

// UnLock re-enables a previously locked user account by running ocpasswd
// with the -u flag. Returns the command output or an error.
func (u *OcservUser) UnLock(username string) (string, error) {
	output, err := utils.RunOcpasswd("-u", "-c", utils.OcpasswdPath, username)
	if err != nil {
		return "", err
	}

	if err = u.UnsuspendCertificate(username); err != nil {
		return output, err
	}

	return output, nil
}

// Delete removes a user account from ocserv by running ocpasswd with the -d flag.
// Returns the command output or an error.
func (u *OcservUser) Delete(username string) (string, error) {
	if err := u.RevokeCertificate(username); err != nil {
		return "", err
	}
	output, err := utils.RunOcpasswd("-d", "-c", utils.OcpasswdPath, username)
	if err != nil {
		return "", err
	}

	if err := u.DeleteConfig(username); err != nil {
		return "", err
	}

	return output, nil

}

func (u *OcservUser) SyncConfig(username, group string, config *models.OcservUserConfig) error {
	filename := utils.UserConfigFilePathCreator(username)

	if err := os.MkdirAll(utils.ConfigUserBaseDir, 0750); err != nil {
		return err
	}

	if err := os.Remove(filename); err != nil && !os.IsNotExist(err) {
		return err
	}

	if hasConfigValues(config) {
		return u.CreateConfig(username, config)
	}

	group = strings.TrimSpace(group)
	if group == "" || group == "defaults" {
		return nil
	}

	groupConfig := utils.GroupConfigFilePathCreator(group)
	if _, err := os.Stat(groupConfig); err != nil {
		return err
	}

	return os.Symlink(groupConfig, filename)
}

func hasConfigValues(config *models.OcservUserConfig) bool {
	if config == nil {
		return false
	}

	for _, value := range utils.ToMap(config) {
		switch v := value.(type) {
		case nil:
			continue
		case bool:
			if v {
				return true
			}
		case string:
			if strings.TrimSpace(v) != "" {
				return true
			}
		case []interface{}:
			if len(v) > 0 {
				return true
			}
		default:
			return true
		}
	}

	return false
}

// CreateConfig writes a per-user configuration file for the given username.
// The configuration is serialized from OcservUserConfig using pkg.ConfigWriter.
// The file is created with permission 0640 and stored in the user config directory.
func (u *OcservUser) CreateConfig(username string, config *models.OcservUserConfig) error {
	if !hasConfigValues(config) {
		return nil
	}

	filename := utils.UserConfigFilePathCreator(username)

	if err := os.MkdirAll(utils.ConfigUserBaseDir, 0750); err != nil {
		return err
	}

	file, err := os.OpenFile(filename, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0640)
	if err != nil {
		return err
	}
	defer file.Close()

	return utils.ConfigWriter(file, utils.ToMap(config))
}

// DeleteConfig removes the per-user configuration file for the given username.
// The config file path is derived from UserConfigFilePathCreator.
func (u *OcservUser) DeleteConfig(username string) error {
	filename := utils.UserConfigFilePathCreator(username)
	if err := os.Remove(filename); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// Ocpasswd reads the ocpasswd file and returns a list of all user entries.
// Each line of the ocpasswd file describes one user, including their username,
// password hash information, and optional attributes such as assigned groups.
//
// For each valid user entry, Sync parses the username and extracts the list of
// groups from the "groups=" attribute if present. Commented or malformed lines
// are skipped silently.
//
// The returned slice contains one OcpasswdSync object per user, including the
// raw line from the file for debugging or additional processing.
//
// If the ocpasswd file cannot be opened or read, an error is returned.
func (u *OcservUser) Ocpasswd(ctx context.Context) ([]Ocpasswd, int, error) {
	f, err := os.Open(utils.OcpasswdPath)
	if err != nil {
		return nil, 0, err
	}
	defer f.Close()

	const maxCapacity = 4 * 1024 * 1024 // 4 MB
	scanner := bufio.NewScanner(f)
	buf := make([]byte, maxCapacity)
	scanner.Buffer(buf, maxCapacity)

	var users []Ocpasswd

	for scanner.Scan() {
		if err = ctx.Err(); err != nil {
			return nil, 0, err
		}

		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.Split(line, ":")
		if len(parts) < 3 {
			continue // malformed
		}

		username := parts[0]
		group := parts[1]
		if group == "*" {
			group = "defaults"
		}

		users = append(users, Ocpasswd{
			Username: username,
			Group:    group,
		})

	}

	if err = scanner.Err(); err != nil {
		return nil, 0, err
	}

	total, err := OcpasswdTotalLines(utils.OcpasswdPath)
	if err != nil {
		total = 0
	}

	return users, total, nil
}
