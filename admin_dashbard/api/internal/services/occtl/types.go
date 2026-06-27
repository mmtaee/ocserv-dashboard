package occtl

type CommandParamsData struct {
	Action int    `query:"action" validate:"required,min=1,max=16"`
	Value  string `query:"value" validate:"omitempty"`
}
