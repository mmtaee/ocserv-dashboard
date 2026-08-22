package ocservuser

type Usecase struct {
	Repository
	accounts AccountStore
	occtl    OCCTL
	reports  Reports
}

func New(repo Repository, accounts AccountStore, occtl OCCTL, reports Reports) *Usecase {
	return &Usecase{Repository: repo, accounts: accounts, occtl: occtl, reports: reports}
}
