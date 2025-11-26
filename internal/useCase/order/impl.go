package order

import (
	"context"

	"github.com/kdv2001/onlySubscription/internal/domain/order"
	domainProducts "github.com/kdv2001/onlySubscription/internal/domain/products"
	"github.com/kdv2001/onlySubscription/internal/domain/user"
)

type productUC interface {
	SaleProduct(ctx context.Context, itemID domainProducts.ItemID) error
	GetProduct(ctx context.Context, id domainProducts.ID) (domainProducts.Product, error)
	GetItem(ctx context.Context, id domainProducts.ItemID) (domainProducts.Item, error)
	PreReserveItem(ctx context.Context, productID domainProducts.ID) (domainProducts.ItemID, error)
	ReserveProduct(ctx context.Context, itemID domainProducts.ItemID) error
	DereserveProduct(ctx context.Context, itemID domainProducts.ItemID) error
}

type orderRepo interface {
	CreateOrder(ctx context.Context, o order.Order) (order.ID, error)
	GetOrder(ctx context.Context, oID order.ID, userID user.ID) (order.Order, error)
	UpdateOrderStatus(ctx context.Context, oID order.ID, userID user.ID,
		status order.Status) error
}

type Implementation struct {
	productUC productUC
	orderRepo orderRepo
}

func NewImplementation(
	productUC productUC,
	orderRepo orderRepo,
) *Implementation {
	return &Implementation{
		productUC: productUC,
		orderRepo: orderRepo,
	}
}

func (i *Implementation) CreateOrder(ctx context.Context, o order.CreateOrder) (order.ID, error) {
	product, err := i.productUC.GetProduct(ctx, o.ProductID)
	if err != nil {
		return order.ID{}, err
	}

	itemID, err := i.productUC.PreReserveItem(ctx, o.ProductID)
	if err != nil {
		return order.ID{}, err
	}

	orderID, err := i.orderRepo.CreateOrder(ctx, order.Order{
		TotalPrice: product.Price,
		Status:     order.Form,
		UserID:     o.UserID,
		Product: order.Product{
			ItemID:    itemID,
			ProductID: product.ID,
		},
	})
	if err != nil {
		return order.ID{}, err
	}

	err = i.productUC.ReserveProduct(ctx, itemID)
	if err != nil {
		return order.ID{}, err
	}

	err = i.orderRepo.UpdateOrderStatus(ctx, orderID, o.UserID, order.ExpectPayments)
	if err != nil {
		// TODO logger
		return orderID, nil
	}

	return orderID, nil
}

func (i *Implementation) GetOrder(ctx context.Context, oID order.ID, userID user.ID) (order.Order, error) {
	return i.orderRepo.GetOrder(ctx, oID, userID)
}

//func (i *Implementation) recoveryOrder() {
//	i.productUC.GetItem()
//}
