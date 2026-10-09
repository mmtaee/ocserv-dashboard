package customer

import (
	"context"

	ocservuser "github.com/mmtaee/ocserv-dashboard/backend/internal/ocserv/user"
)

func (u *Usecase) CiscoSetup(ctx context.Context, username, publicBaseURL string) (*CiscoSetup, error) {
	user, err := u.user(ctx, username)
	if err != nil {
		return nil, err
	}
	system, err := u.systems.System(ctx)
	if err != nil {
		return nil, err
	}
	connectionName, err := ocservuser.NormalizeProfileConnectionName(system.ClientProfileConnectionName)
	if err != nil {
		return nil, err
	}
	serverAddress, err := ocservuser.NormalizeProfileServerAddress(system.ClientProfileServerAddress)
	if err != nil {
		return nil, err
	}
	serverPort, err := ocservuser.NormalizeProfileServerPort(system.ClientProfileServerPort)
	if err != nil {
		return nil, err
	}
	certificateURI, err := ocservuser.BuildAnyConnectImportURI(publicBaseURL + "/api/customers/setup/cisco/certificate")
	if err != nil {
		return nil, err
	}
	connectionURI, err := ocservuser.BuildAnyConnectCreateURI(connectionName, serverAddress, serverPort, user.Username)
	if err != nil {
		return nil, err
	}
	return &CiscoSetup{CertificateImportURI: certificateURI, ConnectionCreateURI: connectionURI, CertificatePassword: user.Password, ConnectionName: connectionName, ServerAddress: serverAddress, ServerPort: serverPort}, nil
}
