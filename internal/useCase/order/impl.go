package order

import (
	"context"
	"errors"
	"time"

	"github.com/kdv2001/onlySubscription/internal/domain/communication"
	"github.com/kdv2001/onlySubscription/internal/domain/consts"
	"github.com/kdv2001/onlySubscription/internal/domain/order"
	domainProducts "github.com/kdv2001/onlySubscription/internal/domain/products"
	"github.com/kdv2001/onlySubscription/internal/domain/subscription"
	domainUser "github.com/kdv2001/onlySubscription/internal/domain/user"
	custom_errors "github.com/kdv2001/onlySubscription/pkg/errors"
)

type productUC interface {
	GetProduct(ctx context.Context, id domainProducts.ID) (domainProducts.Product, error)

	GetItem(ctx context.Context, id domainProducts.ItemID) (domainProducts.Item, error)
	PerformedItem(ctx context.Context, itemID domainProducts.ItemID) error
	PreReserveItem(ctx context.Context, productID domainProducts.ID) (domainProducts.ItemID, error)
	ReserveItem(ctx context.Context, itemID domainProducts.ItemID) error
	DereserveItem(ctx context.Context, itemID domainProducts.ItemID) error
}

type orderRepo interface {
	CreateOrder(ctx context.Context, o order.Order) (order.ID, error)
	GetOrder(ctx context.Context, oID order.ID) (order.Order, error)
	UpdateOrderStatus(ctx context.Context, oID order.ID, status order.ChangeOrderStatus) error
	GetOrders(ctx context.Context, r order.RequestList) ([]order.Order, error)
}

type userUC interface {
	GetUser(ctx context.Context, id domainUser.ID) (domainUser.User, error)
}

type communicationClient interface {
	SendMessage(ctx context.Context, message communication.Message) error
}

type subscriptionUC interface {
	CreateSubscription(ctx context.Context, s subscription.Subscription) error
}

type Implementation struct {
	productUC           productUC
	userUC              userUC
	subscriptionUC      subscriptionUC
	orderRepo           orderRepo
	communicationClient communicationClient
}

func NewImplementation(
	productUC productUC,
	userUC userUC,
	subscriptionUC subscriptionUC,
	orderRepo orderRepo,
	communicationClient communicationClient,
) *Implementation {
	return &Implementation{
		productUC:           productUC,
		orderRepo:           orderRepo,
		userUC:              userUC,
		subscriptionUC:      subscriptionUC,
		communicationClient: communicationClient,
	}
}

// CreateOrder создает заказ
func (i *Implementation) CreateOrder(ctx context.Context, o order.CreateOrder) (order.ID, error) {
	product, err := i.productUC.GetProduct(ctx, o.ProductID)
	if err != nil {
		return order.ID{}, err
	}

	itemID, err := i.productUC.PreReserveItem(ctx, o.ProductID)
	if err != nil {
		return order.ID{}, err
	}

	defaultStatus := order.Form
	orderID, err := i.orderRepo.CreateOrder(ctx, order.Order{
		TotalPrice: product.Price,
		Status:     defaultStatus,
		UserID:     o.UserID,
		Product: order.Product{
			ItemID:    itemID,
			ProductID: product.ID,
		},
		TTL: time.Now().UTC().Add(consts.DefaultOrderTimeLimit),
	})
	if err != nil {
		return order.ID{}, err
	}

	// TODO получить таймлимит от продукта
	err = i.productUC.ReserveItem(ctx, itemID)
	if err != nil {
		return order.ID{}, err
	}

	c, _ := order.NewChangeOrderStatus(defaultStatus, order.ExpectPayments)

	// TODO написать механизм перепроверки статуса заказа, двруг, что-то отвалитс
	err = i.orderRepo.UpdateOrderStatus(ctx, orderID, c)
	if err != nil {
		// TODO logger
		return orderID, nil
	}

	return orderID, nil
}

