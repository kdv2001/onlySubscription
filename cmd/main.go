package main

import (
	"context"
	"fmt"
	"log"
	"os/signal"
	"syscall"
	"time"

	telegramBot "github.com/go-telegram/bot"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"

	"github.com/kdv2001/onlySubscription/internal/domain/order"
	"github.com/kdv2001/onlySubscription/internal/domain/price"
	domainProducts "github.com/kdv2001/onlySubscription/internal/domain/products"
	domainUser "github.com/kdv2001/onlySubscription/internal/domain/user"
	telegramHandlers "github.com/kdv2001/onlySubscription/internal/handlers/telegram_bot"
	orderPostgres "github.com/kdv2001/onlySubscription/internal/repositories/order/postgres"
	productspostgres "github.com/kdv2001/onlySubscription/internal/repositories/products/postgres"
	"github.com/kdv2001/onlySubscription/internal/repositories/user/psotgress"
	orderusecase "github.com/kdv2001/onlySubscription/internal/useCase/order"
	productsusecase "github.com/kdv2001/onlySubscription/internal/useCase/products"
	userusecase "github.com/kdv2001/onlySubscription/internal/useCase/users"
)

func main() {
	ctx := context.Background()
	ctx, cancel := signal.NotifyContext(ctx,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	defer cancel()

	values, err := initFlags()
	if err != nil {
		log.Fatal(err)
	}

	tgBot, err := telegramBot.New(values.TelegramToken,
		telegramBot.WithCheckInitTimeout(20*time.Second))
	if err != nil {
		log.Fatal(err)
	}

	postgresConn, err := pgxpool.New(ctx, values.PostgresDSN)
	if err != nil {
		log.Fatal(err)
	}

	productsPostgresConn, err := productspostgres.NewImplementation(ctx, postgresConn)
	if err != nil {
		log.Fatal(err)
	}

	userPostgresConn, err := user.NewImplementation(ctx, postgresConn)
	if err != nil {
		log.Fatal(err)
	}

	orderPostgresConn, err := orderPostgres.NewImplementation(ctx, postgresConn)
	if err != nil {
		log.Fatal(err)
	}

	productsUseCases := productsusecase.NewImplementation(productsPostgresConn)

	orderUC := orderusecase.NewImplementation(productsUseCases, orderPostgresConn)

	userUC := userusecase.NewImplementation(userPostgresConn)

	_, err = userUC.RegisterByTelegramID(ctx, domainUser.TelegramBotRegister{
		TelegramID: domainUser.NewTelegramID("123434"),
	})

	user, err := userUC.GetUserByTelegramID(ctx, domainUser.NewTelegramID("123434"))

	pID, err := productsPostgresConn.CreateProduct(ctx, domainProducts.Product{
		Type:        domainProducts.SubscriptionType,
		Name:        "SUBB",
		Description: "SUBBBBB",
		Image: domainProducts.Image{
			URL: "image-link",
		},
		Price: price.Price{
			Currency: price.XTR,
			Value:    decimal.NewFromInt(100),
		},
		SubscriptionPeriod: 30 * time.Second,
	})
	if err != nil {
		log.Fatal(err)
	}

	_, err = productsPostgresConn.CreateInventoryItem(ctx, domainProducts.Item{
		ProductID:   pID,
		Status:      domainProducts.SaleStatus,
		Description: "1234",
	})
	if err != nil {
		log.Fatal(err)
	}

	oID, err := orderUC.CreateOrder(ctx, order.CreateOrder{
		UserID:    user.ID,
		ProductID: pID,
	})

	fmt.Println(orderUC.GetOrder(ctx, oID, user.ID))

	go func() {
		errU := productsUseCases.RunUpdateExpiredItems(ctx)
		if errU != nil {
			log.Fatal(err)
		}
	}()

	fmt.Println(productsPostgresConn.GetExpiredPreReservedItems(ctx, 2))

	// // TODO добавить Логгер мидлвар + еррор мидлвар
	_ = telegramHandlers.NewImplementation(productsUseCases, tgBot)

	tgBot.Start(ctx)
}

//
//func callbackHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
//	// answering callback query first to let Telegram know that we received the callback query,
//	// and we're handling it. Otherwise, Telegram might retry sending the update repetitively
//	// as it thinks the callback query doesn't reach to our application. learn more by
//	// reading the footnote of the https://core.telegram.org/bots/api#callbackquery type.
//	if update.CallbackQuery != nil {
//		b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
//			CallbackQueryID: update.CallbackQuery.ID,
//			ShowAlert:       false,
//		})
//		b.SendMessage(ctx, &bot.SendMessageParams{
//			ChatID: update.CallbackQuery.Message.Message.Chat.ID,
//			Text:   "You selected the button: " + update.CallbackQuery.Data,
//		})
//	}
//}
//
//func defaultHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
//	kb := &models.InlineKeyboardMarkup{
//		InlineKeyboard: [][]models.InlineKeyboardButton{
//			{
//				{Text: "Button 1", CallbackData: "button_1"},
//				{Text: "Button 2", CallbackData: "button_2"},
//			}, {
//				{Text: "Button 3", CallbackData: "button_3"},
//			},
//		},
//	}
//
//	b.SendMessage(ctx, &bot.SendMessageParams{
//		ChatID:      update.Message.Chat.ID,
//		Text:        "Click by button",
//		ReplyMarkup: kb,
//	})
//}
