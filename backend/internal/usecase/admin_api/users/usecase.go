package ocservuser

type Usecase struct {
	Repository
	accounts AccountStore
	occtl    OCCTL
	reports  Reports
	bulk     *BulkUsecase
}

func New(repo Repository, accounts AccountStore, occtl OCCTL, reports Reports) *Usecase {
	return &Usecase{Repository: repo, accounts: accounts, occtl: occtl, reports: reports, bulk: NewBulk(repo, accounts, occtl)}
}
