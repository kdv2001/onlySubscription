package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	uuid2 "github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/kdv2001/onlySubscription/internal/domain/price"
	"github.com/kdv2001/onlySubscription/internal/domain/primitives"
	domainProducts "github.com/kdv2001/onlySubscription/internal/domain/products"
	custom_errors "github.com/kdv2001/onlySubscription/pkg/errors"
)

type recordStatus string

const (
	activeRecord recordStatus = ""
	deleteRecord recordStatus = "delete"
)

type Implementation struct {
	conn *pgxpool.Pool
}

func NewImplementation(ctx context.Context, conn *pgxpool.Pool) (*Implementation, error) {
	for _, t := range tables {
		_, err := conn.Exec(ctx, t)
		if err != nil {
			return nil, err
		}
	}

	return &Implementation{
		conn: conn,
	}, nil
}

// GetProducts возвращает продукты
func (i *Implementation) GetProducts(ctx context.Context, req primitives.RequestList) (domainProducts.Products, error) {
	rows, errQ := i.conn.Query(ctx, "select *from products where record_status = $1 offset $2 limit $3",
		activeRecord,
		req.Pagination.Offset,
		req.Pagination.Num)
	if errQ != nil {
		return nil, custom_errors.NewInternalError(errQ).AddDetails("error get pagination req")
	}

	products := make(domainProducts.Products, 0, req.Pagination.Num)
	for rows.Next() {
		var p product
		err := rows.Scan(
			&p.ID,
			&p.Name,
			&p.Description,
			&p.CreatedAt,
			&p.UpdatedAt,
			&p.Type,
			&p.RecordStatus,
			&p.Price,
			&p.SubscriptionPeriod)
		if err != nil {
			return nil, custom_errors.NewInternalError(err).AddDetails("error scan row")
		}

		subDur, err := time.ParseDuration(p.SubscriptionPeriod.String)
		if err != nil {
			return nil, custom_errors.NewInternalError(err).AddDetails("error parse dur")
		}

		products = append(products, domainProducts.Product{
			ID:          domainProducts.NewID(p.ID.String),
			Name:        p.Name.String,
			Description: p.Description.String,
			Type:        domainProducts.TypeFromString(p.Type.String),
			Image:       domainProducts.Image{},
			CreatedAt:   p.CreatedAt.Time,
			UpdatedAt:   p.UpdatedAt.Time,
			Price: price.Price{
				Currency: price.XTR,
				Value:    p.Price.Decimal,
			},
			SubscriptionPeriod: subDur,
		})
	}

	return products, nil
}

// GetProduct возвращает продукт по ID
func (i *Implementation) GetProduct(ctx context.Context, id domainProducts.ID) (domainProducts.Product, error) {
	var p product
	err := i.conn.QueryRow(ctx, `select * from products where uuid_eq(id, $1);`, id.String()).Scan(
		&p.ID,
		&p.Name,
		&p.Description,
		&p.CreatedAt,
		&p.UpdatedAt,
		&p.Type,
		&p.RecordStatus,
		&p.Price,
		&p.SubscriptionPeriod)
	if err != nil {
		return domainProducts.Product{}, err
	}

	subDur, err := time.ParseDuration(p.SubscriptionPeriod.String)
	if err != nil {
		return domainProducts.Product{}, custom_errors.NewInternalError(err).AddDetails("error parse dur")
	}

	return domainProducts.Product{
		ID:          domainProducts.NewID(p.ID.String),
		Name:        p.Name.String,
		Description: p.Description.String,
		Type:        domainProducts.TypeFromString(p.Type.String),
		Image:       domainProducts.Image{},
		CreatedAt:   p.CreatedAt.Time,
		UpdatedAt:   p.UpdatedAt.Time,
		Price: price.Price{
			Currency: price.XTR,
			Value:    p.Price.Decimal,
		},
		SubscriptionPeriod: subDur,
	}, nil
}

func (i *Implementation) CreateProduct(ctx context.Context, req domainProducts.Product) (domainProducts.ID, error) {
	uuid := uuid2.New()
	_, err := i.conn.Exec(ctx, `insert into products (id, name, description, type, price, subscription_period)
	values ($1, $2, $3, $4, $5, $6)`,
		uuid.String(),
		req.Name,
		req.Description,
		req.Type.String(),
		req.Price.Value,
		req.SubscriptionPeriod.String())
	if err != nil {
		return domainProducts.ID{}, err
	}

	return domainProducts.NewID(uuid.String()), nil
}

func (i *Implementation) DeleteProduct(ctx context.Context, id domainProducts.ID) error {
	_, err := i.conn.Exec(ctx, `UPDATE products
				SET record_status = $1, updated_at = now()
				WHERE uuid_eq(id, $2);`, deleteRecord, id.String())
	if err != nil {
		return err
	}

	return nil
}

