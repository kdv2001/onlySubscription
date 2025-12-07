package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	uuid2 "github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/kdv2001/onlySubscription/internal/domain/price"
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
func (i *Implementation) GetProducts(ctx context.Context, req domainProducts.RequestList) (domainProducts.Products, error) {
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
				SET record_status = $1, updated_at = NOW() AT TIME ZONE 'UTC'
				WHERE uuid_eq(id, $2);`, deleteRecord, id.String())
	if err != nil {
		return err
	}

	return nil
}

func (i *Implementation) CreateInventoryItem(ctx context.Context, req domainProducts.Item) (domainProducts.ItemID, error) {
	uuid := uuid2.New()
	_, err := i.conn.Exec(ctx, `insert into inventory (id, description, product_id)
	values ($1, $2, $3)`, uuid, req.Payload, req.ProductID)
	if err != nil {
		return domainProducts.ItemID{}, err
	}

	return domainProducts.NewItemID(uuid.String()), nil
}

func (i *Implementation) DeleteInventoryItem(ctx context.Context, id domainProducts.ItemID) error {
	_, err := i.conn.Exec(ctx, `delete from inventory
			
				WHERE uuid_eq(id, $1) and status = $2;`, id.String(), domainProducts.SaleStatus.String())
	if err != nil {
		return err
	}

	return nil
}

// PreReservedSProduct пререзервирует единицу товара для продажи по айди продукта
func (i *Implementation) PreReservedSProduct(ctx context.Context,
	productID domainProducts.ID,

) (domainProducts.ItemID, error) {
	id := sql.NullString{}
	err := i.transaction(ctx, func(tx pgx.Tx) error {
		// находим единицу товара с указанным ID и статусом "Продается"

		err := tx.QueryRow(ctx, `select id from inventory where uuid_eq(product_id, $1)
                          and  status = $2 limit 1 for update;`, productID, domainProducts.SaleStatus.String()).
			Scan(&id)
		if err != nil {
			return custom_errors.NewInternalError(err)
		}

		// Бронируем единицу товара
		_, err = tx.Exec(ctx, `update inventory set status = $1, 
                     updated_at = NOW() AT TIME ZONE 'UTC' where id = $2;`,
			domainProducts.PreReservedStatus, id)
		if err != nil {
			return custom_errors.NewInternalError(err)
		}

		return nil
	})
	if err != nil {
		return domainProducts.ItemID{}, nil
	}

	return domainProducts.NewItemID(id.String), nil
}

// ChangeItemStatus бронирует товар.
func (i *Implementation) ChangeItemStatus(ctx context.Context,
	itemID domainProducts.ItemID,
	changeItemStatus domainProducts.ChangeItemStatus,
) error {
	return i.transaction(ctx, func(tx pgx.Tx) error {
		id := sql.NullString{}
		err := tx.QueryRow(ctx, `select id from inventory where uuid_eq(id, $1)
                        	limit 1 for update;`, itemID).
			Scan(&id)
		if err != nil {
			return custom_errors.NewInternalError(err)
		}

		t, err := tx.Exec(ctx, `update inventory set status = $1, updated_at = NOW() AT TIME ZONE 'UTC'
                 where uuid_eq(id, $2) and status = $3;`,
			changeItemStatus.To, itemID, changeItemStatus.From)
		if err != nil {
			return custom_errors.NewInternalError(err)
		}

		if t.RowsAffected() == 0 {
			return custom_errors.NewBadRequestError(errors.New("booking is expired"))
		}

		return nil
	})
}

func (i *Implementation) GetItems(
	ctx context.Context,
	req domainProducts.RequestList,
) ([]domainProducts.Item, error) {
	query := `select *from inventory`
	values := []any{}

	if req.Filters != nil {
		query += " where "
		if req.Filters.Statuses != nil {
			query += " status in ("
			for j, s := range req.Filters.Statuses {
				if j != 0 {
					query += ", "
				}
				values = append(values, s.String())
				query += `$` + fmt.Sprint(len(values))
			}
			query += ")"
		}

		if req.Filters.UpdatedAt != nil {
			if !req.Filters.UpdatedAt.From.IsZero() {
				if len(values) != 0 {
					query += ` and`
				}
				values = append(values, req.Filters.UpdatedAt.From.UTC())
				query += ` updated_at > $` +
					fmt.Sprint(len(values))
			}
			if !req.Filters.UpdatedAt.To.IsZero() {
				if len(values) != 0 {
					query += ` and`
				}
				values = append(values, req.Filters.UpdatedAt.To.UTC())
				query += ` updated_at < $` +
					fmt.Sprint(len(values))
			}
		}

		if req.Filters.ProductID.String() != "" {
			if len(values) != 0 {
				query += ` and`
			}
			values = append(values, req.Filters.ProductID.String())
			query += ` product_id = $` +
				fmt.Sprint(len(values))
		}
	}

	if req.Pagination != nil {
		if req.Pagination.Offset != 0 {
			values = append(values, req.Pagination.Offset)
			query += ` offset $` + fmt.Sprint(len(values))
		}
		if req.Pagination.Num != 0 {
			values = append(values, req.Pagination.Num)
			query += ` limit $` + fmt.Sprint(len(values))
		}
	}

	res, err := i.conn.Query(ctx, query, values...)
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
			ID:        domainProducts.NewItemID(curItem.ID.String),
			ProductID: domainProducts.NewID(curItem.ProductID.String),
			Status:    domainProducts.ItemStatusFromString(curItem.Status.String),
			CreatedAt: curItem.CreatedAt.Time,
			UpdatedAt: curItem.UpdatedAt.Time,
			Payload:   curItem.Description.String,
		})
	}

	return itemsResult, nil
}

func (i *Implementation) GetItem(ctx context.Context, id domainProducts.ItemID) (domainProducts.Item, error) {
	var p item
	err := i.conn.QueryRow(ctx, `select * from inventory where uuid_eq(id, $1);`, id.String()).Scan(
		&p.ID,
		&p.ProductID,
		&p.Status,
		&p.CreatedAt,
		&p.UpdatedAt,
		&p.Description)
	if err != nil {
		return domainProducts.Item{}, custom_errors.NewInternalError(err)
	}

	return domainProducts.Item{
		ID:        domainProducts.NewItemID(p.ID.String),
		ProductID: domainProducts.NewID(p.ProductID.String),
		Status:    domainProducts.ItemStatusFromString(p.Status.String),
		CreatedAt: p.CreatedAt.Time,
		UpdatedAt: p.UpdatedAt.Time,
		Payload:   p.Description.String,
	}, nil
}

func (i *Implementation) transaction(ctx context.Context, fnc func(tx pgx.Tx) error) error {
	tx, err := i.conn.Begin(ctx)
	if err != nil {
		return custom_errors.NewInternalError(err)
	}
	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		}
	}()

	err = fnc(tx)
	if err != nil {
		return err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return custom_errors.NewInternalError(err)
	}

	return nil
}

func (i *Implementation) CountItemsForProduct(ctx context.Context, productID domainProducts.ID) (int64, error) {
	num := sql.NullInt64{}
	err := i.conn.QueryRow(ctx, `select count(*) from inventory where uuid_eq(product_id, $1) and status = $2`, productID, domainProducts.SaleStatus).Scan(
		&num)
	if err != nil {
		return 0, custom_errors.NewInternalError(err)
	}

	return num.Int64, nil
}
