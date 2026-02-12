package subscription

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/kdv2001/onlySubscription/internal/domain/communication"
	"github.com/kdv2001/onlySubscription/internal/domain/primitives"
	"github.com/kdv2001/onlySubscription/internal/domain/subscription"
	"github.com/kdv2001/onlySubscription/pkg/logger"
	"github.com/kdv2001/onlySubscription/pkg/parallel"
)

// RunBackgroundProcess запускает фоновые процессы
func (i *Implementation) RunBackgroundProcess(ctx context.Context, wg *sync.WaitGroup) error {
	go parallel.BackgroundPeriodProcess(ctx, wg, 20*time.Second, i.deactivateExpiredSubscription)
	return nil
}

// maxProcessingItems кол-во элементов в выборке для обработки
const maxProcessingItems = 30

// deactivateExpiredSubscription деактивирует просроченные подписки
func (i *Implementation) deactivateExpiredSubscription(ctx context.Context) error {
	subscriptions, errG := i.subscriptionRepo.GetSubscriptions(ctx, subscription.RequestList{
		Pagination: &primitives.Pagination{
			Num: maxProcessingItems,
		},
		Filters: &subscription.Filters{
			Deadline: &primitives.IntervalFilter[time.Time]{
				To: time.Now().UTC(),
			},
			Statuses: []subscription.State{subscription.ActiveState},
		},
	})
	if errG != nil {
		return errG
	}

	for _, s := range subscriptions {
		c, err := subscription.NewChangeItemStatus(s.State, subscription.InactiveState)
		if err != nil {
			if errors.Is(err, subscription.ErrStatusIsEqual) {
				return nil
			}

			return err
		}

		user, err := i.userUC.GetUser(ctx, s.UserID)
		if err != nil {
			logger.Errorf(ctx, "error send subscription msg: %v", err)
			continue
		}

		msg := communication.Message{
			ChatID:      user.Contact.TelegramBotChatID,
			Title:       "Ваша подписка истекла",
			Description: s.Description,
		}

		if err = i.communicationClient.SendMessage(ctx, msg); err != nil {
			logger.Errorf(ctx, "error send subscription msg: %v", err)
			continue
		}

		if err = i.subscriptionRepo.ChangeStatus(ctx, s.ID, c); err != nil {
			return err
		}
	}

	return nil
}
