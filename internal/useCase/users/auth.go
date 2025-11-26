package user

import (
	"context"

	domain "github.com/kdv2001/onlySubscription/internal/domain/user"
)

type userRepo interface {
	RegisterTelegram(ctx context.Context, telegramBotLogin domain.TelegramBotRegister) (domain.ID, error)
	GetUserByTelegramID(ctx context.Context, user domain.TelegramID) (domain.User, error)
}

type Repo interface {
	GetUser()
}

type Implementation struct {
	userRepo userRepo
}

func NewImplementation(authRepo userRepo) *Implementation {
	return &Implementation{
		userRepo: authRepo,
	}
}

type Register struct {
	Login    string
	Password string
}

func (a *Implementation) RegisterByTelegramID(ctx context.Context,
	telegramBotLogin domain.TelegramBotRegister) (domain.ID, error) {
	return a.userRepo.RegisterTelegram(ctx, telegramBotLogin)
}

func (a *Implementation) GetUserByTelegramID(ctx context.Context,
	tgID domain.TelegramID) (domain.User, error) {
	return a.userRepo.GetUserByTelegramID(ctx, tgID)
}
