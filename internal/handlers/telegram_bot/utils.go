package telegram_bot

import (
	"context"
	"fmt"

	telegramBot "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/google/uuid"

	"github.com/kdv2001/onlySubscription/internal/domain/primitives"
	"github.com/kdv2001/onlySubscription/pkg/logger"
)

type msgInlineSender struct {
	ChatID       int64
	CurMessageID int
	Text         string
	Keyboard     *models.InlineKeyboardMarkup
}

func (msg *msgInlineSender) sendInlineMsg(ctx context.Context, bot *telegramBot.Bot) {
	if msg.CurMessageID != 0 {
		_, err := bot.DeleteMessage(ctx, &telegramBot.DeleteMessageParams{
			ChatID:    msg.ChatID,
			MessageID: msg.CurMessageID,
		})
		if err != nil {
			logger.Errorf(ctx, "error: %v", err)
			return
		}
	}

	_, err := bot.SendMessage(ctx, &telegramBot.SendMessageParams{
		ChatID:      msg.ChatID,
		Text:        msg.Text,
		ReplyMarkup: msg.Keyboard,
	})
	if err != nil {
		logger.Errorf(ctx, "error: %v", err)
		return
	}
}

type paginatorHandlerList struct {
	nextHandler        string
	curHandler         string
	maxProductsColumns int
}

type paginatorItem struct {
	id   string
	name string
}

func (p *paginatorItem) GetID() string {
	return p.id
}

func (p *paginatorItem) GetName() string {
	return p.name
}

func (p *paginatorHandlerList) paginationKeyboard(products []*paginatorItem,
	pagination primitives.Pagination) [][]models.InlineKeyboardButton {
	offset := pagination.Offset
	cardsNum := pagination.Num

	results := make([][]models.InlineKeyboardButton, 0, len(products))
	for j := 0; j < len(products)/p.maxProductsColumns; j++ {
		m := make([]models.InlineKeyboardButton, 0, p.maxProductsColumns)
		for k := 0; k < p.maxProductsColumns; k++ {
			pr := products[j*p.maxProductsColumns+k]
			m = append(m, models.InlineKeyboardButton{
				Text:         pr.GetName(),
				CallbackData: fmt.Sprint(p.nextHandler, pr.GetID()),
			})
		}
		results = append(results, m)
	}

	m := make([]models.InlineKeyboardButton, 0, len(products)%2)
	for j := len(products) % p.maxProductsColumns; j > 0; j-- {
		pr := products[len(products)-j]
		m = append(m, models.InlineKeyboardButton{
			Text:         pr.GetName(),
			CallbackData: fmt.Sprint(p.nextHandler, pr.GetID()),
		})
	}
	if len(m) > 0 {
		results = append(results, m)
	}

	paginationButtons := make([]models.InlineKeyboardButton, 0, 2)
	if offset != 0 {
		paginationButtons = append(paginationButtons,
			models.InlineKeyboardButton{
				Text:         "Пред.",
				CallbackData: p.curHandler + fmt.Sprint(offset-1),
			})
	}

	if uint64(len(products)) == cardsNum {
		paginationButtons = append(paginationButtons,
			models.InlineKeyboardButton{
				Text:         "След.",
				CallbackData: p.curHandler + fmt.Sprint(offset+1),
			})
	}

	if len(paginationButtons) > 0 {
		results = append(results, paginationButtons)
	}

	return results
}

const defaultError = "Что-то пошло не так.\nПожалуйста, попробуйте снова и сообщите об этом администратору.\n\nКода ошибки: %s"

func sendErrorMsg(ctx context.Context,
	bot *telegramBot.Bot,
	msg *models.Message,
	err error,
	text string,
) {
	errorUUID := uuid.New()
	if err != nil {
		logger.Errorf(ctx, "uuid: %s; error: %v", errorUUID.String(), err)
	}

	if text == "" {
		text = fmt.Sprintf(defaultError, errorUUID)
	}

	_, err = bot.SendMessage(ctx, &telegramBot.SendMessageParams{
		ChatID: msg.Chat.ID,
		Text:   text,
		ReplyMarkup: &models.InlineKeyboardMarkup{
			InlineKeyboard: [][]models.InlineKeyboardButton{
				{
					{
						Text:         "Меню",
						CallbackData: menu.String(),
					},
				},
			},
		},
	})
	if err != nil {
		logger.Errorf(ctx, "uuid: %s; error: %v", errorUUID.String(), err)
	}
}
