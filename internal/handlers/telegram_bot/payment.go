package telegram_bot

import (
	"context"
	"errors"
	"strings"

	telegramBot "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	domainOrder "github.com/kdv2001/onlySubscription/internal/domain/order"
	domainPayment "github.com/kdv2001/onlySubscription/internal/domain/payment"
)

func (i *Implementation) CreateInvoice(ctx context.Context, bot *telegramBot.Bot, update *models.Update) {
	oldMsg := update.CallbackQuery.Message.Message

	strID := strings.TrimPrefix(update.CallbackQuery.Data, createInvoice.String())
	if strID == "" {
		sendErrorMsg(ctx, bot, oldMsg, errors.New("error parse productID"), "")
		return
	}

	orderID := domainOrder.New(strID)
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		sendErrorMsg(ctx, bot, oldMsg, errors.New("error parse productID"), "")
		return
	}

	_, err = bot.DeleteMessage(ctx, &telegramBot.DeleteMessageParams{
		ChatID:    update.CallbackQuery.Message.Message.Chat.ID,
		MessageID: update.CallbackQuery.Message.Message.ID,
	})
	if err != nil {
		sendErrorMsg(ctx, bot, oldMsg, errors.New("error parse productID"), "")
		return
	}

	invoice, err := i.paymentClient.CreateInvoice(ctx, domainPayment.CreateInvoice{
		OrderID:       orderID,
		UserID:        userID,
		PaymentMethod: domainPayment.TelegramPaymentMethod,
		TelegramData: domainPayment.TelegramData{
			ChatID: domainPayment.NewChatID(update.CallbackQuery.Message.Message.Chat.ID),
		},
	})
	if err != nil {
		sendErrorMsg(ctx, bot, oldMsg, errors.New("error parse productID"), "")
		return
	}

	_, err = bot.SendInvoice(
		ctx,
		&telegramBot.SendInvoiceParams{
			ChatID:      invoice.TelegramData.ChatID,
			Title:       invoice.Product.Title,
			Description: invoice.Product.Description,
			Payload:     invoice.ID.String(),
			Currency:    invoice.Product.Price.Currency.String(),
			Prices: []models.LabeledPrice{
				{
					Label:  "product",
					Amount: int(invoice.Product.Price.Value.IntPart()),
				},
			},
		},
	)
	if err != nil {
		sendErrorMsg(ctx, bot, oldMsg, errors.New("error parse productID"), "")
		return
	}
}
