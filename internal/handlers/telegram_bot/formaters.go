package telegram_bot

import "github.com/kdv2001/onlySubscription/internal/domain/price"

func currencyToIcon(c price.Currency) string {
	switch c {
	case price.XTR:
		return string('⭐')
	}

	return string('⚙')
}
