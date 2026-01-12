package consts

import "time"

const (
	// order
	// DefaultOrderTimeLimit время жизни заказа в минутах
	DefaultOrderTimeLimit = 15 * time.Minute

	// DefaultPrereservedTTL время жизни пререзервации в минутах
	DefaultPrereservedTTL = 1 * time.Minute

	// payment
	// DefaultHandlingDur время обработки платежа в минутах после
	// которого неоюходимо запросить статус платежа
	DefaultHandlingDur = 1 * time.Minute
)
