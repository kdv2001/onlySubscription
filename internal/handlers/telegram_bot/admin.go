package telegram_bot

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	telegramBot "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/shopspring/decimal"

	"github.com/kdv2001/onlySubscription/internal/domain/price"
	"github.com/kdv2001/onlySubscription/internal/domain/primitives"
	domainProducts "github.com/kdv2001/onlySubscription/internal/domain/products"
	"github.com/kdv2001/onlySubscription/pkg/form_parser"
)

func (i *Implementation) GetAdminMenu(ctx context.Context, bot *telegramBot.Bot, update *models.Update) {
	oldMsg := update.CallbackQuery.Message.Message
	results := [][]models.InlineKeyboardButton{
		{
			{
				Text:         "Создать продукт",
				CallbackData: getCreateProductInfoHandler.String(),
			},
			{
				Text:         "Просмотр продуктов",
				CallbackData: getAllProductsHandler.String(),
			},
		},
		{
			{
				Text:         "Назад",
				CallbackData: adminHandler.GetBackHandler().String(),
			},
		},
	}

	sender := msgInlineSender{
		ChatID:       oldMsg.Chat.ID,
		CurMessageID: oldMsg.ID,
		Text:         "Панель администрирования.",
		Keyboard: &models.InlineKeyboardMarkup{
			InlineKeyboard: results,
		},
	}

	sender.sendInlineMsg(ctx, bot)
}

func (i *Implementation) GetAllProducts(ctx context.Context, bot *telegramBot.Bot, update *models.Update) {
	oldMsg := update.CallbackQuery.Message.Message

	l := strings.TrimPrefix(update.CallbackQuery.Data, getAllProductsHandler.String())
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
	for _, curProduct := range products {
		pag = append(pag, &paginatorItem{
			id:   curProduct.ID.String(),
			name: curProduct.Name,
		})
	}

	paginator := paginatorHandlerList{
		nextHandler:        getProductForEditHandler.String(),
		curHandler:         productsHandler.String(),
		maxProductsColumns: maxProductsColumns,
	}

	keyboard := paginator.paginationKeyboard(pag, primitives.Pagination{
		Num:    uint64(cardsNum),
		Offset: uint64(offset),
	})

	keyboard = append(keyboard, []models.InlineKeyboardButton{
		{
			Text:         "Назад",
			CallbackData: getAllProductsHandler.GetBackHandler().String(),
		},
	})

	sender := msgInlineSender{
		ChatID:       update.CallbackQuery.Message.Message.Chat.ID,
		CurMessageID: update.CallbackQuery.Message.Message.ID,
		Text:         "Выбирай товар и переходи к его настройке.",
		Keyboard: &models.InlineKeyboardMarkup{
			InlineKeyboard: keyboard,
		},
	}

	sender.sendInlineMsg(ctx, bot)
}

