package telegram

import tg18n "github.com/mmtaee/ocserv-dashboard/backend/internal/usecase/admin_api/telegram/i18n"

type Usecase struct {
	Repository
	users  UserManager
	client Client
}

func New(repo Repository, users UserManager, client Client) *Usecase {
	tg18n.Init()
	return &Usecase{Repository: repo, users: users, client: client}
}
