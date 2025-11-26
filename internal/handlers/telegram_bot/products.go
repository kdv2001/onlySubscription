package telegram_bot

import (
	"context"
	"fmt"
	"strings"

	telegramBot "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	"github.com/kdv2001/onlySubscription/internal/domain/price"
	"github.com/kdv2001/onlySubscription/internal/domain/primitives"
	domainProducts "github.com/kdv2001/onlySubscription/internal/domain/products"
)

type productsUseCase interface {
	GetProducts(ctx context.Context, req primitives.RequestList) (domainProducts.Products, error)
	GetProduct(ctx context.Context, id domainProducts.ID) (domainProducts.Product, error)
	CreateProduct(ctx context.Context, product domainProducts.Product) error
}
type Implementation struct {
	productsUseCase productsUseCase
	// TODO посмотреть методологии написания юнит-тестов
	bot *telegramBot.Bot
}

const (
	maxLines   = 5
	maxColumns = 2
)

func NewImplementation(
	productsUseCase productsUseCase,
	bot *telegramBot.Bot,
) *Implementation {
	i := &Implementation{
		productsUseCase: productsUseCase,
		bot:             bot,
	}

	bot.RegisterHandler(telegramBot.HandlerTypeCallbackQueryData,
		"products", telegramBot.MatchTypePrefix, i.GetProducts)
	bot.RegisterHandler(telegramBot.HandlerTypeCallbackQueryData,
		"product_", telegramBot.MatchTypePrefix, i.GetProduct)
	bot.RegisterHandler(telegramBot.HandlerTypeCallbackQueryData,
		"back_", telegramBot.MatchTypePrefix, i.Back)
	bot.RegisterHandler(telegramBot.HandlerTypeMessageText,
		"menu", telegramBot.MatchTypeCommand, i.GetMenu)
	bot.RegisterHandler(telegramBot.HandlerTypeCallbackQueryData,
		"admin", telegramBot.MatchTypePrefix, i.GetAdminMenu)

	bot.RegisterHandler(telegramBot.HandlerTypeCallbackQueryData,
		"admin", telegramBot.MatchTypePrefix, i.GetAdminMenu)

	bot.RegisterHandler(telegramBot.HandlerTypeCallbackQueryData,
		"create_product", telegramBot.MatchTypePrefix, i.CreateProduct)
	bot.RegisterHandler(telegramBot.HandlerTypeCallbackQueryData,
		"delete_product", telegramBot.MatchTypePrefix, i.DeleteProduct)
	bot.RegisterHandler(telegramBot.HandlerTypeCallbackQueryData,
		"add_item", telegramBot.MatchTypePrefix, i.AddItem)
	bot.RegisterHandler(telegramBot.HandlerTypeCallbackQueryData,
		"delete_item", telegramBot.MatchTypePrefix, i.DeleteItem)

	return i
}

type createProductJSON struct {
	ID          string
	Type        string
	Name        string
	Description string
	Image       string
	Price       priceJson
}

type priceJson struct {
}

func (i *Implementation) CreateProduct(ctx context.Context, bot *telegramBot.Bot, update *models.Update) {

}

func (i *Implementation) DeleteProduct(ctx context.Context, bot *telegramBot.Bot, update *models.Update) {
}

func (i *Implementation) AddItem(ctx context.Context, bot *telegramBot.Bot, update *models.Update) {

}

func (i *Implementation) DeleteItem(ctx context.Context, bot *telegramBot.Bot, update *models.Update) {
}

func (i *Implementation) Start(ctx context.Context, bot *telegramBot.Bot, update *models.Update) {
}

func (i *Implementation) GetMenu(ctx context.Context, bot *telegramBot.Bot, update *models.Update) {
	results := [][]models.InlineKeyboardButton{
		{
			{
				Text:         "Продукты",
				CallbackData: "products",
			},
			{
				Text:         "Профиль",
				CallbackData: "profile",
			},
		},
		{
			{
				Text:         "Помощь",
				CallbackData: "Help",
			},
		},

		{
			{
				Text:         "Администрирование",
				CallbackData: "admin",
			},
		},
	}

	var chatID int64
	if update.CallbackQuery != nil {
		chatID = update.CallbackQuery.Message.Message.Chat.ID
	} else {
		chatID = update.Message.Chat.ID
	}

	_, err := bot.SendMessage(ctx, &telegramBot.SendMessageParams{
		ChatID: chatID,
		Text:   "Привет, друг! Добро пожаловать в мой магазин!",
		ReplyMarkup: &models.InlineKeyboardMarkup{
			InlineKeyboard: results,
		},
	})
	if err != nil {
		return
	}
}

