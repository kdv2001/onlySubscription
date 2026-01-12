package telegram_bot

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	telegramBot "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	domainOrder "github.com/kdv2001/onlySubscription/internal/domain/order"
	"github.com/kdv2001/onlySubscription/internal/domain/primitives"
	domainProducts "github.com/kdv2001/onlySubscription/internal/domain/products"
)

func (i *Implementation) CreateOrder(ctx context.Context, bot *telegramBot.Bot, update *models.Update) {
	oldMsg := update.CallbackQuery.Message.Message
	strID := strings.TrimPrefix(update.CallbackQuery.Data, createOrder.String())
	if strID == "" {
		return
	}

	appUserID, err := getUserIDFromContext(ctx)
	if err != nil {
		sendErrorMsg(ctx, bot, oldMsg, err, "")
		return
	}

	productID := domainProducts.NewID(strID)
	orderID, err := i.orderUseCase.CreateOrder(ctx, domainOrder.CreateOrder{
		UserID:    appUserID,
		ProductID: domainProducts.NewID(strID),
	})
	if err != nil {
		sendErrorMsg(ctx, bot, oldMsg, err, "")
		return
	}

	_, err = i.orderUseCase.GetOrder(ctx, orderID, appUserID)
	if err != nil {
		sendErrorMsg(ctx, bot, oldMsg, err, "")
		return
	}

	keyboard := [][]models.InlineKeyboardButton{
		{
			{
				Text:         "Оплатить⭐️",
				CallbackData: fmt.Sprint(createInvoice.String(), orderID),
			},
		},
		{
			{
				Text:         "Назад",
				CallbackData: createOrder.GetBackHandler().String() + productID.String(),
			},
		},
	}

	sender := msgInlineSender{
		ChatID:       oldMsg.Chat.ID,
		CurMessageID: oldMsg.ID,
		Text:         "Заказ создан",
		Keyboard: &models.InlineKeyboardMarkup{
			InlineKeyboard: keyboard,
		},
	}

	sender.sendInlineMsg(ctx, bot)
}

func (i *Implementation) GetOrder(ctx context.Context, bot *telegramBot.Bot, update *models.Update) {
	oldMsg := update.CallbackQuery.Message.Message
	strID := strings.TrimPrefix(update.CallbackQuery.Data, getOrderHandler.String())
	if strID == "" {
		sendErrorMsg(ctx, bot, oldMsg, errors.New("empty order id"), "")
		return
	}

	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		sendErrorMsg(ctx, bot, oldMsg, err, "")
		return
	}

	order, err := i.orderUseCase.GetOrder(ctx, domainOrder.New(strID), userID)
	if err != nil {
		sendErrorMsg(ctx, bot, oldMsg, err, "")
		return
	}

	var keyboard [][]models.InlineKeyboardButton
	if order.Status == domainOrder.ExpectPayments {
		keyboard = [][]models.InlineKeyboardButton{
			{
				{
					Text:         "Оплатить",
					CallbackData: fmt.Sprint(createInvoice.String(), order.ID),
					Pay:          true,
				},
			},
		}
	}

	keyboard = append(keyboard, []models.InlineKeyboardButton{
		{
			Text:         "Назад",
			CallbackData: getOrderHandler.GetBackHandler().String(),
		},
	})

	sender := msgInlineSender{
		ChatID:       oldMsg.Chat.ID,
		CurMessageID: oldMsg.ID,
		Text: fmt.Sprintf("Заказ: %s\nописание: %s\n\n цена: %s %s \nпродукт:%s",
			order.ID,
			order.Product.ProductID,
			order.TotalPrice.Value,
			order.Product.Title,
			currencyToIcon(order.TotalPrice.Currency)),
		Keyboard: &models.InlineKeyboardMarkup{
			InlineKeyboard: keyboard,
		},
	}
	sender.sendInlineMsg(ctx, bot)
}

const (
	maxOrdersLines   = 5
	maxOrdersColumns = 2
)

func (i *Implementation) OrderList(ctx context.Context, bot *telegramBot.Bot, update *models.Update) {
	offsetStr := strings.TrimPrefix(update.CallbackQuery.Data, getOrderListHandler.String())
	oldMsg := update.CallbackQuery.Message.Message

	var offset int64
	if offsetStr != "" {
		var err error
		offset, err = strconv.ParseInt(offsetStr, 10, 64)
		if err != nil {
			sendErrorMsg(ctx, bot, oldMsg, err, "")
			return
		}
	}

	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		sendErrorMsg(ctx, bot, oldMsg, err, "")
		return
	}

	cardsNum := maxOrdersLines * maxOrdersColumns
	orders, err := i.orderUseCase.GetOrderList(ctx, userID, domainOrder.RequestList{
		Pagination: &primitives.Pagination{
			Num:    uint64(cardsNum),
			Offset: uint64(offset) * uint64(cardsNum),
		},
		Filters: &domainOrder.Filters{
			UserID: userID,
		},
		Sort: &domainOrder.Sort{
			CreatedAt: primitives.Descending,
		},
	})
	if err != nil {
		sendErrorMsg(ctx, bot, oldMsg, err, "")
		return
	}

	pag := make([]*paginatorItem, 0, len(orders))
	for _, o := range orders {
		pag = append(pag, &paginatorItem{
			id:   o.ID.String(),
			name: o.Product.Title,
		})
	}

	list := paginatorHandlerList{
		nextHandler:        getOrderHandler.String(),
		curHandler:         getOrderListHandler.String(),
		maxProductsColumns: maxOrdersColumns,
	}

	keyboard := list.paginationKeyboard(pag, primitives.Pagination{
		Num:    uint64(cardsNum),
		Offset: uint64(offset),
	})

	keyboard = append(keyboard, []models.InlineKeyboardButton{
		{
			Text:         "Назад",
			CallbackData: profileHandler.GetBackHandler().String(),
		},
	})

	sender := msgInlineSender{
		ChatID:       update.CallbackQuery.Message.Message.Chat.ID,
		CurMessageID: update.CallbackQuery.Message.Message.ID,
		Text:         "Страница заказов.",
		Keyboard: &models.InlineKeyboardMarkup{
			InlineKeyboard: keyboard,
		},
	}

	sender.sendInlineMsg(ctx, bot)
}
