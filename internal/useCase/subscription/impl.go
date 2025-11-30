package subscription

import (
	"context"

	"github.com/kdv2001/onlySubscription/internal/domain/communication"
	domainProducts "github.com/kdv2001/onlySubscription/internal/domain/products"
	"github.com/kdv2001/onlySubscription/internal/domain/subscription"
	domainUser "github.com/kdv2001/onlySubscription/internal/domain/user"
)

type subscriptionRepo interface {
	GetSubscriptions(ctx context.Context, r subscription.RequestList) ([]subscription.Subscription, error)
	CreateSubscription(ctx context.Context, req subscription.Subscription) (domainProducts.ID, error)
	ChangeStatus(ctx context.Context,
		subID subscription.ID,
		changeState subscription.ChangeState,
	) error
}

type communicationClient interface {
	SendMessage(ctx context.Context, message communication.Message) error
}

type userUC interface {
	GetUser(ctx context.Context, id domainUser.ID) (domainUser.User, error)
}

type Implementation struct {
	communicationClient communicationClient
	subscriptionRepo    subscriptionRepo
	userUC              userUC
}

func NewImplementation(communicationClient communicationClient,
	subscriptionRepo subscriptionRepo,
	userUC userUC,
) *Implementation {
	return &Implementation{
		communicationClient: communicationClient,
		subscriptionRepo:    subscriptionRepo,
		userUC:              userUC,
	}
}

// CreateSubscription создает подписку
func (i *Implementation) CreateSubscription(ctx context.Context, s subscription.Subscription) error {
	s.State = subscription.ActiveState
	_, err := i.subscriptionRepo.CreateSubscription(ctx, s)
	return err
}
