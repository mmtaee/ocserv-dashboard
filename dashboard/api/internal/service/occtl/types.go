package occtl

type CommandParamsData struct {
	Action int    `json:"action" query:"action" validate:"required,min=1,max=16"`
	Value  string `json:"value" query:"value" validate:"omitempty"`
}
