package user

import (
	"bytes"
	"github.com/mmtaee/ocserv-users-management/common/models"
	"github.com/mmtaee/ocserv-users-management/common/pkg/utils"
	"os"
	"os/exec"
	"path/filepath"
)

type OcservUser struct{}

type OcservUserInterface interface {
	Create(username, group, password string, config *models.OcservUserConfig) error
	Lock(username string) (string, error)
	UnLock(username string) (string, error)
	Delete(username string) (string, error)
	CreateConfig(username string, config *models.OcservUserConfig) error
	DeleteConfig(username string) error
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

	if config != nil {
		filename := filepath.Join(utils.ConfigUserBaseDir, username)

		var file *os.File

		// Open file with create, truncate, write-only flags and permission 0640
		file, err = os.OpenFile(filename, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0640)
		if err != nil {
			return err
		}
		defer file.Close()

		err = utils.ConfigWriter(file, utils.ToMap(config))
		if err != nil {
			return err
		}
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
	return output, nil
}

// UnLock re-enables a previously locked user account by running ocpasswd
// with the -u flag. Returns the command output or an error.
func (u *OcservUser) UnLock(username string) (string, error) {
	output, err := utils.RunOcpasswd("-u", "-c", utils.OcpasswdPath, username)
	if err != nil {
		return "", err
	}
	return output, nil
}

// Delete removes a user account from ocserv by running ocpasswd with the -d flag.
// Returns the command output or an error.
func (u *OcservUser) Delete(username string) (string, error) {
	output, err := utils.RunOcpasswd("-d", "-c", utils.OcpasswdPath, username)
	if err != nil {
		return "", err
	}
	return output, nil
}

// CreateConfig writes a per-user configuration file for the given username.
// The configuration is serialized from OcservUserConfig using pkg.ConfigWriter.
// The file is created with permission 0640 and stored in the user config directory.
func (u *OcservUser) CreateConfig(username string, config *models.OcservUserConfig) error {
	filename := utils.ConfigFilePathCreator(username)
	// Open file with create, truncate, write-only flags and permission 0640
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0640)
	if err != nil {
		return err
	}
	err = utils.ConfigWriter(file, utils.ToMap(config))
	if err != nil {
		return err
	}
	return nil
}

// DeleteConfig removes the per-user configuration file for the given username.
// The config file path is derived from ConfigFilePathCreator. If the file does
// not exist or cannot be removed, an error is returned.
func (u *OcservUser) DeleteConfig(username string) error {
	filename := utils.ConfigFilePathCreator(username)
	if err := os.Remove(filename); err != nil {
		return err
	}
	return nil
}
