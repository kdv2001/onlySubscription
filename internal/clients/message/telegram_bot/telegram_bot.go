package telegram_bot

import (
	"context"
	"time"

	telegramBot "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	"github.com/kdv2001/onlySubscription/internal/domain/communication"
	domainPayment "github.com/kdv2001/onlySubscription/internal/domain/payment"
	"github.com/kdv2001/onlySubscription/internal/domain/primitives"
	custom_errors "github.com/kdv2001/onlySubscription/pkg/errors"
)

type messageClient interface {
	SendMessage(ctx context.Context, params *telegramBot.SendMessageParams) (*models.Message, error)
	GetStarTransactions(ctx context.Context, params *telegramBot.GetStarTransactionsParams) (*models.StarTransactions, error)
}

type Implementation struct {
	b messageClient
}

func NewImplementation(getBot messageClient) *Implementation {
	return &Implementation{
		b: getBot,
	}
}

func (i *Implementation) SendMessage(ctx context.Context, message communication.Message) error {
	_, err := i.b.SendMessage(ctx, &telegramBot.SendMessageParams{
		ChatID: message.ChatID,
		Text:   message.Title + "\n" + message.Description,
	})
	if err != nil {
		return custom_errors.NewInternalError(err)
	}

	return nil
}

func (i *Implementation) GetTransactions(ctx context.Context,
	pagination primitives.Pagination) ([]domainPayment.ProviderTransaction, error) {
	res, err := i.b.GetStarTransactions(ctx, &telegramBot.GetStarTransactionsParams{
		Offset: int(pagination.Offset),
		Limit:  int(pagination.Num),
	})
	if err != nil {
		return nil, err
	}

	result := make([]domainPayment.ProviderTransaction, 0, len(res.Transactions))
	for _, r := range res.Transactions {
		result = append(result, domainPayment.ProviderTransaction{
			ProviderID: domainPayment.NewProviderID(r.ID),
			InternalID: domainPayment.New(r.Source.User.InvoicePayload),
			Date:       time.Unix(int64(r.Date), 0),
		})
	}

	return result, nil
}
