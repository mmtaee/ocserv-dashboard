package user

import (
	"encoding/base64"
	"fmt"
	"github.com/mmtaee/ocserv-dashboard/common/models"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

const (
	certSSLDir       = "/etc/ocserv/ssl"
	certUsersDir     = certSSLDir + "/users"
	certDisabledDir  = certSSLDir + "/disabled"
	certCACertPath   = certSSLDir + "/ca-cert.pem"
	certCAKeyPath    = certSSLDir + "/ca-key.pem"
	certCRLPath      = certSSLDir + "/crl.pem"
	certCRLTmplPath  = certSSLDir + "/crl.tmpl"
	certRevokedPath  = certSSLDir + "/revoked.pem"
	certSuspendedPEM = certSSLDir + "/suspended.pem"
	certtoolExec     = "/usr/bin/certtool"
	opensslExec      = "/usr/bin/openssl"
)

var certificateUsernameRe = regexp.MustCompile(`^[A-Za-z0-9._-]{1,64}$`)

type CertificateStatus struct {
	Available bool
	Enabled   bool
}

func (u *OcservUser) CertificateStatus(username string) CertificateStatus {
	active := fileExists(userCertificateFile(username, "cer"))
	suspended := latestSuspendedCertificateDir(username) != ""

	return CertificateStatus{
		Available: active || suspended,
		Enabled:   active,
	}
}

func (u *OcservUser) CertificatePath(username string) (string, error) {
	if !validCertificateUsername(username) {
		return "", fmt.Errorf("invalid username: %s", username)
	}

	activePath := userCertificateFile(username, "p12")
	if fileExists(activePath) {
		return activePath, nil
	}

	suspendedDir := latestSuspendedCertificateDir(username)
	if suspendedDir != "" {
		p12Path := filepath.Join(suspendedDir, username+".p12")
		if fileExists(p12Path) {
			return p12Path, nil
		}
	}

	return "", fmt.Errorf("certificate not found for user %s", username)
}

func (u *OcservUser) CreateCertificate(username, password string) error {
	if !validCertificateUsername(username) {
		return fmt.Errorf("invalid username: %s", username)
	}

	status := u.CertificateStatus(username)
	if status.Available {
		return nil
	}

	if err := ensureCertificatePKI(); err != nil {
		return err
	}

	userDir := filepath.Join(certUsersDir, username)
	if err := os.MkdirAll(userDir, 0700); err != nil {
		return err
	}

	cleanup := true
	defer func() {
		if cleanup {
			_ = os.RemoveAll(userDir)
		}
	}()

	keyPath := filepath.Join(userDir, username+"-key.pem")
	tmplPath := filepath.Join(userDir, username+".tmpl")
	certPath := filepath.Join(userDir, username+".cer")
	p12Path := filepath.Join(userDir, username+".p12")

	if err := runCommand(certtoolExec, "--generate-privkey", "--outfile", keyPath); err != nil {
		return err
	}

	activationDate := time.Now().UTC().AddDate(0, 0, -1).Format("2006-01-02 15:04:05 UTC")

	tmpl := fmt.Sprintf(`cn = "%s"
tls_www_client
encryption_key
signing_key
activation_date = "%s"
expiration_days = 825
`, username, activationDate)

	if err := os.WriteFile(tmplPath, []byte(tmpl), 0600); err != nil {
		return err
	}

	if err := runCommand(
		certtoolExec,
		"--generate-certificate",
		"--load-privkey", keyPath,
		"--load-ca-certificate", certCACertPath,
		"--load-ca-privkey", certCAKeyPath,
		"--template", tmplPath,
		"--outfile", certPath,
	); err != nil {
		return err
	}

	if err := runCommand(
		opensslExec,
		"pkcs12",
		"-export",
		"-legacy",
		"-keypbe", "PBE-SHA1-3DES",
		"-certpbe", "PBE-SHA1-3DES",
		"-macalg", "sha1",
		"-inkey", keyPath,
		"-in", certPath,
		"-certfile", certCACertPath,
		"-name", "AnyConnect VPN - "+username,
		"-out", p12Path,
		"-passout", "pass:"+password,
	); err != nil {
		return err
	}

	if err := os.Chmod(keyPath, 0600); err != nil {
		return err
	}
	if err := os.Chmod(p12Path, 0600); err != nil {
		return err
	}

	cleanup = false
	return nil
}

func (u *OcservUser) SuspendCertificate(username string) error {
	if !validCertificateUsername(username) {
		return fmt.Errorf("invalid username: %s", username)
	}

	activeDir := filepath.Join(certUsersDir, username)
	if !fileExists(filepath.Join(activeDir, username+".cer")) {
		return nil
	}

	if err := ensureCertificatePKI(); err != nil {
		return err
	}

	targetDir := filepath.Join(
		certDisabledDir,
		fmt.Sprintf("%s-susp-%s", username, time.Now().Format("20060102-150405")),
	)

	if err := os.MkdirAll(certDisabledDir, 0700); err != nil {
		return err
	}

	if err := os.Rename(activeDir, targetDir); err != nil {
		return err
	}

	return rebuildCertificateCRL()
}

func (u *OcservUser) UnsuspendCertificate(username string) error {
	if !validCertificateUsername(username) {
		return fmt.Errorf("invalid username: %s", username)
	}

	activeDir := filepath.Join(certUsersDir, username)
	if fileExists(filepath.Join(activeDir, username+".cer")) {
		return nil
	}

	suspendedDir := latestSuspendedCertificateDir(username)
	if suspendedDir == "" {
		return nil
	}

	if err := os.MkdirAll(certUsersDir, 0700); err != nil {
		return err
	}

	if err := os.Rename(suspendedDir, activeDir); err != nil {
		return err
	}

	return rebuildCertificateCRL()
}

func (u *OcservUser) RevokeCertificate(username string) error {
	if !validCertificateUsername(username) {
		return fmt.Errorf("invalid username: %s", username)
	}

	if err := ensureCertificatePKI(); err != nil {
		return err
	}

	changed := false

	activeDir := filepath.Join(certUsersDir, username)
	activeCert := filepath.Join(activeDir, username+".cer")
	if fileExists(activeCert) {
		if err := appendCertificateToFile(activeCert, certRevokedPath); err != nil {
			return err
		}
		if err := os.RemoveAll(activeDir); err != nil {
			return err
		}
		changed = true
	}

	for _, dir := range suspendedCertificateDirs(username) {
		certPath := filepath.Join(dir, username+".cer")
		if fileExists(certPath) {
			if err := appendCertificateToFile(certPath, certRevokedPath); err != nil {
				return err
			}
			if err := os.RemoveAll(dir); err != nil {
				return err
			}
			changed = true
		}
	}

	if changed {
		return rebuildCertificateCRL()
	}

	return nil
}

func (u *OcservUser) CertificateBackup(username string) (*models.OcservUserCertificateBackup, error) {
	if !validCertificateUsername(username) {
		return nil, fmt.Errorf("invalid username: %s", username)
	}

	status := "active"
	certDir := filepath.Join(certUsersDir, username)
	if !fileExists(filepath.Join(certDir, username+".cer")) {
		certDir = latestSuspendedCertificateDir(username)
		status = "suspended"
	}

	if certDir == "" {
		return nil, nil
	}

	keyPEM, err := os.ReadFile(filepath.Join(certDir, username+"-key.pem"))
	if err != nil {
		return nil, err
	}

	certPEM, err := os.ReadFile(filepath.Join(certDir, username+".cer"))
	if err != nil {
		return nil, err
	}

	p12, err := os.ReadFile(filepath.Join(certDir, username+".p12"))
	if err != nil {
		return nil, err
	}

	return &models.OcservUserCertificateBackup{
		Status:    status,
		KeyPEM:    string(keyPEM),
		CertPEM:   string(certPEM),
		P12Base64: base64.StdEncoding.EncodeToString(p12),
	}, nil
}

func (u *OcservUser) RestoreCertificateBackup(username string, cert *models.OcservUserCertificateBackup) error {
	if cert == nil {
		return nil
	}

	if !validCertificateUsername(username) {
		return fmt.Errorf("invalid username: %s", username)
	}

	if err := ensureCertificatePKI(); err != nil {
		return err
	}

	_ = os.RemoveAll(filepath.Join(certUsersDir, username))
	for _, dir := range suspendedCertificateDirs(username) {
		_ = os.RemoveAll(dir)
	}

	targetDir := filepath.Join(certUsersDir, username)
	if cert.Status == "suspended" {
		targetDir = filepath.Join(
			certDisabledDir,
			fmt.Sprintf("%s-susp-%s", username, time.Now().Format("20060102-150405")),
		)
	}

	if err := os.MkdirAll(targetDir, 0700); err != nil {
		return err
	}

	if err := os.WriteFile(filepath.Join(targetDir, username+"-key.pem"), []byte(cert.KeyPEM), 0600); err != nil {
		return err
	}

	if err := os.WriteFile(filepath.Join(targetDir, username+".cer"), []byte(cert.CertPEM), 0600); err != nil {
		return err
	}

	p12, err := base64.StdEncoding.DecodeString(cert.P12Base64)
	if err != nil {
		return err
	}

	if err := os.WriteFile(filepath.Join(targetDir, username+".p12"), p12, 0600); err != nil {
		return err
	}

	return rebuildCertificateCRL()
}

func ensureCertificatePKI() error {
	if err := ensureCertificateDirsAndFiles(); err != nil {
		return err
	}

	caCertExists := fileExists(certCACertPath)
	caKeyExists := fileExists(certCAKeyPath)

	if caCertExists != caKeyExists {
		return fmt.Errorf("incomplete certificate CA: both %s and %s must exist", certCACertPath, certCAKeyPath)
	}

	if !caCertExists {
		if err := generateCertificateCA(); err != nil {
			return err
		}
	}

	if !fileExists(certCRLPath) {
		if err := generateCertificateCRL(""); err != nil {
			return err
		}
	}

	return nil
}

func ensureCertificateDirsAndFiles() error {
	if err := os.MkdirAll(certSSLDir, 0700); err != nil {
		return err
	}
	if err := os.MkdirAll(certUsersDir, 0700); err != nil {
		return err
	}
	if err := os.MkdirAll(certDisabledDir, 0700); err != nil {
		return err
	}

	if !fileExists(certCRLTmplPath) {
		if err := os.WriteFile(certCRLTmplPath, []byte("crl_next_update = 365\ncrl_number = 1\n"), 0600); err != nil {
			return err
		}
	}

	for _, path := range []string{certRevokedPath, certSuspendedPEM} {
		if !fileExists(path) {
			if err := os.WriteFile(path, []byte{}, 0600); err != nil {
				return err
			}
		}
	}

	return nil
}

func generateCertificateCA() error {
	tmplPath := filepath.Join(certSSLDir, "ca.tmpl")

	tmpl := `cn = "Ocserv Dashboard CA"
organization = "Ocserv Dashboard"
serial = 1
expiration_days = 3650
ca
signing_key
cert_signing_key
crl_signing_key
`

	if err := os.WriteFile(tmplPath, []byte(tmpl), 0600); err != nil {
		return err
	}

	if err := runCommand(certtoolExec, "--generate-privkey", "--outfile", certCAKeyPath); err != nil {
		return err
	}

	return runCommand(
		certtoolExec,
		"--generate-self-signed",
		"--load-privkey", certCAKeyPath,
		"--template", tmplPath,
		"--outfile", certCACertPath,
	)
}

func rebuildCertificateCRL() error {
	if err := ensureCertificateDirsAndFiles(); err != nil {
		return err
	}

	if err := rebuildSuspendedPEM(); err != nil {
		return err
	}

	tmpPath := filepath.Join(certSSLDir, ".crl-input.tmp")
	if err := os.WriteFile(tmpPath, []byte{}, 0600); err != nil {
		return err
	}
	defer os.Remove(tmpPath)

	for _, path := range []string{certRevokedPath, certSuspendedPEM} {
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if len(strings.TrimSpace(string(content))) > 0 {
			if err := appendBytesToFile(tmpPath, content); err != nil {
				return err
			}
		}
	}

	if fileExists(tmpPath) {
		content, err := os.ReadFile(tmpPath)
		if err != nil {
			return err
		}
		if len(strings.TrimSpace(string(content))) > 0 {
			if err := generateCertificateCRL(tmpPath); err != nil {
				return err
			}
			signalOcservReloadCRL()
			return nil
		}
	}

	if err := generateCertificateCRL(""); err != nil {
		return err
	}

	signalOcservReloadCRL()
	return nil
}

func rebuildSuspendedPEM() error {
	if err := os.WriteFile(certSuspendedPEM, []byte{}, 0600); err != nil {
		return err
	}

	for _, dir := range suspendedCertificateDirs("*") {
		username := suspendedUsernameFromDir(dir)
		if username == "" {
			continue
		}

		certPath := filepath.Join(dir, username+".cer")
		if fileExists(certPath) {
			if err := appendCertificateToFile(certPath, certSuspendedPEM); err != nil {
				return err
			}
		}
	}

	return nil
}

func generateCertificateCRL(inputPath string) error {
	args := []string{
		"--generate-crl",
		"--load-ca-privkey", certCAKeyPath,
		"--load-ca-certificate", certCACertPath,
		"--template", certCRLTmplPath,
	}

	if inputPath != "" {
		args = append(args, "--load-certificate", inputPath)
	}

	args = append(args, "--outfile", certCRLPath)

	return runCommand(certtoolExec, args...)
}

func signalOcservReloadCRL() {
	pidContent, err := os.ReadFile("/var/run/ocserv.pid")
	if err != nil {
		return
	}

	pid := strings.TrimSpace(string(pidContent))
	if pid == "" {
		return
	}

	_ = exec.Command("/bin/kill", "-HUP", pid).Run()
}

func validCertificateUsername(username string) bool {
	return certificateUsernameRe.MatchString(username) && username != "." && username != ".."
}

func userCertificateFile(username, ext string) string {
	return filepath.Join(certUsersDir, username, username+"."+ext)
}

func latestSuspendedCertificateDir(username string) string {
	dirs := suspendedCertificateDirs(username)
	if len(dirs) == 0 {
		return ""
	}

	sort.Slice(dirs, func(i, j int) bool {
		iInfo, iErr := os.Stat(dirs[i])
		jInfo, jErr := os.Stat(dirs[j])
		if iErr != nil || jErr != nil {
			return dirs[i] > dirs[j]
		}
		return iInfo.ModTime().After(jInfo.ModTime())
	})

	return dirs[0]
}

func suspendedCertificateDirs(username string) []string {
	pattern := filepath.Join(certDisabledDir, username+"-susp-*")
	dirs, _ := filepath.Glob(pattern)
	return dirs
}

func suspendedUsernameFromDir(dir string) string {
	base := filepath.Base(dir)
	parts := strings.Split(base, "-susp-")
	if len(parts) != 2 {
		return ""
	}
	return parts[0]
}

func appendCertificateToFile(srcPath, dstPath string) error {
	content, err := os.ReadFile(srcPath)
	if err != nil {
		return err
	}
	return appendBytesToFile(dstPath, content)
}

func appendBytesToFile(dstPath string, content []byte) error {
	file, err := os.OpenFile(dstPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err = file.Write(content); err != nil {
		return err
	}

	if len(content) > 0 && content[len(content)-1] != '\n' {
		if _, err = file.WriteString("\n"); err != nil {
			return err
		}
	}

	return nil
}

func runCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s %s failed: %s: %w", name, strings.Join(args, " "), strings.TrimSpace(string(out)), err)
	}
	return nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
