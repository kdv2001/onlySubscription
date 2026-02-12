package user

import (
	"fmt"
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

// IsEmpty возвращает признак пустого ID
func (id ID) IsEmpty() bool {
	return id.ID == ""
}

// NewID создает объект ID
func NewID[T string | int | int64](i T) ID {
	return ID{
		ID: fmt.Sprint(i),
	}
}

// Auth авторизация
type Auth struct {
	UserID ID
}

// User пользователь
type User struct {
	// ID пользователя
	ID ID
	// Contact контактные данные
	Contact Contact
}

// Contact контакты
type Contact struct {
	// TelegramBotChatID ID телеграм чата
	TelegramBotChatID int64
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
	ChatID     int64
}
