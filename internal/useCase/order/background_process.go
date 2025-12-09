package order

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/kdv2001/onlySubscription/internal/domain/communication"
	"github.com/kdv2001/onlySubscription/internal/domain/order"
	"github.com/kdv2001/onlySubscription/internal/domain/primitives"
	"github.com/kdv2001/onlySubscription/internal/domain/subscription"
	"github.com/kdv2001/onlySubscription/pkg/parallel"
)

// RunBackgroundProcess запускает фоновый процесс
func (i *Implementation) RunBackgroundProcess(ctx context.Context, wg *sync.WaitGroup) error {
	go parallel.BackgroundPeriodProcess(ctx, wg, 5*time.Second, i.processingOrders)
	go parallel.BackgroundPeriodProcess(ctx, wg, 5*time.Second, i.cancelOrders)
	return nil

}

// processingOrders обработка оплаченных заказов
func (i *Implementation) processingOrders(ctx context.Context) error {
	orders, errG := i.orderRepo.GetOrders(ctx, order.RequestList{
		Pagination: &primitives.Pagination{
			Num: 15,
		},
		Filters: &order.Filters{
			Statuses: []order.Status{order.Processing},
		},
	})
	if errG != nil {
		return errG
	}

	for _, o := range orders {
		changeStatus, err := order.NewChangeOrderStatus(o.Status, order.Performed)
		if err != nil {
			if errors.Is(err, order.ErrStatusIsEqual) {
				return nil
			}

			return err
		}

		err = i.productUC.PerformedItem(ctx, o.Product.ItemID)
		if err != nil {
			continue
		}

		user, errU := i.userUC.GetUser(ctx, o.UserID)
		if errU != nil {
			return errU
		}

		item, errI := i.productUC.GetItem(ctx, o.Product.ItemID)
		if errI != nil {
			return errI
		}

		p, errI := i.productUC.GetProduct(ctx, item.ProductID)
		if errI != nil {
			return errI
		}

		err = i.subscriptionUC.CreateSubscription(ctx, subscription.Subscription{
			UserID:      o.UserID,
			OrderID:     o.ID,
			Deadline:    time.Now().UTC().Add(p.SubscriptionPeriod),
			Description: "Обновите продукт",
		})
		if err != nil {
			return err
		}

		err = i.communicationClient.SendMessage(ctx, communication.Message{
			ChatID:      user.Contact.TelegramBotChatID,
			Title:       "Заказ № " + o.ID.String(),
			Description: "Полезная нагрузка: " + item.Payload,
		})
		if err != nil {
			return err
		}

		err = i.orderRepo.UpdateOrderStatus(ctx, o.ID, changeStatus)
		if err != nil {
			return err
		}
	}

	return nil
}

// cancelOrders отмена заказов
func (i *Implementation) cancelOrders(ctx context.Context) error {
	orders, errG := i.orderRepo.GetOrders(ctx, order.RequestList{
		Pagination: &primitives.Pagination{
			Num: 15,
		},
		Filters: &order.Filters{
			Statuses: []order.Status{order.ExpectPayments, order.Form},
			TTL: &primitives.IntervalFilter[time.Time]{
				To: time.Now().UTC(),
			},
		},
	})
	if errG != nil {
		return errG
	}

	for _, o := range orders {
		changeStatus, err := order.NewChangeOrderStatus(o.Status, order.Cancelled)
		if err != nil {
			if errors.Is(err, order.ErrStatusIsEqual) {
				return nil
			}

			return err
		}

		err = i.productUC.DereserveItem(ctx, o.Product.ItemID)
		if err != nil {
			continue
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

		err = i.orderRepo.UpdateOrderStatus(ctx, o.ID, changeStatus)
		if err != nil {
			return err
		}
	}

	return nil
}

// TODO спасение заказов
//func (i *Implementation) recoveryOrder() {
//	i.productUC.GetItem()
//}
