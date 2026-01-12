package products

import (
	"context"
	"errors"
	"fmt"
	"time"

	domainProducts "github.com/kdv2001/onlySubscription/internal/domain/products"
	custom_errors "github.com/kdv2001/onlySubscription/pkg/errors"
)

// maxProductsSize максимальное кол-во продуктов в ответе
const maxProductsSize = 15
const maxProcessingItems = 15

type productsRepo interface {
	GetProducts(ctx context.Context, req domainProducts.RequestList) (domainProducts.Products, error)
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
	GetItems(
		ctx context.Context,
		req domainProducts.RequestList,
	) ([]domainProducts.Item, error)
	CountItemsForProduct(ctx context.Context, productID domainProducts.ID) (int64, error)
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
func (i *Implementation) GetProducts(ctx context.Context, req domainProducts.RequestList) (domainProducts.Products, error) {
	if req.Pagination != nil && req.Pagination.Num > maxProductsSize {
		return nil, custom_errors.NewBadRequestError(errors.New("bad products num")).
			SetDescription(fmt.Sprintf("request cards num grater than %d", maxProductsSize))
	}

	res, err := i.productsRepo.GetProducts(ctx, req)
	if err != nil {
		return nil, err
	}

	if req.Filters != nil && req.Filters.ItemsExist {
		result := make(domainProducts.Products, 0, len(res))
		for _, r := range res {
			num, err := i.productsRepo.CountItemsForProduct(ctx, r.ID)
			if err != nil {
				return nil, err
			}
			if num == 0 {
				continue
			}

			result = append(result, r)
		}
		res = result
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

func (i *Implementation) DeactivateProduct(ctx context.Context, id domainProducts.ID) error {
	err := i.productsRepo.DeleteProduct(ctx, id)
	if err != nil {
		return err
	}

	return nil
}

// CreateProduct возвращает набор продуктов
func (i *Implementation) CreateProduct(ctx context.Context, product domainProducts.Product) error {
	_, err := i.productsRepo.CreateProduct(ctx, product)
	if err != nil {
		return err
	}

	return nil
}

func (i *Implementation) AddItem(ctx context.Context, item domainProducts.Item) error {
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

func (i *Implementation) DeleteItem(ctx context.Context, id domainProducts.ItemID) error {
	// Реализация удаления товара из инвентаря
	err := i.productsRepo.DeleteInventoryItem(ctx, id)
	if err != nil {
		return err
	}

	return nil
}

func (i *Implementation) GetItem(ctx context.Context, id domainProducts.ItemID) (domainProducts.Item, error) {
	return i.productsRepo.GetItem(ctx, id)
}

func (i *Implementation) GetItems(ctx context.Context, req domainProducts.RequestList) ([]domainProducts.Item, error) {
	if req.Pagination != nil && req.Pagination.Num > maxProductsSize {
		return nil, custom_errors.NewBadRequestError(errors.New("bad products num")).
			SetDescription(fmt.Sprintf("request cards num grater than %d", maxProductsSize))
	}

	availableStatuses := []domainProducts.ItemStatus{domainProducts.SaleStatus}
	if req.Filters == nil {
		req.Filters = &domainProducts.Filters{}
	}
	req.Filters.Statuses = availableStatuses

	items, err := i.productsRepo.GetItems(ctx, req)
	if err != nil {
		return items, err
	}
	return items, nil
}

// PreReserveItem резервирует товар для заказа. Возвращает ID забронированного товара и ошибку, если произошла ошибка.
func (i *Implementation) PreReserveItem(ctx context.Context, productID domainProducts.ID) (domainProducts.ItemID, error) {
	reservedItemID, err := i.productsRepo.PreReservedSProduct(ctx, productID)
	if err != nil {
		return domainProducts.ItemID{}, err
	}

	return reservedItemID, nil
}

func (i *Implementation) ReserveItem(ctx context.Context, itemID domainProducts.ItemID) error {
	return i.changeItemStatus(ctx, itemID, domainProducts.ReservedStatus)
}

func (i *Implementation) DereserveItem(ctx context.Context, itemID domainProducts.ItemID) error {
	return i.changeItemStatus(ctx, itemID, domainProducts.SaleStatus)
}

func (i *Implementation) PerformedItem(ctx context.Context, itemID domainProducts.ItemID) error {
	return i.changeItemStatus(ctx, itemID, domainProducts.PerformedStatus)
}

func (i *Implementation) changeItemStatus(ctx context.Context, itemID domainProducts.ItemID, toStatus domainProducts.ItemStatus) error {
	item, err := i.productsRepo.GetItem(ctx, itemID)
	if err != nil {
		return err
	}

	c, err := domainProducts.NewChangeItemStatus(item.Status, toStatus)
	if err != nil {
		if errors.Is(err, domainProducts.ErrStatusIsEqual) {
			return nil
		}

		return err
	}

	err = i.productsRepo.ChangeItemStatus(ctx, itemID, c)
	if err != nil {

		return err
	}

	return nil
}
