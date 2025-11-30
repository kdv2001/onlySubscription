package payment

import (
	"context"
	"errors"

	"github.com/kdv2001/onlySubscription/internal/domain/order"
	domainPayment "github.com/kdv2001/onlySubscription/internal/domain/payment"
	"github.com/kdv2001/onlySubscription/internal/domain/primitives"
	"github.com/kdv2001/onlySubscription/internal/domain/user"
	custom_errors "github.com/kdv2001/onlySubscription/pkg/errors"
)

type paymentRepo interface {
	GetInvoice(
		ctx context.Context,
		iID domainPayment.ID,
	) (domainPayment.Invoice, error)
	UpdateInvoice(
		ctx context.Context,
		id domainPayment.ID,
		changeState domainPayment.ChangeInvoice,
	) error
	CreateInvoice(
		ctx context.Context,
		invoice domainPayment.Invoice,
	) (domainPayment.ID, error)
	GetProcessingInvoices(
		ctx context.Context,
		rrq domainPayment.RequestList,
	) ([]domainPayment.Invoice, error)
}

type orderUC interface {
	GetOrder(ctx context.Context, oID order.ID, userID user.ID) (order.Order, error)
	PaymentHandling(ctx context.Context, oID order.ID) error
	Processing(ctx context.Context, oID order.ID) error
	Canceled(ctx context.Context, oID order.ID) error
}

type provider interface {
	GetTransactions(ctx context.Context,
		pagination primitives.Pagination,
	) ([]domainPayment.ProviderTransaction, error)
}

type Implementation struct {
	paymentRepo paymentRepo
	orderUC     orderUC
	provider    provider
}

func NewImplementation(
	paymentRepo paymentRepo,
	orderUC orderUC,
	provider provider) *Implementation {
	return &Implementation{
		paymentRepo: paymentRepo,
		orderUC:     orderUC,
		provider:    provider,
	}
}

func (i *Implementation) CreateInvoice(ctx context.Context,
	invoice domainPayment.CreateInvoice) (domainPayment.ReleaseInvoice, error) {
	o, err := i.orderUC.GetOrder(ctx, invoice.OrderID, invoice.UserID)
	if err != nil {
		return domainPayment.ReleaseInvoice{}, err
	}

	invoiceID, err := i.paymentRepo.CreateInvoice(ctx, domainPayment.Invoice{
		OrderID:       invoice.OrderID,
		State:         domainPayment.ExpectPaymentState,
		Price:         o.TotalPrice,
		PaymentMethod: invoice.PaymentMethod,
	})
	if err != nil {
		return domainPayment.ReleaseInvoice{}, err
	}

	return domainPayment.ReleaseInvoice{
		ID:            invoiceID,
		Price:         o.TotalPrice,
		PaymentMethod: invoice.PaymentMethod,
		TelegramData:  invoice.TelegramData,
		Position: domainPayment.Position{
			Title:       o.Product.Title,
			Description: o.Product.Description,
			Price:       o.TotalPrice,
		},
	}, nil
}

func (i *Implementation) GetInvoice(ctx context.Context, id domainPayment.ID) (domainPayment.Invoice, error) {
	return i.paymentRepo.GetInvoice(ctx, id)
}

func (i *Implementation) Handling(ctx context.Context,
	id domainPayment.ID,
	providerID domainPayment.ProviderID,
) error {
	invoice, err := i.paymentRepo.GetInvoice(ctx, id)
	if err != nil {
		return err
	}

	c, err := domainPayment.NewChangeState(invoice.State, domainPayment.HandlingState)
	if err != nil && !errors.Is(err, order.ErrStatusIsEqual) {
		return err
	}

	err = i.orderUC.PaymentHandling(ctx, invoice.OrderID)
	if err != nil {
		// если не удалось подтвердить заказ, значит он уже протух, отменяем платеж
		if errors.Is(err, custom_errors.ErrorBadRequest) {
			c, err = domainPayment.NewChangeState(invoice.State, domainPayment.CanceledState)
			if err != nil && !errors.Is(err, order.ErrStatusIsEqual) {
				return err
			}
		}
		return err
	}

	err = i.paymentRepo.UpdateInvoice(ctx, id, domainPayment.ChangeInvoice{
		ProviderID:  providerID,
		ChangeState: c,
	})
	if err != nil {
		return err
	}

	return nil
}

func (i *Implementation) Processing(ctx context.Context,
	id domainPayment.ID,
	transactionalProviderID domainPayment.ProviderID,
) error {
	invoice, err := i.paymentRepo.GetInvoice(ctx, id)
	if err != nil {
		return err
	}

	c, err := domainPayment.NewChangeState(invoice.State, domainPayment.ProcessingState)
	if err != nil && !errors.Is(err, domainPayment.ErrStatusIsEqual) {
		return err
	}

	err = i.orderUC.PaymentHandling(ctx, invoice.OrderID)
	if err != nil {
		return err
	}

	err = i.paymentRepo.UpdateInvoice(ctx, id, domainPayment.ChangeInvoice{
		ProviderID:  transactionalProviderID,
		ChangeState: c,
	})
	if err != nil {
		return err
	}

	return nil
}
