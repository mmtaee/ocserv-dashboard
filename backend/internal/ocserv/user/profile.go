package user

import (
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
)

func NormalizeProfileConnectionName(connectionName string) (string, error) {
	connectionName = strings.TrimSpace(connectionName)

	if connectionName == "" {
		return "", fmt.Errorf("client profile connection name is required")
	}

	if len(connectionName) > 64 {
		return "", fmt.Errorf("client profile connection name must be 64 characters or less")
	}

	return connectionName, nil
}

func NormalizeProfileServerAddress(serverAddress string) (string, error) {
	serverAddress = strings.TrimSpace(serverAddress)

	if serverAddress == "" {
		return "", fmt.Errorf("client profile server address is required")
	}

	if strings.Contains(serverAddress, "://") {
		return "", fmt.Errorf("client profile server address must not include a URL scheme")
	}

	if strings.ContainsAny(serverAddress, "/\\ \t\r\n") {
		return "", fmt.Errorf("client profile server address must be a host name or IP address")
	}

	if strings.Contains(serverAddress, ":") {
		return "", fmt.Errorf("client profile server address must not include the port")
	}

	return serverAddress, nil
}

func NormalizeProfileServerPort(serverPort int) (int, error) {
	if serverPort < 1 || serverPort > 65535 {
		return 0, fmt.Errorf("client profile server port must be between 1 and 65535")
	}

	return serverPort, nil
}

func BuildProfileServerHostPort(serverAddress string, serverPort int) (string, error) {
	serverAddress, err := NormalizeProfileServerAddress(serverAddress)
	if err != nil {
		return "", err
	}

	serverPort, err = NormalizeProfileServerPort(serverPort)
	if err != nil {
		return "", err
	}

	return net.JoinHostPort(serverAddress, strconv.Itoa(serverPort)), nil
}

func BuildAnyConnectCreateURI(connectionName, serverAddress string, serverPort int, certCommonName string) (string, error) {
	connectionName, err := NormalizeProfileConnectionName(connectionName)
	if err != nil {
		return "", err
	}

	hostPort, err := BuildProfileServerHostPort(serverAddress, serverPort)
	if err != nil {
		return "", err
	}

	values := url.Values{}
	values.Set("name", connectionName)
	values.Set("host", hostPort)
	values.Set("usecert", "true")

	certCommonName = strings.TrimSpace(certCommonName)
	if certCommonName != "" {
		values.Set("certcommonname", certCommonName)
	}

	return "anyconnect://create/?" + values.Encode(), nil
}

func BuildAnyConnectImportURI(certificateURL string) (string, error) {
	certificateURL = strings.TrimSpace(certificateURL)

	if certificateURL == "" {
		return "", fmt.Errorf("certificate URL is required")
	}

	if !strings.HasPrefix(certificateURL, "https://") {
		return "", fmt.Errorf("certificate URL must use HTTPS")
	}

	values := url.Values{}
	values.Set("type", "pkcs12")
	values.Set("uri", certificateURL)

	return "anyconnect://import/?" + values.Encode(), nil
}
