package secondaryserver

type SecondaryServerRequest struct {
	Title string `json:"title" validate:"required"`
	IP    string `json:"ip" validate:"required"`
	Port  int    `json:"port" validate:"required"`
	Token string `json:"token" validate:"required"`
}