func (i *Implementation) GetAdminMenu(ctx context.Context, bot *telegramBot.Bot, update *models.Update) {
	results := [][]models.InlineKeyboardButton{
		{
			{
				Text:         "Создать продукт",
				CallbackData: "create_product",
			},
			{
				Text:         "Удалить продукт",
				CallbackData: "delete_product",
			},
		},
		{
			{
				Text:         "Добавить элемент в инвентарь",
				CallbackData: "add_item",
			},
			{
				Text:         "Удалить элемент из инвентаря",
				CallbackData: "delete_item",
			},
		},
	}

	var chatID int64
	if update.CallbackQuery != nil {
		chatID = update.CallbackQuery.Message.Message.Chat.ID
	} else {
		chatID = update.Message.Chat.ID
	}

	_, err := bot.DeleteMessage(ctx, &telegramBot.DeleteMessageParams{
		ChatID:    update.CallbackQuery.Message.Message.Chat.ID,
		MessageID: update.CallbackQuery.Message.Message.ID,
	})
	if err != nil {
		return
	}

	_, err = bot.SendMessage(ctx, &telegramBot.SendMessageParams{
		ChatID: chatID,
		Text:   "Привет, друг! Добро пожаловать в мой магазин!",
		ReplyMarkup: &models.InlineKeyboardMarkup{
			InlineKeyboard: results,
		},
	})
	if err != nil {
		return
	}
}

func (i *Implementation) Back(ctx context.Context, bot *telegramBot.Bot, update *models.Update) {
	back := strings.TrimPrefix(update.CallbackQuery.Data, "back_")

	switch back {
	case "menu":
		_, err := bot.DeleteMessage(ctx, &telegramBot.DeleteMessageParams{
			ChatID:    update.CallbackQuery.Message.Message.Chat.ID,
			MessageID: update.CallbackQuery.Message.Message.ID,
		})
		if err != nil {
			return
		}
		i.GetMenu(ctx, bot, update)
	case "products":
		i.GetProducts(ctx, bot, update)
	}

}

func (i *Implementation) GetProducts(ctx context.Context, bot *telegramBot.Bot, update *models.Update) {
	products, err := i.productsUseCase.GetProducts(ctx, primitives.RequestList{
		Pagination: primitives.Pagination{
			Num:    maxLines * maxColumns,
			Offset: 0,
		},
	})
	if err != nil {
		return
	}

	results := make([][]models.InlineKeyboardButton, 0, len(products))
	for j := 0; j < len(products)/maxColumns; j++ {
		m := make([]models.InlineKeyboardButton, 0, maxColumns)
		for k := 0; k < maxColumns; k++ {
			pr := products[j*maxColumns+k]
			m = append(m, models.InlineKeyboardButton{
				Text:         pr.Name,
				CallbackData: fmt.Sprintf("product_%s", pr.ID),
			})
		}
		results = append(results, m)
	}

	m := make([]models.InlineKeyboardButton, 0, len(products)%2)
	for j := len(products) % maxColumns; j > 0; j-- {
		pr := products[len(products)-j]
		m = append(m, models.InlineKeyboardButton{
			Text:         pr.Name,
			CallbackData: fmt.Sprintf("product_%s", pr.ID),
		})
	}
	results = append(results, m)

	// TODO добавить пагинацию

	results = append(results, []models.InlineKeyboardButton{
		{
			Text:         "Назад",
			CallbackData: "back_menu",
		},
	})

	_, err = bot.DeleteMessage(ctx, &telegramBot.DeleteMessageParams{
		ChatID:    update.CallbackQuery.Message.Message.Chat.ID,
		MessageID: update.CallbackQuery.Message.Message.ID,
	})
	if err != nil {
		return
	}

	_, err = bot.SendMessage(ctx, &telegramBot.SendMessageParams{
		ChatID: update.CallbackQuery.Message.Message.Chat.ID,
		Text:   "Выбирай товар и переходи к подробному описанию.",
		ReplyMarkup: &models.InlineKeyboardMarkup{
			InlineKeyboard: results,
		},
	})
	if err != nil {
		return
	}
}

func (i *Implementation) GetProduct(ctx context.Context, bot *telegramBot.Bot, update *models.Update) {
	strID := strings.TrimPrefix(update.CallbackQuery.Data, "product_")
	if strID == "" {
		return
	}

	id := domainProducts.NewID(strID)
	product, err := i.productsUseCase.GetProduct(ctx, id)
	if err != nil {
		return
	}

	results := [][]models.InlineKeyboardButton{
		{
			{
				Text:         "Купить",
				CallbackData: fmt.Sprintf("pay_product_%s", id),
				Pay:          true,
			},
		},
		{
			{
				Text:         "Назад",
				CallbackData: "back_products",
			},
		},
	}

	_, err = bot.DeleteMessage(ctx, &telegramBot.DeleteMessageParams{
		ChatID:    update.CallbackQuery.Message.Message.Chat.ID,
		MessageID: update.CallbackQuery.Message.Message.ID,
	})
	if err != nil {
		return
	}

	_, err = bot.SendMessage(ctx, &telegramBot.SendMessageParams{
		ChatID: update.CallbackQuery.Message.Message.Chat.ID,
		Text: fmt.Sprintf("Товар: %s\nописание: %s\n\n цена: %s %s",
			product.Name,
			product.Description,
			product.Price.Value,
			currencyToIcon(product.Price.Currency)),
		ReplyMarkup: &models.InlineKeyboardMarkup{
			InlineKeyboard: results,
		},
	})
	if err != nil {
		return
	}
}

func currencyToIcon(c price.Currency) string {
	switch c {
	case price.XTR:
		return string('⭐')
	}

	return string('⚙')
}
