package payment

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/kdv2001/onlySubscription/internal/domain/consts"
	domainPayment "github.com/kdv2001/onlySubscription/internal/domain/payment"
	"github.com/kdv2001/onlySubscription/internal/domain/primitives"
	"github.com/kdv2001/onlySubscription/pkg/parallel"
)

// RunBackgroundProcess запускает фоновые процессы
func (i *Implementation) RunBackgroundProcess(ctx context.Context, wg *sync.WaitGroup) error {
	go parallel.BackgroundPeriodProcess(ctx, wg, 5*time.Second, i.processingInvoices)
	go parallel.BackgroundPeriodProcess(ctx, wg, 30*time.Second, i.handlingInvoice)

	return nil
}

// processingInvoices фоново обрабатывает счета
func (i *Implementation) processingInvoices(ctx context.Context) error {
	invoices, err := i.paymentRepo.GetProcessingInvoices(ctx, domainPayment.RequestList{
		Pagination: &primitives.Pagination{
			Num: 15,
		},
		Filters: &domainPayment.Filters{
			Statuses: []domainPayment.State{
				domainPayment.ProcessingState,
			},
		},
	})
	if err != nil {
		return err
	}

	for _, invoice := range invoices {
		err = i.performedInvoice(ctx, invoice)
		if err != nil {
			return err
		}
	}

	return nil
}

// performedInvoice переводит счета в состояние оплачено
func (i *Implementation) performedInvoice(ctx context.Context, invoice domainPayment.Invoice) error {
	c, err := domainPayment.NewChangeState(invoice.State, domainPayment.PerformedState)
	if err != nil {
		if errors.Is(err, domainPayment.ErrStatusIsEqual) {
			return nil
		}
		return err
	}

	err = i.orderUC.Processing(ctx, invoice.OrderID)
	if err != nil {
		return err
	}

	err = i.paymentRepo.UpdateInvoice(ctx, invoice.ID, domainPayment.ChangeInvoice{
		ProviderID:  invoice.ProviderID,
		ChangeState: c,
	})
	if err != nil {
		return err
	}

	return nil
}

// canceledInvoices отменяет счета ожидающие оплаты
func (i *Implementation) handlingInvoice(ctx context.Context) error {
	invoices, errG := i.paymentRepo.GetProcessingInvoices(ctx, domainPayment.RequestList{
		Pagination: &primitives.Pagination{
			Num: 15,
		},
		Filters: &domainPayment.Filters{
			Statuses: []domainPayment.State{
				domainPayment.HandlingState,
			},
			UpdatedAt: &primitives.IntervalFilter[time.Time]{
				To: time.Now().UTC().Add(-consts.DefaultHandlingDur),
			},
		},
		Sort: &domainPayment.Sort{
			UpdateAt: primitives.Descending,
		},
	})
	if errG != nil {
		return errG
	}

	if len(invoices) == 0 {
		return nil
	}

	transactions, errG := i.getTransactions(ctx, invoices[len(invoices)-1])
	if errG != nil {
		return errG
	}

	// [NOTE] здесь можно сделать более качественный алгоритм
	// мы знаем, ориентируясь на кол-во счетов у нас
	// можно сразу определить нужный оффсет,
	// но пока сделаем на быструю руку
	// TODO прихранивать счетчик транзакций
	for _, invoice := range invoices {
		state := domainPayment.CanceledState
		for _, transaction := range transactions {
			if transaction.InternalID == invoice.ID && !transaction.ProviderID.IsEmpty() {
				state = domainPayment.ProcessingState
			}
		}

		err := i.orderUC.Canceled(ctx, invoice.OrderID)
		if err != nil {
			return err
		}

		err = i.updateInvoiceState(ctx, invoice, state)
		if err != nil {
			return err
		}
	}

	return nil
}

// updateInvoiceState обновляет состояние счета
func (i *Implementation) updateInvoiceState(ctx context.Context,
	invoice domainPayment.Invoice, state domainPayment.State) error {
	c, err := domainPayment.NewChangeState(invoice.State, state)
	if err != nil {
		if errors.Is(err, domainPayment.ErrStatusIsEqual) {
			return nil
		}
		return err
	}

	err = i.paymentRepo.UpdateInvoice(ctx, invoice.ID, domainPayment.ChangeInvoice{
		ProviderID:  invoice.ProviderID,
		ChangeState: c,
	})
	if err != nil {
		return err
	}

	return nil
}

// getTransactions получает транзакции от поставщика
func (i *Implementation) getTransactions(ctx context.Context, lastInvoice domainPayment.Invoice) ([]domainPayment.ProviderTransaction, error) {
	offset := 0
	num := 18
	transactions := make([]domainPayment.ProviderTransaction, 0, num)
	for {
		trans, err := i.provider.GetTransactions(ctx, primitives.Pagination{
			Num:    uint64(num),
			Offset: uint64(offset),
		})
		if err != nil {
			return nil, err
		}

		if len(trans) == 0 {
			return transactions, nil
		}

		transactions = append(transactions, trans...)

		lastTrans := trans[len(trans)-1]
		if lastTrans.Date.After(lastInvoice.CreatedAt) ||
			lastTrans.Date.Equal(lastInvoice.CreatedAt) {
			break
		}

		offset += num
	}

	return transactions, nil
}
