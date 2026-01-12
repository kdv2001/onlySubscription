package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"
	"time"

	telegramBot "github.com/go-telegram/bot"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/kdv2001/onlySubscription/internal/clients/message/telegram_bot"
	telegramHandlers "github.com/kdv2001/onlySubscription/internal/handlers/telegram_bot"
	orderPostgres "github.com/kdv2001/onlySubscription/internal/repositories/order/postgres"
	paymentpostgres "github.com/kdv2001/onlySubscription/internal/repositories/payment/postgress"
	productspostgres "github.com/kdv2001/onlySubscription/internal/repositories/products/postgres"
	subscriptionpostgres "github.com/kdv2001/onlySubscription/internal/repositories/subscription/postgres"
	"github.com/kdv2001/onlySubscription/internal/repositories/user/psotgress"
	orderusecase "github.com/kdv2001/onlySubscription/internal/useCase/order"
	paymentusecase "github.com/kdv2001/onlySubscription/internal/useCase/payment"
	productsusecase "github.com/kdv2001/onlySubscription/internal/useCase/products"
	subscriptionusecase "github.com/kdv2001/onlySubscription/internal/useCase/subscription"
	userusecase "github.com/kdv2001/onlySubscription/internal/useCase/users"
	"github.com/kdv2001/onlySubscription/pkg/logger"
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

	zapLog, err := zap.NewDevelopment()
	if err != nil {
		log.Fatal(err)
	}
	sugarLogger := zapLog.Sugar()
	ctx = logger.ToContext(ctx, sugarLogger)

	// repositories
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

	subscriptionPostgresConn, err := subscriptionpostgres.NewImplementation(ctx, postgresConn)
	if err != nil {
		log.Fatal(err)
	}

	paymentPostgresConn, err := paymentpostgres.NewImplementation(ctx, postgresConn)
	if err != nil {
		log.Fatal(err)
	}

	// костыль, чтобы держать только один экземпляр ТГ клиента
	g := &getBot{}

	// clients
	messageClient := telegram_bot.NewImplementation(g)

	// usecases
	userUC := userusecase.NewImplementation(userPostgresConn)
	productsUseCases := productsusecase.NewImplementation(productsPostgresConn)
	subscriptionUC := subscriptionusecase.NewImplementation(messageClient, subscriptionPostgresConn, userUC)
	orderUC := orderusecase.NewImplementation(productsUseCases,
		userUC,
		subscriptionUC,
		orderPostgresConn,
		messageClient)
	paymentUC := paymentusecase.NewImplementation(paymentPostgresConn, orderUC, messageClient)

	authMW := telegramHandlers.NewAuthMiddleware(userUC)
	paymentMW := telegramHandlers.NewPaymentMiddleware(paymentUC)
	tgBot, err := telegramBot.New(values.TelegramToken,
		telegramBot.WithCheckInitTimeout(20*time.Second),
		telegramBot.WithMiddlewares(authMW.Middleware),
		telegramBot.WithMiddlewares(telegramHandlers.AddLoggerToContextMiddleware(sugarLogger)),
		telegramBot.WithMiddlewares(paymentMW.Middleware),
	)
	if err != nil {
		log.Fatal(err)
	}

	g.Bot = tgBot

	// запускаем фоновые процессы
	if err = productsUseCases.RunUpdateExpiredItems(ctx); err != nil {
		logger.Errorf(ctx, "panic: %v", err)
	}

	if err = orderUC.StartOrderBackgroundOrders(ctx); err != nil {
		log.Fatal(err)
	}

	if err = subscriptionUC.RunUpdateExpiredItems(ctx); err != nil {
		log.Fatal(err)
	}
	if err = paymentUC.RunBackgroundProcess(ctx); err != nil {
		log.Fatal(err)
	}

	_ = telegramHandlers.NewImplementation(productsUseCases, userUC, orderUC, paymentUC, tgBot)

	tgBot.Start(ctx)
}

type getBot struct {
	*telegramBot.Bot
}
