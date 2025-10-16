package group

import (
	"encoding/json"
	"github.com/mmtaee/ocserv-users-management/common/models"
	"github.com/mmtaee/ocserv-users-management/common/pkg/utils"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
)

type OcservGroup struct{}

type OcservGroupInterface interface {
	Create(name string, config *models.OcservGroupConfig) error
	Delete(name string) error
	DefaultsGroup() (*models.OcservGroupConfig, error)
	UpdateDefaultsGroup(config *models.OcservGroupConfig) error
}

func NewOcservGroup() *OcservGroup {
	return &OcservGroup{}
}

// Create creates a new group configuration file for the given group name.
// The file is written to ocserv.ConfigGroupBaseDir/<name> with permission 0640.
// It serializes the provided OcservGroupConfig into the file using pkg.ConfigWriter.
func (g *OcservGroup) Create(name string, config *models.OcservGroupConfig) error {
	filename := filepath.Join(utils.ConfigGroupBaseDir, name)

	// Open file with create, truncate, write-only flags and permission 0640
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0640)
	if err != nil {
		return err
	}
	defer file.Close()

	err = utils.ConfigWriter(file, utils.ToMap(config))
	if err != nil {
		return err
	}
	return nil
}

// Delete removes the configuration file for the given group name
// from ocserv.ConfigGroupBaseDir. It also resets the group assignment
// for all users belonging to this group back to the default (empty).
// User group resets are done concurrently, but errors from resetting
// are ignored.
func (g *OcservGroup) Delete(name string) error {
	filename := filepath.Join(utils.ConfigGroupBaseDir, name)
	if err := os.Remove(filename); err != nil {
		return err
	}
	users, err := utils.GetUsersByGroup(name)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup

	// Reset group assignment for each user to default (empty)
	for _, user := range users {
		wg.Add(1)
		go func(u string) {
			defer wg.Done()
			_, _ = exec.Command(utils.OcpasswdExec, "-g", "", "-c", utils.OcpasswdPath, u).CombinedOutput()
		}(user)
	}
	wg.Wait()

	return nil
}

// DefaultsGroup loads and returns the default group configuration
// from ocserv.DefaultGroupFile. The configuration is parsed into
// OcservGroupConfig via an intermediate map and JSON conversion.
func (g *OcservGroup) DefaultsGroup() (*models.OcservGroupConfig, error) {
	configInterface, err := utils.ParseOcservConfigFile(utils.DefaultGroupFile)
	if err != nil {
		return nil, err
	}

	configJson, err := json.Marshal(configInterface)
	if err != nil {
		return nil, err
	}

	var config models.OcservGroupConfig

	if err = json.Unmarshal(configJson, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// UpdateDefaultsGroup overwrites the ocserv.DefaultGroupFile with
// the provided OcservGroupConfig. The file is opened with create,
// truncate, and write-only flags and written with permission 0640.
func (g *OcservGroup) UpdateDefaultsGroup(config *models.OcservGroupConfig) error {
	file, err := os.OpenFile(utils.DefaultGroupFile, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0640)
	if err != nil {
		return err
	}
	defer file.Close()

	err = utils.ConfigWriter(file, utils.ToMap(config))
	if err != nil {
		return err
	}
	return nil
}
