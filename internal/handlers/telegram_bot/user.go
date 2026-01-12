package telegram_bot

import (
	"context"
	"fmt"

	telegramBot "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func (i *Implementation) GetProfile(ctx context.Context, bot *telegramBot.Bot, update *models.Update) {
	oldMsg := update.CallbackQuery.Message.Message
	results := [][]models.InlineKeyboardButton{
		{
			{
				Text:         "Список заказов",
				CallbackData: getOrderListHandler.String(),
			},
		},
		{
			{
				Text:         "Назад",
				CallbackData: menu.GetBackHandler().String(),
			},
		},
	}

	sender := msgInlineSender{
		ChatID:       oldMsg.Chat.ID,
		CurMessageID: oldMsg.ID,
		Text:         fmt.Sprintf("Твой userID: %d", oldMsg.From.ID),
		Keyboard: &models.InlineKeyboardMarkup{
			InlineKeyboard: results,
		},
	}

	sender.sendInlineMsg(ctx, bot)
}

func (i *Implementation) GetHelp(ctx context.Context, bot *telegramBot.Bot, update *models.Update) {
	oldMsg := update.CallbackQuery.Message.Message
	results := [][]models.InlineKeyboardButton{
		{
			{
				Text:         "Назад",
				CallbackData: menu.GetBackHandler().String(),
			},
		},
	}

	sender := msgInlineSender{
		ChatID:       oldMsg.Chat.ID,
		CurMessageID: oldMsg.ID,
		Text:         "Подробное описание магазина",
		Keyboard: &models.InlineKeyboardMarkup{
			InlineKeyboard: results,
		},
	}

	sender.sendInlineMsg(ctx, bot)
}
