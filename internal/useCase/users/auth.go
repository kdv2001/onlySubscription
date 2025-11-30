package user

import (
	"context"

	domain "github.com/kdv2001/onlySubscription/internal/domain/user"
)

type userRepo interface {
	RegisterTelegram(ctx context.Context, telegramBotLogin domain.TelegramBotRegister) (domain.ID, error)
	GetUserByTelegramID(ctx context.Context, user domain.TelegramID) (domain.User, error)
	GetUser(ctx context.Context, user domain.ID) (domain.User, error)
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

// RegisterByTelegramID регистрирует пользователя по данным telegram
func (a *Implementation) RegisterByTelegramID(ctx context.Context,
	telegramBotLogin domain.TelegramBotRegister) (domain.ID, error) {
	return a.userRepo.RegisterTelegram(ctx, telegramBotLogin)
}

// GetUserByTelegramID возвращает пользователя по telegramID
func (a *Implementation) GetUserByTelegramID(ctx context.Context,
	tgID domain.TelegramID) (domain.User, error) {
	return a.userRepo.GetUserByTelegramID(ctx, tgID)
}

// GetUser возвращает пользователя
func (a *Implementation) GetUser(ctx context.Context, id domain.ID) (domain.User, error) {
	return a.userRepo.GetUser(ctx, id)
}
