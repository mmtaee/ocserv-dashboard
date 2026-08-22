package occtl

// Usecase owns the occtl application operations.
type Usecase struct {
	Repository
}

func New(repo Repository) *Usecase {
	return &Usecase{Repository: repo}
}
