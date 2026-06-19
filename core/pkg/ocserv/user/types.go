package user

type Ocpasswd struct {
	Username string `json:"username"`
	Group    string `json:"group"`
}

type CertificateStatus struct {
	Enabled  bool `json:"enabled"`
	Available bool `json:"available"`
}
