package user

// Usecase owns the user application operations.
type Usecase struct {
	Repository
}

func New(repo Repository) *Usecase {
	return &Usecase{Repository: repo}
}
