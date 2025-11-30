package products

import (
	"context"
	"errors"
	"time"

	"github.com/kdv2001/onlySubscription/internal/domain/consts"
	"github.com/kdv2001/onlySubscription/internal/domain/primitives"
	domainProducts "github.com/kdv2001/onlySubscription/internal/domain/products"
	"github.com/kdv2001/onlySubscription/pkg/parallel"
)

// RunBackgroundProcess запускает фоновые процессы
func (i *Implementation) RunBackgroundProcess(ctx context.Context) error {
	go parallel.BackgroundPeriodProcess(ctx, 20*time.Second, i.updateExpiredItems)
	return nil
}

// updateExpiredItems снимает пререзервацию с продуктов
func (i *Implementation) updateExpiredItems(ctx context.Context) error {
	items, errG := i.productsRepo.GetItems(ctx, domainProducts.RequestList{
		Pagination: &primitives.Pagination{
			Num: maxProcessingItems,
		},
		Filters: &domainProducts.Filters{
			UpdatedAt: &primitives.IntervalFilter[time.Time]{
				To: time.Now().UTC().Add(-consts.DefaultPrereservedTTL),
			},
			Statuses: []domainProducts.ItemStatus{domainProducts.PreReservedStatus},
		},
	})
	if errG != nil {
		return errG
	}

	for _, item := range items {
		c, err := domainProducts.NewChangeItemStatus(item.Status, domainProducts.SaleStatus)
		if err != nil {
			if errors.Is(err, domainProducts.ErrStatusIsEqual) {
				return nil
			}

			return err
		}

		err = i.productsRepo.ChangeItemStatus(ctx, item.ID, c)
		if err != nil {

			return err
		}
	}

	return nil
}

// realizedItems переводит item в статус реализован
func (i *Implementation) realizedItems(ctx context.Context) error {
	items, errG := i.productsRepo.GetItems(ctx, domainProducts.RequestList{
		Pagination: &primitives.Pagination{
			Num: maxProcessingItems,
		},
		Filters: &domainProducts.Filters{
			Statuses: []domainProducts.ItemStatus{domainProducts.PerformedStatus},
		},
	})
	if errG != nil {
		return errG
	}

	for _, item := range items {
		c, err := domainProducts.NewChangeItemStatus(item.Status, domainProducts.RealizedStatus)
		if err != nil {
			if errors.Is(err, domainProducts.ErrStatusIsEqual) {
				return nil
			}

			return err
		}
		// TODO какая-то постобработка элемента
		err = i.productsRepo.ChangeItemStatus(ctx, item.ID, c)
		if err != nil {

			return err
		}
	}

	return nil
}
