package super_admin

// CreateAdminRequest represents the request body for creating an admin
type CreateAdminRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required,min=8"`
}

// UpdateAdminRequest represents the request body for updating an admin
type UpdateAdminRequest struct {
	Username string `json:"username" validate:"required"`
}

// ChangeAdminPasswordRequest represents the request body for changing an admin's password
type ChangeAdminPasswordRequest struct {
	Password string `json:"password" validate:"required,min=8"`
}

// SuspendAdminRequest represents the request body for suspending an admin
type SuspendAdminRequest struct {
	Reason string `json:"reason" validate:"required"`
}
