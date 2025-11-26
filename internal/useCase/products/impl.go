package products

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/kdv2001/onlySubscription/internal/domain/primitives"
	domainProducts "github.com/kdv2001/onlySubscription/internal/domain/products"
	custom_errors "github.com/kdv2001/onlySubscription/pkg/errors"
)

// maxProductsSize максимальное кол-во продуктов в ответе
const maxProductsSize = 10

type productsRepo interface {
	GetProducts(ctx context.Context, req primitives.RequestList) (domainProducts.Products, error)
	CreateProduct(ctx context.Context, req domainProducts.Product) (domainProducts.ID, error)
	DeleteProduct(ctx context.Context, id domainProducts.ID) error
	GetProduct(ctx context.Context, id domainProducts.ID) (domainProducts.Product, error)

	CreateInventoryItem(ctx context.Context, req domainProducts.Item) (domainProducts.ItemID, error)
	DeleteInventoryItem(ctx context.Context, id domainProducts.ItemID) error

	PreReservedSProduct(ctx context.Context, itemID domainProducts.ID) (domainProducts.ItemID, error)
	ChangeItemStatus(ctx context.Context,
		itemID domainProducts.ItemID,
		changeItemStatus domainProducts.ChangeItemStatus,
	) error
	GetItem(ctx context.Context, id domainProducts.ItemID) (domainProducts.Item, error)
	GetExpiredPreReservedItems(ctx context.Context, num int) ([]domainProducts.Item, error)
}

type Implementation struct {
	productsRepo   productsRepo
	preselectedTTL time.Duration
}

func NewImplementation(productsRepo productsRepo) *Implementation {
	return &Implementation{
		productsRepo: productsRepo,
	}
}

// GetProducts возвращает набор продуктов
func (i *Implementation) GetProducts(ctx context.Context, req primitives.RequestList) (domainProducts.Products, error) {
	if req.Pagination.Num > maxProductsSize {
		return nil, custom_errors.NewBadRequestError(errors.New("bad products num")).
			SetDescription(fmt.Sprintf("request cards num grater than %d", maxProductsSize))
	}

	res, err := i.productsRepo.GetProducts(ctx, req)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (i *Implementation) GetProduct(ctx context.Context, id domainProducts.ID) (domainProducts.Product, error) {
	product, err := i.productsRepo.GetProduct(ctx, id)
	if err != nil {
		return product, err
	}

	return product, nil
}

// CreateProduct возвращает набор продуктов
func (i *Implementation) CreateProduct(ctx context.Context, product domainProducts.Product) error {
	_, err := i.productsRepo.CreateProduct(ctx, product)
	if err != nil {
		return err
	}

	return nil
}

func (i *Implementation) AddInventoryItem(ctx context.Context, item domainProducts.Item) error {
	_, err := i.productsRepo.GetProduct(ctx, item.ProductID)
	if err != nil {
		return err
	}

	// Реализация добавления товара в инвентарь
	_, err = i.productsRepo.CreateInventoryItem(ctx, item)
	if err != nil {
		return err
	}

	return nil
}

func (i *Implementation) RemoveInventoryItem(ctx context.Context, id domainProducts.ItemID) error {
	// Реализация удаления товара из инвентаря
	err := i.productsRepo.DeleteInventoryItem(ctx, id)
	if err != nil {
		return err
	}

	return nil
}

// PreReserveItem резервирует товар для заказа. Возвращает ID забронированного товара и ошибку, если произошла ошибка.
func (i *Implementation) PreReserveItem(ctx context.Context, productID domainProducts.ID) (domainProducts.ItemID, error) {
	reservedItemID, err := i.productsRepo.PreReservedSProduct(ctx, productID)
	if err != nil {
		return domainProducts.ItemID{}, err
	}

	return reservedItemID, nil
}

func (i *Implementation) ReserveProduct(ctx context.Context, itemID domainProducts.ItemID) error {
	err := i.productsRepo.ChangeItemStatus(ctx, itemID, domainProducts.ChangeItemStatus{
		From: domainProducts.PreReservedStatus,
		To:   domainProducts.ReservedStatus,
	})
	if err != nil {

		return err
	}

	return nil
}

func (i *Implementation) DereserveProduct(ctx context.Context, itemID domainProducts.ItemID) error {
	err := i.productsRepo.ChangeItemStatus(ctx, itemID, domainProducts.ChangeItemStatus{
		From: domainProducts.ReservedStatus,
		To:   domainProducts.SaleStatus,
	})
	if err != nil {

		return err
	}

	return nil
}

func (i *Implementation) SaleProduct(ctx context.Context, itemID domainProducts.ItemID) error {
	err := i.productsRepo.ChangeItemStatus(ctx, itemID, domainProducts.ChangeItemStatus{
		From: domainProducts.ReservedStatus,
		To:   domainProducts.PerformedStatus,
	})
	if err != nil {
		return err
	}

	return nil
}

func (i *Implementation) RunUpdateExpiredItems(ctx context.Context) error {
	ticker := time.NewTicker(60 * time.Second)
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			err := i.updateExpiredItems(ctx)
			if err != nil {
				return err
			}
		}
	}
}

func (i *Implementation) updateExpiredItems(ctx context.Context) error {
	// TODO добавить воркер пулл
	items, err := i.productsRepo.GetExpiredPreReservedItems(ctx, 30)
	if err != nil {
		return err
	}

	for _, item := range items {
		err = i.productsRepo.ChangeItemStatus(ctx, item.ID, domainProducts.ChangeItemStatus{
			From: domainProducts.PreReservedStatus,
			To:   domainProducts.SaleStatus,
		})
		if err != nil {

			return err
		}
	}

	return nil
}

func (i *Implementation) CreateSubscription() error {
	return nil
}

func (i *Implementation) GetItem(ctx context.Context, id domainProducts.ItemID) (domainProducts.Item, error) {
	return i.productsRepo.GetItem(ctx, id)
}
