package telegram_bot

import (
	"context"

	telegramBot "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	domainOrder "github.com/kdv2001/onlySubscription/internal/domain/order"
	domainPayment "github.com/kdv2001/onlySubscription/internal/domain/payment"
	domainProducts "github.com/kdv2001/onlySubscription/internal/domain/products"
	domainUser "github.com/kdv2001/onlySubscription/internal/domain/user"
	"github.com/kdv2001/onlySubscription/pkg/logger"
)

type productsUseCase interface {
	GetProducts(ctx context.Context, req domainProducts.RequestList) (domainProducts.Products, error)
	GetProduct(ctx context.Context, id domainProducts.ID) (domainProducts.Product, error)
	CreateProduct(ctx context.Context, product domainProducts.Product) error
	DeactivateProduct(ctx context.Context, id domainProducts.ID) error

	AddItem(ctx context.Context, item domainProducts.Item) error
	GetItems(ctx context.Context, req domainProducts.RequestList) ([]domainProducts.Item, error)
	GetItem(ctx context.Context, id domainProducts.ItemID) (domainProducts.Item, error)
	DeleteItem(ctx context.Context, id domainProducts.ItemID) error
}

type userUseCase interface {
	RegisterByTelegramID(ctx context.Context,
		telegramBotLogin domainUser.TelegramBotRegister) (domainUser.ID, error)
	GetUserByTelegramID(ctx context.Context,
		tgID domainUser.TelegramID) (domainUser.User, error)
}

type orderUseCase interface {
	CreateOrder(ctx context.Context, o domainOrder.CreateOrder) (domainOrder.ID, error)
	GetOrder(ctx context.Context, oID domainOrder.ID, userID domainUser.ID) (domainOrder.Order, error)
	GetOrderList(ctx context.Context, userID domainUser.ID, list domainOrder.RequestList) ([]domainOrder.Order, error)
}

type paymentClient interface {
	CreateInvoice(ctx context.Context,
		invoice domainPayment.CreateInvoice,
	) (domainPayment.ReleaseInvoice, error)
	GetInvoice(ctx context.Context,
		id domainPayment.ID,
	) (domainPayment.Invoice, error)
	Handling(ctx context.Context,
		id domainPayment.ID,
		providerID domainPayment.ProviderID,
	) error
	Processing(ctx context.Context,
		id domainPayment.ID,
		providerID domainPayment.ProviderID,
	) error
}

type handlerName string

func (h handlerName) String() string {
	return string(h)
}

const (
	menu                handlerName = "menu"
	startHandler        handlerName = "start"
	profileHandler      handlerName = "profile"
	helpHandler         handlerName = "help"
	productsHandler     handlerName = "products"
	productHandler      handlerName = "product"
	backHandler         handlerName = "back"
	createOrder         handlerName = "create_order"
	getOrderHandler     handlerName = "get_order"
	getOrderListHandler handlerName = "order_list"
	createInvoice       handlerName = "create_invoice"

	// administration
	adminHandler                handlerName = "admin"
	getAllProductsHandler       handlerName = "get_all_products"
	getProductForEditHandler    handlerName = "get_product_for_edit"
	createProductHandler        handlerName = "create_product"
	getCreateProductInfoHandler handlerName = "get_create_product_info"
	getCreateItemInfoHandler    handlerName = "get_create_item_info"
	deleteProductHandler        handlerName = "delete_product"
	createItemHandler           handlerName = "add_item"
	getItemHandler              handlerName = "get_item"
	deleteItemHandler           handlerName = "delete_item"
)

func (h handlerName) GetBackHandler() handlerName {
	switch h {
	case adminHandler, helpHandler, profileHandler, productsHandler:
		return menu
	case productHandler:
		return productsHandler
	case createOrder:
		return productHandler
	case createInvoice:
		return createOrder
	case createProductHandler, getAllProductsHandler:
		return adminHandler
	case getProductForEditHandler:
		return getAllProductsHandler
	case getCreateItemInfoHandler:
		return getProductForEditHandler
	case getItemHandler:
		return getProductForEditHandler
	case deleteItemHandler:
		return getItemHandler
	case deleteProductHandler:
		return getAllProductsHandler
	case getOrderListHandler:
		return profileHandler
	case getOrderHandler:
		return getOrderListHandler
	}

	return menu
}

type Implementation struct {
	productsUseCase productsUseCase
	userUseCase     userUseCase
	orderUseCase    orderUseCase
	paymentClient   paymentClient
	bot             *telegramBot.Bot
}

const (
	maxProductsLines   = 5
	maxProductsColumns = 2
)

