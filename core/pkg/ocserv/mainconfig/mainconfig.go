package mainconfig

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/mmtaee/ocserv-dashboard/core/models"
	"github.com/mmtaee/ocserv-dashboard/core/pkg/utils"
)

// ConfigPath is the default path to ocserv main config file
const ConfigPath = "/etc/ocserv/ocserv.conf"

// MainConfigInterface defines methods for main config operations
type MainConfigInterface interface {
	Read(ctx context.Context) (*models.OcservMainConfig, error)
	Write(ctx context.Context, config *models.OcservMainConfig) error
}

// MainConfigRepository implements MainConfigInterface
type MainConfigRepository struct {
	configPath string
	lastHeader []string // preserve header between reads and writes
}

// NewMainConfigRepository creates a new MainConfigRepository
func NewMainConfigRepository() MainConfigInterface {
	return &MainConfigRepository{
		configPath: ConfigPath,
	}
}

// Read reads and parses the main config file into a models.OcservMainConfig
func (m *MainConfigRepository) Read(ctx context.Context) (*models.OcservMainConfig, error) {
	contentBytes, err := os.ReadFile(m.configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}
	content := string(contentBytes)

	configMap, header, err := utils.ParseOcservConfigContent(content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}
	m.lastHeader = header

	config, err := utils.MainConfigToModel(configMap)
	if err != nil {
		return nil, fmt.Errorf("failed to convert to model: %w", err)
	}
	return &config, nil
}

// Write serializes and writes a models.OcservMainConfig to the main config file
func (m *MainConfigRepository) Write(ctx context.Context, config *models.OcservMainConfig) error {
	// Read current file to get header if not already stored
	if len(m.lastHeader) == 0 {
		contentBytes, err := os.ReadFile(m.configPath)
		if err == nil {
			_, m.lastHeader, _ = utils.ParseOcservConfigContent(string(contentBytes))
		}
	}

	configMap := utils.ToMap(config)

	buffer := &bytes.Buffer{}
	err := utils.ConfigWriterForMain(buffer, configMap, m.lastHeader)
	if err != nil {
		return fmt.Errorf("failed to serialize config: %w", err)
	}

	// Use sudo to write the file
	cmd := exec.CommandContext(ctx, "sudo", "tee", m.configPath)
	cmd.Stdin = buffer
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to write config file: %w, stderr: %s", err, stderr.String())
	}
	return nil
}
