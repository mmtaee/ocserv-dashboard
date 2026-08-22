package telegram

// Usecase owns the telegram application operations.
type Usecase struct {
	Repository
}

func New(repo Repository) *Usecase {
	return &Usecase{Repository: repo}
}