func NewImplementation(
	productsUseCase productsUseCase,
	userUseCase userUseCase,
	orderUseCase orderUseCase,
	paymentClient paymentClient,
	bot *telegramBot.Bot,
) *Implementation {
	i := &Implementation{
		productsUseCase: productsUseCase,
		orderUseCase:    orderUseCase,
		userUseCase:     userUseCase,
		paymentClient:   paymentClient,
		bot:             bot,
	}

	bot.RegisterHandler(telegramBot.HandlerTypeMessageText,
		menu.String(), telegramBot.MatchTypeCommand, i.GetMenu)
	bot.RegisterHandler(telegramBot.HandlerTypeCallbackQueryData,
		menu.String(), telegramBot.MatchTypePrefix, i.GetMenu)
	bot.RegisterHandler(telegramBot.HandlerTypeMessageText,
		startHandler.String(), telegramBot.MatchTypeCommand, i.Start)
	bot.RegisterHandler(telegramBot.HandlerTypeCallbackQueryData,
		helpHandler.String(), telegramBot.MatchTypePrefix, i.GetHelp)
	bot.RegisterHandler(telegramBot.HandlerTypeCallbackQueryData,
		profileHandler.String(), telegramBot.MatchTypePrefix, i.GetProfile)

	bot.RegisterHandler(telegramBot.HandlerTypeCallbackQueryData,
		productsHandler.String(), telegramBot.MatchTypePrefix, i.GetProducts)
	bot.RegisterHandler(telegramBot.HandlerTypeCallbackQueryData,
		productHandler.String(), telegramBot.MatchTypePrefix, i.GetProduct)

	bot.RegisterHandler(telegramBot.HandlerTypeCallbackQueryData,
		getAllProductsHandler.String(), telegramBot.MatchTypePrefix, i.GetAllProducts)
	bot.RegisterHandler(telegramBot.HandlerTypeCallbackQueryData,
		getProductForEditHandler.String(), telegramBot.MatchTypePrefix, i.GetProductForEdit)
	bot.RegisterHandler(telegramBot.HandlerTypeCallbackQueryData,
		adminHandler.String(), telegramBot.MatchTypePrefix, i.GetAdminMenu)
	bot.RegisterHandler(telegramBot.HandlerTypeMessageText,
		createProductHandler.String(), telegramBot.MatchTypeCommand, i.CreateProduct)
	bot.RegisterHandler(telegramBot.HandlerTypeCallbackQueryData,
		getCreateProductInfoHandler.String(), telegramBot.MatchTypePrefix, i.GetCreateProductInfo)
	bot.RegisterHandler(telegramBot.HandlerTypeCallbackQueryData,
		deleteProductHandler.String(), telegramBot.MatchTypePrefix, i.DeleteProduct)
	bot.RegisterHandler(telegramBot.HandlerTypeCallbackQueryData,
		getCreateItemInfoHandler.String(), telegramBot.MatchTypePrefix, i.GetCreateItemInfo)
	bot.RegisterHandler(telegramBot.HandlerTypeCallbackQueryData,
		deleteItemHandler.String(), telegramBot.MatchTypePrefix, i.DeleteItem)
	bot.RegisterHandler(telegramBot.HandlerTypeMessageText,
		createItemHandler.String(), telegramBot.MatchTypeCommand, i.AddItem)
	bot.RegisterHandler(telegramBot.HandlerTypeCallbackQueryData,
		getItemHandler.String(), telegramBot.MatchTypePrefix, i.GetItem)

	bot.RegisterHandler(telegramBot.HandlerTypeCallbackQueryData,
		createOrder.String(), telegramBot.MatchTypePrefix, i.CreateOrder)
	bot.RegisterHandler(telegramBot.HandlerTypeCallbackQueryData,
		getOrderHandler.String(), telegramBot.MatchTypePrefix, i.GetOrder)
	bot.RegisterHandler(telegramBot.HandlerTypeCallbackQueryData,
		getOrderListHandler.String(), telegramBot.MatchTypePrefix, i.OrderList)

	bot.RegisterHandler(telegramBot.HandlerTypeCallbackQueryData,
		createInvoice.String(), telegramBot.MatchTypePrefix, i.CreateInvoice)

	return i
}

func (i *Implementation) Start(ctx context.Context, bot *telegramBot.Bot, update *models.Update) {
	id := update.Message.From.ID
	_, err := i.userUseCase.RegisterByTelegramID(ctx, domainUser.TelegramBotRegister{
		TelegramID: domainUser.NewTelegramID(id),
		ChatID:     update.Message.Chat.ID,
	})
	if err != nil {
		logger.Errorf(ctx, "err: %v", err)
		return
	}

	_, err = bot.SendMessage(ctx, &telegramBot.SendMessageParams{
		BusinessConnectionID: "",
		ChatID:               update.Message.Chat.ID,
		Text:                 "Привет! ты успешно зарегистрирован. Выполни команду /menu, чтобы начать ",
	})
	if err != nil {
		logger.Errorf(ctx, "err: %v", err)
		return
	}
}

func (i *Implementation) GetMenu(ctx context.Context, bot *telegramBot.Bot, update *models.Update) {
	results := [][]models.InlineKeyboardButton{
		{
			{
				Text:         "Продукты",
				CallbackData: productsHandler.String(),
			},
			{
				Text:         "Профиль",
				CallbackData: profileHandler.String(),
			},
		},
		{
			{
				Text:         "Помощь",
				CallbackData: helpHandler.String(),
			},
		},

		{
			{
				Text:         "Администрирование",
				CallbackData: adminHandler.String(),
			},
		},
	}

	var chatID int64
	if update.CallbackQuery != nil {
		chatID = update.CallbackQuery.Message.Message.Chat.ID
		_, err := bot.DeleteMessage(ctx, &telegramBot.DeleteMessageParams{
			ChatID:    update.CallbackQuery.Message.Message.Chat.ID,
			MessageID: update.CallbackQuery.Message.Message.ID,
		})
		if err != nil {
			return
		}
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
