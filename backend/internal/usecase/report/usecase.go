package report

// Usecase owns the report application operations.
type Usecase struct {
	Repository
}

func New(repo Repository) *Usecase {
	return &Usecase{Repository: repo}
}
