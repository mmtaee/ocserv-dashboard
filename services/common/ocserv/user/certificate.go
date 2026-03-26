package user

import (
	"fmt"
	"github.com/mmtaee/ocserv-users-management/common/pkg/utils"
	"os"
	"os/exec"
	"path/filepath"
)

func GenerateUserCertificate(username string) error {
	if _, err := os.Stat(utils.CertBaseDir); os.IsNotExist(err) {
		err := os.MkdirAll(utils.CertBaseDir, 0755)
		if err != nil {
			return fmt.Errorf("failed to create cert directory: %w", err)
		}
	}

	tempDir, err := os.MkdirTemp("", "cert_*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempDir)

	userKey := filepath.Join(tempDir, "user-key.pem")
	userCert := filepath.Join(tempDir, "user-cert.pem")
	userTmpl := filepath.Join(tempDir, "user.tmpl")
	p12File := utils.UserCertPathCreator(username)

	// 1. Generate private key
	cmd := exec.Command("certtool", "--generate-privkey", "--outfile", userKey)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to generate private key: %v, output: %s", err, string(out))
	}

	// 2. Create template
	tmplContent := fmt.Sprintf("cn = \"%s\"\nuid = \"%s\"\nunit = \"users\"\nexpiration_days = 3650\nsigning_key\nencryption_key\ntls_www_client\n", username, username)
	if err := os.WriteFile(userTmpl, []byte(tmplContent), 0644); err != nil {
		return err
	}

	// 3. Generate certificate
	cmd = exec.Command("certtool", "--generate-certificate",
		"--load-privkey", userKey,
		"--load-ca-certificate", utils.ClientCACertPath,
		"--load-ca-privkey", utils.ClientCAKeyPath,
		"--template", userTmpl,
		"--outfile", userCert)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to generate certificate: %v, output: %s", err, string(out))
	}

	// 4. Export to P12
	cmd = exec.Command("certtool", "--to-p12",
		"--p12-name", username,
		"--load-privkey", userKey,
		"--load-certificate", userCert,
		"--pkcs-cipher", "3des-pkcs12",
		"--password", "1234",
		"--outfile", p12File,
		"--outder")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to export to p12: %v, output: %s", err, string(out))
	}

	return nil
}

