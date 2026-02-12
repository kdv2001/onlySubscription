package telegram_bot

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	telegramBot "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	"github.com/kdv2001/onlySubscription/internal/domain/primitives"
	domainProducts "github.com/kdv2001/onlySubscription/internal/domain/products"
)

func (i *Implementation) GetProducts(ctx context.Context, bot *telegramBot.Bot, update *models.Update) {
	l := strings.TrimPrefix(update.CallbackQuery.Data, productsHandler.String())
	oldMsg := update.CallbackQuery.Message.Message

	var offset int64
	if l != "" {
		var err error
		offset, err = strconv.ParseInt(l, 10, 64)
		if err != nil {
			sendErrorMsg(ctx, bot, oldMsg, err, "")
			return
		}
	}

	cardsNum := maxProductsLines * maxProductsColumns
	products, err := i.productsUseCase.GetProducts(ctx, domainProducts.RequestList{
		Filters: &domainProducts.Filters{
			ItemsExist: true,
		},
		Pagination: &primitives.Pagination{
			Num:    maxProductsLines * maxProductsColumns,
			Offset: uint64(offset) * uint64(cardsNum),
		},
	})
	if err != nil {
		sendErrorMsg(ctx, bot, oldMsg, err, "")
		return
	}

	pag := make([]*paginatorItem, 0, len(products))
	for _, curItem := range products {
		pag = append(pag, &paginatorItem{
			id:   curItem.ID.String(),
			name: curItem.Name,
		})
	}

	list := paginatorHandlerList{
		nextHandler:        productHandler.String(),
		curHandler:         productsHandler.String(),
		maxProductsColumns: maxProductsColumns,
	}

	keyboard := list.paginationKeyboard(pag, primitives.Pagination{
		Num:    uint64(cardsNum),
		Offset: uint64(offset),
	})

	keyboard = append(keyboard, []models.InlineKeyboardButton{
		{
			Text:         "Назад",
			CallbackData: productsHandler.GetBackHandler().String(),
		},
	})

	sender := msgInlineSender{
		ChatID:       update.CallbackQuery.Message.Message.Chat.ID,
		CurMessageID: update.CallbackQuery.Message.Message.ID,
		Text:         "Выбирай товар и переходи к подробному описанию.",
		Keyboard: &models.InlineKeyboardMarkup{
			InlineKeyboard: keyboard,
		},
	}

	sender.sendInlineMsg(ctx, bot)
}

func (i *Implementation) GetProduct(ctx context.Context, bot *telegramBot.Bot, update *models.Update) {
	oldMsg := update.CallbackQuery.Message.Message

	productIDStr := strings.TrimPrefix(update.CallbackQuery.Data, productHandler.String())
	if productIDStr == "" {
		sendErrorMsg(ctx, bot, oldMsg, errors.New("error parse productID"), "")
		return
	}

	productID := domainProducts.NewID(productIDStr)
	product, err := i.productsUseCase.GetProduct(ctx, productID)
	if err != nil {
		sendErrorMsg(ctx, bot, oldMsg, err, "")
		return
	}

	keyboard := [][]models.InlineKeyboardButton{
		{
			{
				Text:         "Забронировать",
				CallbackData: fmt.Sprint(createOrder, productID),
				Pay:          true,
			},
		},
		{
			{
				Text:         "Назад",
				CallbackData: productHandler.GetBackHandler().String(),
			},
		},
	}

	sender := msgInlineSender{
		ChatID:       update.CallbackQuery.Message.Message.Chat.ID,
		CurMessageID: update.CallbackQuery.Message.Message.ID,
		Text: fmt.Sprintf("Товар: %s\nописание: %s\n\n цена: %s %s",
			product.Name,
			product.Description,
			product.Price.Value,
			currencyToIcon(product.Price.Currency)),
		Keyboard: &models.InlineKeyboardMarkup{
			InlineKeyboard: keyboard,
		},
	}

	sender.sendInlineMsg(ctx, bot)
}
