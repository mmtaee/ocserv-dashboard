package occtl

type CommandParamsData struct {
	Action int    `query:"action" validate:"required,min=1,max=13"`
	Value  string `query:"value" validate:"omitempty"`
}
