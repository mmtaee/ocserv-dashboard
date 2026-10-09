package user

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

func OcpasswdTotalLines(filePath string) (int, error) {
	// Build the full shell command
	cmd := exec.Command("sh", "-c", fmt.Sprintf("grep -v '^#' %s | grep -v '^$' | wc -l", filePath))

	// Run the command and capture output
	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}

	// Convert output to integer
	countStr := strings.TrimSpace(string(output))
	count, err := strconv.Atoi(countStr)
	if err != nil {
		return 0, err
	}

	return count, nil
}
