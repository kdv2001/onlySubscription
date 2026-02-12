package telegram_bot

import (
	"context"
	"errors"
	"strings"

	telegramBot "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"go.uber.org/zap"

	domainPayment "github.com/kdv2001/onlySubscription/internal/domain/payment"
	domainUser "github.com/kdv2001/onlySubscription/internal/domain/user"
	custom_errors "github.com/kdv2001/onlySubscription/pkg/errors"
	"github.com/kdv2001/onlySubscription/pkg/logger"
)

type userUC interface {
	RegisterByTelegramID(ctx context.Context,
		telegramBotLogin domainUser.TelegramBotRegister) (domainUser.ID, error)
	GetUserByTelegramID(ctx context.Context,
		tgID domainUser.TelegramID) (domainUser.User, error)
}

type AuthMiddleware struct {
	auth userUC
}

func NewAuthMiddleware(auth userUC) *AuthMiddleware {
	return &AuthMiddleware{
		auth: auth,
	}
}

type userIDKeyStruct struct{}

var userIDKey = userIDKeyStruct{}

func getUserIDFromContext(ctx context.Context) (domainUser.ID, error) {
	id, ok := ctx.Value(userIDKey).(string)
	if !ok {
		return domainUser.ID{}, custom_errors.NewBadRequestError(errors.New("not found user id"))
	}

	return domainUser.NewID(id), nil
}

func (am *AuthMiddleware) Middleware(next telegramBot.HandlerFunc) telegramBot.HandlerFunc {
	return func(ctx context.Context, bot *telegramBot.Bot, update *models.Update) {
		if update.Message != nil && strings.Contains(update.Message.Text, "/start") {
			next(ctx, bot, update)
			return
		}

		var tgID domainUser.TelegramID
		var chatID int64
		switch {
		case update.Message != nil:
			chatID = update.Message.Chat.ID
			tgID = domainUser.NewTelegramID(update.Message.From.ID)
		case update.CallbackQuery != nil:
			chatID = update.CallbackQuery.Message.Message.Chat.ID
			tgID = domainUser.NewTelegramID(update.CallbackQuery.From.ID)
		case update.PreCheckoutQuery != nil:
			tgID = domainUser.NewTelegramID(update.PreCheckoutQuery.From.ID)
		default:
			return
		}

		user, err := am.auth.GetUserByTelegramID(ctx, tgID)
		if err != nil {
			if errors.Is(err, custom_errors.ErrorNotFound) && chatID != 0 {
				_, err = bot.SendMessage(ctx, &telegramBot.SendMessageParams{
					ChatID: chatID,
					Text: "Вы не зарегистрированы в боте." +
						" Выполните команду /start и затем повторите действия",
				})
				if err != nil {
					logger.Errorf(ctx, "error send unauthorized msg: %v", err)
					return
				}

				return
			}

			logger.Errorf(ctx, "error send unauthorized msg: %v", err)
			return
		}

		ctx = context.WithValue(ctx, userIDKey, user.ID.String())
		next(ctx, bot, update)
	}
}

// AddLoggerToContextMiddleware помещает logger в context
func AddLoggerToContextMiddleware(sugarLogger *zap.SugaredLogger) func(next telegramBot.HandlerFunc) telegramBot.HandlerFunc {
	return func(next telegramBot.HandlerFunc) telegramBot.HandlerFunc {
		fnc := telegramBot.HandlerFunc(func(ctx context.Context, bot *telegramBot.Bot, update *models.Update) {
			ctx = logger.ToContext(ctx, sugarLogger)
			next(ctx, bot, update)
		})
		return fnc
	}
}

type PaymentMiddleware struct {
	paymentClient paymentClient
}

func NewPaymentMiddleware(paymentClient paymentClient) *PaymentMiddleware {
	return &PaymentMiddleware{
		paymentClient: paymentClient,
	}
}

func (am *PaymentMiddleware) Middleware(next telegramBot.HandlerFunc) telegramBot.HandlerFunc {
	return func(ctx context.Context, bot *telegramBot.Bot, update *models.Update) {
		if update.PreCheckoutQuery != nil {
			am.preCheckout(ctx, bot, update.PreCheckoutQuery)
			return
		}
		if update.Message != nil && update.Message.SuccessfulPayment != nil {
			am.successfulPayment(ctx, bot, update.Message.SuccessfulPayment)
			return
		}

		next(ctx, bot, update)
		return
	}
}

func (am *PaymentMiddleware) preCheckout(ctx context.Context, bot *telegramBot.Bot, update *models.PreCheckoutQuery) {
	paymentID := domainPayment.New(update.InvoicePayload)
	providerID := domainPayment.NewProviderID(update.ID)

	_, err := am.paymentClient.GetInvoice(ctx, paymentID)
	if err != nil {
		_, err = bot.AnswerPreCheckoutQuery(ctx, &telegramBot.AnswerPreCheckoutQueryParams{
			PreCheckoutQueryID: "",
			OK:                 false,
			ErrorMessage:       "not found invoice",
		})
		if err != nil {
			logger.Errorf(ctx, "%v", err)
			return
		}
		return
	}

	err = am.paymentClient.Handling(ctx, paymentID, providerID)
	if err != nil {
		_, err = bot.AnswerPreCheckoutQuery(ctx, &telegramBot.AnswerPreCheckoutQueryParams{
			ErrorMessage: "not found invoice",
		})
		if err != nil {
			logger.Errorf(ctx, "%v", err)
			return
		}
		return
	}

	_, err = bot.AnswerPreCheckoutQuery(ctx, &telegramBot.AnswerPreCheckoutQueryParams{
		PreCheckoutQueryID: providerID.String(),
		OK:                 true,
	})
	if err != nil {
		logger.Errorf(ctx, "%v", err)
		return
	}

	return
}

func (am *PaymentMiddleware) successfulPayment(ctx context.Context, bot *telegramBot.Bot, update *models.SuccessfulPayment) {
	paymentID := domainPayment.New(update.InvoicePayload)
	providerID := domainPayment.NewProviderID(update.TelegramPaymentChargeID)

	err := am.paymentClient.Processing(ctx, paymentID, providerID)
	if err != nil {
		logger.Errorf(ctx, "%v", err)
		return
	}

}