func (i *Implementation) GetProductForEdit(ctx context.Context, bot *telegramBot.Bot, update *models.Update) {
	oldMsg := update.CallbackQuery.Message.Message

	separator := "?"

	splited := strings.Split(update.CallbackQuery.Data, separator)
	productIDStr := strings.TrimPrefix(splited[0], getProductForEditHandler.String())
	if productIDStr == "" {
		sendErrorMsg(ctx, bot, oldMsg, errors.New("error parse productID"), "")
		return
	}

	var offset int64
	if len(splited) > 1 {
		var err error
		offset, err = strconv.ParseInt(splited[1], 10, 64)
		if err != nil {
			sendErrorMsg(ctx, bot, oldMsg, err, "")
			return
		}
	}

	productID := domainProducts.NewID(productIDStr)
	product, err := i.productsUseCase.GetProduct(ctx, productID)
	if err != nil {
		sendErrorMsg(ctx, bot, oldMsg, err, "")
		return
	}

	cardsNum := maxProductsLines * maxProductsColumns
	cardOffset := offset * int64(cardsNum)
	items, err := i.productsUseCase.GetItems(ctx, domainProducts.RequestList{
		Filters: &domainProducts.Filters{
			ProductID: productID,
		},
		Pagination: &primitives.Pagination{
			Num:    maxProductsLines * maxProductsColumns,
			Offset: uint64(cardOffset),
		},
	})
	if err != nil {
		sendErrorMsg(ctx, bot, oldMsg, err, "")
		return
	}

	pag := make([]*paginatorItem, 0, len(items))
	for j, curItem := range items {
		pag = append(pag, &paginatorItem{
			id:   curItem.ID.String(),
			name: fmt.Sprint(int64(j) + cardOffset + 1),
		})
	}

	paginator := paginatorHandlerList{
		nextHandler:        getItemHandler.String(),
		curHandler:         getProductForEditHandler.String() + productID.String() + separator,
		maxProductsColumns: maxProductsColumns,
	}

	keyboard := paginator.paginationKeyboard(pag, primitives.Pagination{
		Num:    uint64(cardsNum),
		Offset: uint64(offset),
	})

	keyboard = append(keyboard, [][]models.InlineKeyboardButton{
		{
			{
				Text:         "Добавить item",
				CallbackData: fmt.Sprint(getCreateItemInfoHandler, productID),
			},
		},
		{
			{
				Text:         "Удалить продукт",
				CallbackData: fmt.Sprint(deleteProductHandler, productID),
			},
			{
				Text:         "Назад",
				CallbackData: getProductForEditHandler.GetBackHandler().String(),
			},
		},
	}...)

	sender := msgInlineSender{
		ChatID:       update.CallbackQuery.Message.Message.Chat.ID,
		CurMessageID: update.CallbackQuery.Message.Message.ID,
		Text: fmt.Sprintf("ID: %s\nИзображение: %s\nТип: %s\nТовар: %s\nОписание: %s\n\n Цена: %s %s",
			product.ID,
			product.Image.URL,
			product.Type.String(),
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

func (i *Implementation) GetCreateProductInfo(ctx context.Context, bot *telegramBot.Bot, update *models.Update) {
	str := make([]string, 0)
	msg := productFormParser.Format(str)

	sender := msgInlineSender{
		ChatID:       update.CallbackQuery.Message.Message.Chat.ID,
		CurMessageID: update.CallbackQuery.Message.Message.ID,
		Text:         "/" + createProductHandler.String() + "\n" + strings.Join(msg, ",\n"),
		Keyboard: &models.InlineKeyboardMarkup{
			InlineKeyboard: [][]models.InlineKeyboardButton{
				{
					{
						Text:         "Назад",
						CallbackData: getCreateProductInfoHandler.GetBackHandler().String(),
					},
				},
			},
		},
	}

	sender.sendInlineMsg(ctx, bot)
}

func (i *Implementation) CreateProduct(ctx context.Context, bot *telegramBot.Bot, update *models.Update) {
	oldMsg := update.Message

	inputMessage := strings.TrimPrefix(update.Message.Text, "/"+createProductHandler.String())
	inputMessage = strings.TrimSpace(inputMessage)
	var p = &product{}
	msg, err := productFormParser.Execute(inputMessage, p)
	if err != nil {
		sendErrorMsg(ctx, bot, oldMsg, err, "")
		return
	}
	if msg != nil {
		_, err := bot.SendMessage(ctx, &telegramBot.SendMessageParams{
			Text: "Заполните поле: " + msg.Field + " и повторите снова",
			ReplyMarkup: &models.InlineKeyboardMarkup{
				InlineKeyboard: [][]models.InlineKeyboardButton{
					{
						{
							Text:         "Отмена",
							CallbackData: backHandler.String() + menu.String(),
						},
					},
				},
			},
		})
		if err != nil {
			return
		}
	}

	d, err := decimal.NewFromString(p.Price)
	if err != nil {
		sendErrorMsg(ctx, bot, oldMsg, err, "")
		return
	}

	err = i.productsUseCase.CreateProduct(ctx, domainProducts.Product{
		Type:        domainProducts.TypeFromString(p.Type),
		Name:        p.Name,
		Description: p.Description,
		Image: domainProducts.Image{
			URL: p.Image,
		},
		Price: price.Price{
			Currency: price.CurrencyFromString(p.Currency),
			Value:    d,
		},
		SubscriptionPeriod: 0,
	})
	if err != nil {
		sendErrorMsg(ctx, bot, oldMsg, err, "")
		return
	}

	sender := msgInlineSender{
		ChatID:       oldMsg.Chat.ID,
		CurMessageID: oldMsg.ID,
		Text:         "Продукт успешно создан",
		Keyboard: &models.InlineKeyboardMarkup{
			InlineKeyboard: [][]models.InlineKeyboardButton{
				{
					{
						Text:         "Назад",
						CallbackData: createProductHandler.GetBackHandler().String(),
					},
				},
			},
		},
	}

	sender.sendInlineMsg(ctx, bot)
}

var productFormParser = newProductParser()

func newProductParser() form_parser.FormParser {
	defaultPlaceholder := "<значение>"
	stg := form_parser.NewStage("Название", "name", defaultPlaceholder)
	stg.SetNext(form_parser.NewStage("Описание", "desc", defaultPlaceholder)).
		SetNext(form_parser.NewStage("Изображение", "image", defaultPlaceholder)).
		SetNext(form_parser.NewStage("Стоимость", "price", defaultPlaceholder)).
		SetNext(form_parser.NewStage("Валюта", "currency", defaultPlaceholder)).
		SetNext(form_parser.NewStage("Период подписки", "subscription_period", defaultPlaceholder)).
		SetNext(form_parser.NewStage("Тип продукта", "type", defaultPlaceholder))
	return stg
}

type product struct {
	Name               string `field:"name"`
	Type               string `field:"type"`
	Description        string `field:"desc"`
	Image              string `field:"image"`
	Currency           string `field:"currency"`
	Price              string `field:"price"`
	SubscriptionPeriod string `field:"subscription_period"`
}

var productItemParser = newItemParser()

func newItemParser() form_parser.FormParser {
	defaultPlaceholder := "<значение>"
	stg := form_parser.NewStage("ID продукта", "id", defaultPlaceholder, true)
	stg.SetNext(form_parser.NewStage("Полезная нагрузка", "payload", defaultPlaceholder))

	return stg
}

type item struct {
	ProductID string `field:"id"`
	Payload   string `field:"payload"`
}

func (i *Implementation) DeleteProduct(ctx context.Context, bot *telegramBot.Bot, update *models.Update) {
	oldMsg := update.CallbackQuery.Message.Message

	productIDStr := strings.TrimPrefix(update.CallbackQuery.Data, deleteProductHandler.String())
	productID := domainProducts.NewID(productIDStr)

	err := i.productsUseCase.DeactivateProduct(ctx, productID)
	if err != nil {
		sendErrorMsg(ctx, bot, oldMsg, err, "")
		return
	}

	buttons := [][]models.InlineKeyboardButton{
		{
			{
				Text:         "Назад",
				CallbackData: deleteProductHandler.GetBackHandler().String(),
			},
		},
	}

	_, err = bot.SendMessage(ctx, &telegramBot.SendMessageParams{
		ChatID: update.CallbackQuery.Message.Message.Chat.ID,
		Text: fmt.Sprintf("Продукт деактивирован\nID: %s",
			productID.ID),
		ReplyMarkup: &models.InlineKeyboardMarkup{
			InlineKeyboard: buttons,
		},
	})
	if err != nil {
		return
	}
}

func (i *Implementation) GetCreateItemInfo(ctx context.Context, bot *telegramBot.Bot, update *models.Update) {
	strID := strings.TrimPrefix(update.CallbackQuery.Data, getCreateItemInfoHandler.String())
	if strID == "" {
		return
	}

	_, err := bot.DeleteMessage(ctx, &telegramBot.DeleteMessageParams{
		ChatID:    update.CallbackQuery.Message.Message.Chat.ID,
		MessageID: update.CallbackQuery.Message.Message.ID,
	})

	str := make([]string, 0)
	msg := productItemParser.Format(str)

	_, err = bot.SendMessage(ctx, &telegramBot.SendMessageParams{
		ChatID: update.CallbackQuery.Message.Message.Chat.ID,
		Text:   fmt.Sprintf("/%s\nID продукта: %s\n%s", createItemHandler.String(), strID, strings.Join(msg, ",\n")),
		ReplyMarkup: &models.InlineKeyboardMarkup{
			InlineKeyboard: [][]models.InlineKeyboardButton{
				{
					{
						Text:         "Назад",
						CallbackData: getCreateItemInfoHandler.GetBackHandler().String() + strID,
					},
				},
			},
		},
	})
	if err != nil {
		return
	}
}

func (i *Implementation) AddItem(ctx context.Context, bot *telegramBot.Bot, update *models.Update) {
	oldMsg := update.Message
	inputMessage := strings.TrimPrefix(update.Message.Text, "/"+createItemHandler.String())
	inputMessage = strings.TrimSpace(inputMessage)
	var p = &item{}
	msg, err := productItemParser.Execute(inputMessage, p)
	if err != nil {
		return
	}

	if msg != nil {
		_, err := bot.SendMessage(ctx, &telegramBot.SendMessageParams{
			Text: "Заполните поле: " + msg.Field + " и повторите снова",
			ReplyMarkup: &models.InlineKeyboardMarkup{
				InlineKeyboard: [][]models.InlineKeyboardButton{
					{
						{
							Text:         "Отмена",
							CallbackData: createItemHandler.GetBackHandler().String(),
						},
					},
				},
			},
		})
		if err != nil {
			return
		}
	}

	err = i.productsUseCase.AddItem(ctx, domainProducts.Item{
		ProductID: domainProducts.NewID(p.ProductID),
		Payload:   p.Payload,
	})
	if err != nil {
		sendErrorMsg(ctx, bot, oldMsg, err, "")
		return
	}

	sender := msgInlineSender{
		ChatID:       oldMsg.Chat.ID,
		CurMessageID: oldMsg.ID,
		Text:         "Единица товара успешно создана",
		Keyboard: &models.InlineKeyboardMarkup{
			InlineKeyboard: [][]models.InlineKeyboardButton{
				{
					{
						Text:         "Администрирование",
						CallbackData: adminHandler.String(),
					},
				},
			},
		},
	}

	sender.sendInlineMsg(ctx, bot)
}

func (i *Implementation) DeleteItem(ctx context.Context, bot *telegramBot.Bot, update *models.Update) {
	oldMsg := update.CallbackQuery.Message.Message

	itemIDStr := strings.TrimPrefix(update.CallbackQuery.Data, deleteItemHandler.String())
	itemID := domainProducts.NewItemID(itemIDStr)

	curItem, err := i.productsUseCase.GetItem(ctx, itemID)
	if err != nil {
		sendErrorMsg(ctx, bot, oldMsg, err, "")
		return
	}

	err = i.productsUseCase.DeleteItem(ctx, itemID)
	if err != nil {
		sendErrorMsg(ctx, bot, oldMsg, err, "")
		return
	}

	buttons := [][]models.InlineKeyboardButton{
		{
			{
				Text:         "Назад",
				CallbackData: deleteItemHandler.GetBackHandler().GetBackHandler().String() + curItem.ProductID.String(),
			},
		},
	}

	sender := msgInlineSender{
		ChatID:       update.CallbackQuery.Message.Message.Chat.ID,
		CurMessageID: update.CallbackQuery.Message.Message.ID,
		Text: fmt.Sprintf("Удален\nID: %s\nПолезная нагрузка: %s\n",
			curItem.ID,
			curItem.Payload),
		Keyboard: &models.InlineKeyboardMarkup{
			InlineKeyboard: buttons,
		},
	}

	sender.sendInlineMsg(ctx, bot)
}

func (i *Implementation) GetItem(ctx context.Context, bot *telegramBot.Bot, update *models.Update) {
	oldMsg := update.CallbackQuery.Message.Message

	itemIDStr := strings.TrimPrefix(update.CallbackQuery.Data, getItemHandler.String())
	itemID := domainProducts.NewItemID(itemIDStr)

	curItem, err := i.productsUseCase.GetItem(ctx, itemID)
	if err != nil {
		sendErrorMsg(ctx, bot, oldMsg, err, "")
		return
	}

	buttons := [][]models.InlineKeyboardButton{
		{
			{
				Text:         "Удалить item",
				CallbackData: fmt.Sprint(deleteItemHandler, curItem.ID),
			},
			{
				Text:         "Назад",
				CallbackData: getItemHandler.GetBackHandler().String() + curItem.ProductID.String(),
			},
		},
	}

	sender := msgInlineSender{
		ChatID:       update.CallbackQuery.Message.Message.Chat.ID,
		CurMessageID: update.CallbackQuery.Message.Message.ID,
		Text: fmt.Sprintf("ID: %s\nПолезная нагрузка: %s\n",
			curItem.ID,
			curItem.Payload),
		Keyboard: &models.InlineKeyboardMarkup{
			InlineKeyboard: buttons,
		},
	}

	sender.sendInlineMsg(ctx, bot)
}
