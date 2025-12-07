package user

import (
	"context"
	"database/sql"
	"errors"
	"strconv"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	domain "github.com/kdv2001/onlySubscription/internal/domain/user"
	custom_errors "github.com/kdv2001/onlySubscription/pkg/errors"
)

type Implementation struct {
	c *pgxpool.Pool
}

var usersTable = `create table if not exists users (	
    id     uuid primary key,
    state varchar NOT NULL
)`

var authBotTelegramTable = `create table if not exists auth_bot_telegram(
    id            uuid primary key,
    user_id       uuid NOT NULL, FOREIGN KEY (user_id)  REFERENCES users (id),
    telegram_id   varchar NOT NULL UNIQUE,
    created_at    timestamp WITHOUT TIME ZONE NOT NULL DEFAULT (NOW() AT TIME ZONE 'UTC'),
    state         varchar NOT NULL,
    chat_id       varchar NOT NULL)                       
    `

var tables = []string{
	usersTable,
	authBotTelegramTable,
}

// NewImplementation создает объект репо
func NewImplementation(ctx context.Context, c *pgxpool.Pool) (*Implementation, error) {
	for _, t := range tables {
		_, err := c.Exec(ctx, t)
		if err != nil {
			return nil, nil
		}
	}

	return &Implementation{
		c: c,
	}, nil
}

type user struct {
	UserID sql.NullString
	ChatID sql.NullString
}

type telegramBotAuth struct {
	ID         sql.NullString
	UserID     sql.NullString
	TelegramID sql.NullString
	CreatedAt  sql.NullTime
}

func (repo *Implementation) RegisterTelegram(ctx context.Context, telegramBotLogin domain.TelegramBotRegister) (domain.ID, error) {
	tx, err := repo.c.Begin(ctx)
	if err != nil {
		return domain.ID{}, err
	}
	defer func() {
		if err != nil {
			// TODO  обработать ошибку
			_ = tx.Rollback(ctx)
		}
	}()

	uid := uuid.New()
	_, err = tx.Exec(ctx, `INSERT INTO users(id, state) values($1, $2);`, uid.String(), domain.VerifiedState.String())
	if err != nil {
		return domain.ID{}, err
	}

	authUid := uuid.New()
	_, err = tx.Exec(ctx, `INSERT INTO auth_bot_telegram(id, user_id, telegram_id, state, chat_id) 
      values($1, $2, $3, $4, $5);`, authUid.String(),
		uid.String(),
		telegramBotLogin.TelegramID,
		domain.VerifiedState.String(),
		strconv.FormatInt(telegramBotLogin.ChatID, 10))
	if err != nil {
		return domain.ID{}, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return domain.ID{}, err
	}

	return domain.NewID(uid.String()), nil
}

func (repo *Implementation) GetUser(ctx context.Context,
	id domain.ID) (domain.User, error) {

	u := user{}
	err := repo.c.QueryRow(ctx, `select users.id, auth_bot_telegram.chat_id
	from users left join auth_bot_telegram on
    auth_bot_telegram.user_id = users.id 
          where users.id = $1`, id).
		Scan(
			&u.UserID,
			&u.ChatID,
		)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, custom_errors.NewNotFoundError(err)
		}
		return domain.User{}, err
	}

	cID, err := strconv.ParseInt(u.ChatID.String, 10, 64)
	if err != nil {
		return domain.User{}, custom_errors.NewInternalError(errors.New("invalid chatID"))
	}

	return domain.User{
		ID: domain.NewID(u.UserID.String),
		Contact: domain.Contact{
			TelegramBotChatID: cID,
		},
	}, nil
}

func (repo *Implementation) GetUserByTelegramID(ctx context.Context,
	tgID domain.TelegramID) (domain.User, error) {

	u := user{}
	err := repo.c.QueryRow(ctx, `select users.id, auth_bot_telegram.chat_id
	from users left join auth_bot_telegram on
    auth_bot_telegram.user_id = users.id 
          where telegram_id = $1`, tgID).
		Scan(
			&u.UserID,
			&u.ChatID,
		)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, custom_errors.NewNotFoundError(err)
		}
		return domain.User{}, err
	}

	cID, err := strconv.ParseInt(u.ChatID.String, 10, 64)
	if err != nil {
		return domain.User{}, custom_errors.NewInternalError(errors.New("invalid chatID"))
	}

	return domain.User{
		ID: domain.NewID(u.UserID.String),
		Contact: domain.Contact{
			TelegramBotChatID: cID,
		},
	}, nil
}
