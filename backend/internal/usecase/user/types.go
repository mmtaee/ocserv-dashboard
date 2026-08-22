package user

import "github.com/mmtaee/ocserv-dashboard/backend/internal/repository"

// Repository is the persistence boundary required by this usecase.
type Repository interface {
	repository.UserRepositoryInterface
}
