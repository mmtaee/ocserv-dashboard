package systemd

// Usecase owns the systemd application operations.
type Usecase struct {
	Repository
}

func New(repo Repository) *Usecase {
	return &Usecase{Repository: repo}
}
