package price

import "github.com/shopspring/decimal"

type Currency string

const (
	UNKNOWN Currency = ""
	// RUB рубли
	RUB Currency = "RUB"
	// XTR телеграм звезды
	XTR Currency = "XTR"
)

type Price struct {
	Currency Currency
	Value    decimal.Decimal
}

func (c Currency) String() string {
	return string(c)
}

func CurrencyFromString(str string) Currency {
	switch str {
	case string(RUB):
		return RUB
	case string(XTR):
		return XTR
	}

	return UNKNOWN
}
