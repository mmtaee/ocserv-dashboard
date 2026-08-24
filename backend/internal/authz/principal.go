package authz

import "errors"

var ErrForbidden = errors.New("permission denied")

type Principal struct {
	UserID     uint
	Username   string
	Superadmin bool
}

func (p Principal) RequireSuperadmin() error {
	if p.UserID == 0 || !p.Superadmin {
		return ErrForbidden
	}
	return nil
}

func (p Principal) RequireOwner(ownerID uint) error {
	if p.UserID != 0 && (p.Superadmin || p.UserID == ownerID) {
		return nil
	}
	return ErrForbidden
}