func (i *Implementation) CreateInventoryItem(ctx context.Context, req domainProducts.Item) (domainProducts.ItemID, error) {
	uuid := uuid2.New()
	_, err := i.conn.Exec(ctx, `insert into inventory (id, description, product_id)
	values ($1, $2, $3)`, uuid, req.Description, req.ProductID)
	if err != nil {
		return domainProducts.ItemID{}, err
	}

	return domainProducts.NewItemID(uuid.String()), nil
}

func (i *Implementation) DeleteInventoryItem(ctx context.Context, id domainProducts.ItemID) error {
	_, err := i.conn.Exec(ctx, `UPDATE inventory
				SET status = $1, updated_at = now()
				WHERE uuid_eq(id, $2);`, deleteRecord, id.String())
	if err != nil {
		return err
	}

	return nil
}

// PreReservedSProduct пререзервирует единицу товара для продажи
func (i *Implementation) PreReservedSProduct(ctx context.Context,
	productID domainProducts.ID,
) (domainProducts.ItemID, error) {
	tx, err := i.conn.Begin(ctx)
	if err != nil {
		return domainProducts.ItemID{}, custom_errors.NewInternalError(err)
	}
	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		}
	}()

	// находим единицу товара с указанным ID и статусом "Продается"
	id := sql.NullString{}
	err = tx.QueryRow(ctx, `select id from inventory where uuid_eq(product_id, $1)
                          and  status = $2 limit 1 for update;`, productID, domainProducts.SaleStatus.String()).
		Scan(&id)
	if err != nil {
		return domainProducts.ItemID{}, custom_errors.NewInternalError(err)
	}

	// Бронируем единицу товара
	_, err = tx.Exec(ctx, `update inventory set status = $1, updated_at = now() where id = $2;`,
		domainProducts.PreReservedStatus, id)
	if err != nil {
		return domainProducts.ItemID{}, custom_errors.NewInternalError(err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return domainProducts.ItemID{}, custom_errors.NewInternalError(err)
	}

	return domainProducts.NewItemID(id.String), nil
}

// ChangeItemStatus бронирует товар.
func (i *Implementation) ChangeItemStatus(ctx context.Context,
	itemID domainProducts.ItemID,
	changeItemStatus domainProducts.ChangeItemStatus,
) error {
	tx, err := i.conn.Begin(ctx)
	if err != nil {
		return custom_errors.NewInternalError(err)
	}
	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		}
	}()

	id := sql.NullString{}
	err = tx.QueryRow(ctx, `select id from inventory where uuid_eq(id, $1)
                        	limit 1 for update;`, itemID).
		Scan(&id)
	if err != nil {
		return custom_errors.NewInternalError(err)
	}

	t, err := tx.Exec(ctx, `update inventory set status = $1, updated_at = now()
                 where uuid_eq(id, $2) and status = $3;`,
		changeItemStatus.To, itemID, changeItemStatus.From)
	if err != nil {
		return custom_errors.NewInternalError(err)
	}

	if t.RowsAffected() == 0 {
		return custom_errors.NewBadRequestError(errors.New("booking is expired"))
	}

	err = tx.Commit(ctx)
	if err != nil {
		return custom_errors.NewInternalError(err)
	}

	return nil
}

func (i *Implementation) GetExpiredPreReservedItems(ctx context.Context, num int) ([]domainProducts.Item, error) {
	res, err := i.conn.Query(ctx, `select * from inventory where status = $1
                          and updated_at + interval '15 mins' < now() limit $2;`, domainProducts.PreReservedStatus, num)
	if err != nil {
		return nil, err
	}

	itemsResult := make([]domainProducts.Item, 0)
	for res.Next() {
		var curItem item
		err = res.Scan(
			&curItem.ID,
			&curItem.ProductID,
			&curItem.Status,
			&curItem.CreatedAt,
			&curItem.UpdatedAt,
			&curItem.Description,
		)
		if err != nil {
			return nil, custom_errors.NewInternalError(err)
		}

		itemsResult = append(itemsResult, domainProducts.Item{
			ID:          domainProducts.NewItemID(curItem.ID.String),
			ProductID:   domainProducts.NewID(curItem.ProductID.String),
			Status:      domainProducts.ItemStatusFromString(curItem.Status.String),
			CreatedAt:   curItem.CreatedAt.Time,
			UpdatedAt:   curItem.UpdatedAt.Time,
			Description: curItem.Description.String,
		})
	}

	return itemsResult, nil
}

func (i *Implementation) GetItem(ctx context.Context, id domainProducts.ItemID) (domainProducts.Item, error) {
	return domainProducts.Item{}, nil
}