func (i *Implementation) GetOrder(ctx context.Context, oID order.ID, userID domainUser.ID) (order.Order, error) {
	o, err := i.orderRepo.GetOrder(ctx, oID)
	if err != nil {
		return order.Order{}, err
	}

	if o.UserID != userID {
		return order.Order{}, custom_errors.NewForbiddenError(errors.New("not user order"))
	}

	item, err := i.productUC.GetItem(ctx, o.Product.ItemID)
	if err != nil {
		return order.Order{}, err
	}

	product, err := i.productUC.GetProduct(ctx, item.ProductID)
	if err != nil {
		return order.Order{}, err
	}

	o.SetProduct(order.Product{
		ItemID:      o.Product.ItemID,
		ProductID:   item.ProductID,
		Title:       product.Name,
		Description: product.Description,
	})

	o.Product.ProductID = item.ProductID

	return o, nil
}

// GetOrderList список заказов
func (i *Implementation) GetOrderList(ctx context.Context, userID domainUser.ID, list order.RequestList) ([]order.Order, error) {
	if list.Filters == nil {
		list.Filters.UserID = userID
	} else {
		list.Filters = &order.Filters{
			UserID: userID,
		}
	}

	orders, errG := i.orderRepo.GetOrders(ctx, list)
	if errG != nil {
		return []order.Order{}, errG
	}

	result := make([]order.Order, 0, len(orders))
	for _, o := range orders {
		item, err := i.productUC.GetItem(ctx, o.Product.ItemID)
		if err != nil {
			return []order.Order{}, err
		}

		product, err := i.productUC.GetProduct(ctx, item.ProductID)
		if err != nil {
			return []order.Order{}, err
		}

		o.SetProduct(order.Product{
			ItemID:      o.Product.ItemID,
			ProductID:   item.ProductID,
			Title:       product.Name,
			Description: product.Description,
		})

		o.Product.ProductID = item.ProductID

		result = append(result, o)
	}

	return result, nil
}

// PaymentHandling перевод заказа в статус "обработка оплаты"
func (i *Implementation) PaymentHandling(ctx context.Context, oID order.ID) error {
	o, err := i.orderRepo.GetOrder(ctx, oID)
	if err != nil {
		return err
	}

	c, err := order.NewChangeOrderStatus(o.Status, order.Handling)
	if err != nil {
		if errors.Is(err, order.ErrStatusIsEqual) {
			return nil
		}

		return err
	}

	err = i.orderRepo.UpdateOrderStatus(ctx, oID, c)
	if err != nil {
		return err
	}

	return nil
}

// Processing перевод заказа в статус "обработка"
func (i *Implementation) Processing(ctx context.Context, oID order.ID) error {
	o, err := i.orderRepo.GetOrder(ctx, oID)
	if err != nil {
		return err
	}

	c, err := order.NewChangeOrderStatus(o.Status, order.Processing)
	if err != nil {
		if errors.Is(err, order.ErrStatusIsEqual) {
			return nil
		}

		return err
	}

	user, errU := i.userUC.GetUser(ctx, o.UserID)
	if errU != nil {
		return errU
	}

	err = i.communicationClient.SendMessage(ctx, communication.Message{
		ChatID:      user.Contact.TelegramBotChatID,
		Title:       "Заказ № " + o.ID.String(),
		Description: "Заказ оплачен, ожидайте товар придет отдельным сообщением",
	})
	if err != nil {
		return err
	}

	err = i.orderRepo.UpdateOrderStatus(ctx, oID, c)
	if err != nil {
		return err
	}

	return nil
}

// Canceled перевод заказа в статус "отменен"
func (i *Implementation) Canceled(ctx context.Context, oID order.ID) error {
	o, err := i.orderRepo.GetOrder(ctx, oID)
	if err != nil {
		return err
	}

	c, err := order.NewChangeOrderStatus(o.Status, order.Cancelled)
	if err != nil {
		if errors.Is(err, order.ErrStatusIsEqual) {
			return nil
		}

		return err
	}

	user, errU := i.userUC.GetUser(ctx, o.UserID)
	if errU != nil {
		return errU
	}

	err = i.communicationClient.SendMessage(ctx, communication.Message{
		ChatID:      user.Contact.TelegramBotChatID,
		Title:       "Заказ № " + o.ID.String(),
		Description: "Отменен по истечению времени жизни",
	})
	if err != nil {
		return err
	}

	err = i.orderRepo.UpdateOrderStatus(ctx, oID, c)
	if err != nil {
		return err
	}

	return nil
}
