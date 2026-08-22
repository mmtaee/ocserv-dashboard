package repository

import shared "github.com/mmtaee/ocserv-dashboard/backend/internal/repository"

type Repository = shared.TelegramBotRepository

func New() *Repository {
	return shared.NewTelegramBotRepository()
}
