package price

import "github.com/shopspring/decimal"

// Currency Валюта
type Currency string

const (
	UNKNOWN Currency = ""
	// RUB рубли
	RUB Currency = "RUB"
	// XTR телеграм звезды
	XTR Currency = "XTR"
)

// Price стоимость
type Price struct {
	// Currency валюта
	Currency Currency
	// Value денежное значение
	Value decimal.Decimal
}

// String строковое представление
func (c Currency) String() string {
	return string(c)
}

// CurrencyFromString валюта из строки
func CurrencyFromString(str string) Currency {
	switch str {
	case string(RUB):
		return RUB
	case string(XTR):
		return XTR
	}

	return UNKNOWN
}
