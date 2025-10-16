package utils

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

const (
	OcpasswdPath       = "/etc/ocserv/ocpasswd"
	OcpasswdExec       = "/usr/bin/ocpasswd"
	ConfigGroupBaseDir = "/etc/ocserv/groups/"
	DefaultGroupFile   = "/etc/ocserv/defaults/group.conf"
	ConfigUserBaseDir  = "/etc/ocserv/users/"
)

var listKeys = map[string]bool{
	"dns":       true,
	"no-route":  true,
	"route":     true,
	"split-dns": true,
}

// ToMap converts any struct or value into a map[string]interface{}.
// It marshals the value into JSON and then unmarshals it into a map.
// Returns nil if marshaling or unmarshaling fails.
func ToMap(data interface{}) map[string]interface{} {
	by, err := json.Marshal(data)
	if err != nil {
		return nil
	}

	var result map[string]interface{}
	err = json.Unmarshal(by, &result)
	if err != nil {
		return nil
	}
	return result
}

// ConfigWriter writes key-value pairs from a configuration map to the given file.
// It skips nil values and boolean false values. For keys like dns, route, no-route,
// and split-dns, it writes multiple lines for each entry. Other keys are written
// as "key=value". Returns an error if writing fails.
func ConfigWriter(file *os.File, config map[string]interface{}) error {
	for k, v := range config {
		if b, ok := v.(bool); ok && !b {
			continue
		}
		if v == nil {
			continue
		}

		if k == "dns" {
			for _, dns := range v.([]interface{}) {
				if _, err := file.WriteString(fmt.Sprintf("dns=%s\n", dns)); err != nil {
					return fmt.Errorf("failed to write to file: %w", err)
				}
			}
			continue
		} else if k == "route" {
			for _, route := range v.([]interface{}) {
				if _, err := file.WriteString(fmt.Sprintf("route=%s\n", route)); err != nil {
					return fmt.Errorf("failed to write to file: %w", err)
				}
			}
			continue
		} else if k == "no-route" {
			for _, route := range v.([]interface{}) {
				if _, err := file.WriteString(fmt.Sprintf("no-route=%s\n", route)); err != nil {
					return fmt.Errorf("failed to write to file: %w", err)
				}
			}
			continue
		} else if k == "split-dns" {
			for _, dns := range v.([]interface{}) {
				if _, err := file.WriteString(fmt.Sprintf("split-dns=%s\n", dns)); err != nil {
					return fmt.Errorf("failed to write to file: %w", err)
				}
			}
			continue
		} else {
			if _, err := file.WriteString(fmt.Sprintf("%s=%v\n", k, v)); err != nil {
				return fmt.Errorf("failed to write to file: %w", err)
			}
		}
	}
	return nil
}

// GetUsersByGroup parses the ocpasswd file and returns usernames
// belonging to the given group. It ignores comments, empty lines,
// and malformed entries. Assumes group is stored as the third field
// in colon-separated records. Returns an error if scanning fails.
func GetUsersByGroup(groupName string) ([]string, error) {
	file, err := os.Open(OcpasswdPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var users []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		if strings.TrimSpace(line) == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.Split(line, ":")
		if len(parts) < 3 {
			continue // skip malformed lines
		}

		username := parts[0]
		group := parts[2]

		if group == groupName {
			users = append(users, username)
		}
	}

	if err = scanner.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

// ParseOcservConfigFile parses an ocserv config file into a map[string]interface{}.
// Keys with multiple values (like dns, route, no-route, split-dns) are stored as slices.
// Values are converted into bool, int, float64, or string via ParseTypedValue.
// Comments and empty lines are ignored.
func ParseOcservConfigFile(filePath string) (map[string]interface{}, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	config := make(map[string]interface{})
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		rawValue := strings.TrimSpace(parts[1])

		if listKeys[key] {
			if existing, exists := config[key]; exists {
				config[key] = append(existing.([]string), rawValue)
			} else {
				config[key] = []string{rawValue}
			}
			continue
		}

		parsedValue := ParseTypedValue(rawValue)

		if existing, exists := config[key]; exists {
			switch v := existing.(type) {
			case []interface{}:
				config[key] = append(v, parsedValue)
			default:
				config[key] = []interface{}{v, parsedValue}
			}
		} else {
			config[key] = parsedValue
		}
	}

	if err = scanner.Err(); err != nil {
		return nil, err
	}

	return config, nil
}

// ParseTypedValue attempts to convert a string into a typed value.
// It returns a bool if the string is "true" or "false", an int if it
// parses with strconv.Atoi, a float64 if it parses with strconv.ParseFloat,
// or the original string otherwise.
func ParseTypedValue(s string) interface{} {
	s = strings.TrimSpace(s)

	if s == "true" {
		return true
	}
	if s == "false" {
		return false
	}
	if i, err := strconv.Atoi(s); err == nil {
		return i
	}
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return f
	}
	return s
}

// GetOcservVersion runs "ocserv --version" and extracts the semantic
// version number (X.Y.Z) using regex. Returns an empty string if
// the command fails or no version is found.
func GetOcservVersion() string {
	cmd := exec.Command("ocserv", "--version")

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		log.Println("Command error:", err)
		return ""
	}
	// Combine stdout and stderr for pattern matching
	fullOutput := out.String() + stderr.String()
	// Regex to find the version number
	re := regexp.MustCompile(`OpenConnect VPN Server\s+([0-9]+\.[0-9]+\.[0-9]+)`)
	match := re.FindStringSubmatch(fullOutput)
	if len(match) >= 2 {
		return strings.TrimSpace(match[1])
	}
	return ""
}

// GetOCCTLVersion runs "occtl --version" and extracts its version output.
// It removes empty lines, stops parsing at the Copyright line, and
// joins remaining lines into a clean version string. Returns an empty
// string if parsing fails.
func GetOCCTLVersion() string {
	cmd := exec.Command("occtl", "--version")

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		log.Println("Command error:", err)
		return ""
	}

	fullOutput := out.String() + stderr.String()
	lines := strings.Split(fullOutput, "\n")
	var resultLines []string
	for _, line := range lines {
		if trim := strings.TrimSpace(line); trim != "" {
			if strings.HasPrefix(trim, "Copyright") {
				break
			}
			resultLines = append(resultLines, trim)
		}
	}

	if len(resultLines) == 0 {
		return ""
	}

	finalOutput := strings.Join(resultLines, "\n")
	return finalOutput
}

// RunOcpasswd runs the ocpasswd command with the given arguments.
// Returns combined output and error. If the command fails without
// output, the error string is used as output.
func RunOcpasswd(args ...string) (string, error) {
	cmd := exec.Command(OcpasswdExec, args...)
	out, err := cmd.CombinedOutput()
	output := string(out)
	if err != nil {
		if output == "" {
			output = err.Error()
		}
		return output, err
	}
	return output, nil
}

// ConfigFilePathCreator constructs the absolute file path for a
// user-specific config file using ocserv.ConfigUserBaseDir.
func ConfigFilePathCreator(username string) string {
	return filepath.Join(ConfigUserBaseDir, username)
}

//// FixTrailingComma removes a trailing comma after the "in_use" key
//// in JSON output. This is used to clean up invalid JSON emitted by
//// some ocserv/occtl commands before unmarshalling.
//func FixTrailingComma(jsonBytes []byte) []byte {
//	re := regexp.MustCompile(`("in_use"\s*:\s*\d+)\s*,`)
//	return re.ReplaceAll(jsonBytes, []byte("$1"))
//}
