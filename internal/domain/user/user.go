package user

import (
	"fmt"
	"time"
)

type State string

const (
	// UnknownState ...
	UnknownState State = "unknown"
	// VerifiedState ...
	VerifiedState State = "verified"
)

func (s State) String() string {
	return string(s)
}

// ID айди
type ID struct {
	ID string
}

// String строковое представление
func (id ID) String() string {
	return id.ID
}

// NewID создает объект ID
func NewID[T string | int | int64](i T) ID {
	return ID{
		ID: fmt.Sprint(i),
	}
}

type Auth struct {
	UserID ID
}

type User struct {
	ID ID
}

type SessionInfo struct {
	ID        ID
	CreatedAt time.Time
	UserID    ID
	Device    string
	// какая-то еще мета информация
}

type SessionToken struct {
	Token string
}

// TelegramID айди
type TelegramID struct {
	ID string
}

// String строковое представление
func (id TelegramID) String() string {
	return id.ID
}

// NewTelegramID создает объект ID
func NewTelegramID[T string | int | int64](i T) TelegramID {
	return TelegramID{
		ID: fmt.Sprint(i),
	}
}

// TelegramBotRegister авторизация через телеграм
type TelegramBotRegister struct {
	TelegramID TelegramID
}
