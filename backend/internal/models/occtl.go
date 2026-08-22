package models

type IPBan struct {
	IP       string `json:"IP"`
	Since    string `json:"Since"`
	SinceAlt string `json:"_Since"` // maps to "_Since" in JSON
	Score    int    `json:"Score"`
}

type Iroute struct {
	ID       int      `json:"ID"`
	Username string   `json:"Username"`
	VHost    string   `json:"vhost"`
	Device   string   `json:"Device"`
	IP       string   `json:"IP"`
	IRoutes  []string `json:"iRoutes"`
}

type OnlineUserSession struct {
	ID               int    `json:"ID" validate:"required"`
	Username         string `json:"Username"`
	Group            string `json:"Groupname"`
	AverageRX        string `json:"Average RX"`
	AverageTX        string `json:"Average TX"`
	LastConnectedAt  string `json:"_Last connected at"`
	IPv4             string `json:"IPv4" validate:"required"`
	VHost            string `json:"vhost" validate:"required"`
	Device           string `json:"Device" validate:"required"`
	SessionStartedAt string `json:"Session started at" validate:"required"`
}

type ServerVersion struct {
	OcservVersion string `json:"ocserv_version"`
	OcctlVersion  string `json:"occtl_version"`
}

type OcservInfo struct {
	Version *ServerVersion `json:"version" validate:"required"`
	Status  string         `json:"status" validate:"required"`
}

type IPBanPoints struct {
	IP    string `json:"IP"`
	Since string `json:"Since"`
	Until string `json:"_Since"`
	Score int    `json:"Score"`
}

type IRoute struct {
	ID       string `json:"ID"`
	Username string `json:"Username"`
	Vhost    string `json:"vhost"`
	Device   string `json:"Device"`
	IP       string `json:"IP"`
	IRoute   string `json:"iRoutes"`
}
