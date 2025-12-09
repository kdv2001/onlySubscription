package parallel

import (
	"context"
	"sync"
	"time"

	"github.com/kdv2001/onlySubscription/pkg/logger"
)

func BackgroundPeriodProcess(ctx context.Context,
	wg *sync.WaitGroup,
	dur time.Duration,
	fcn func(ctx context.Context) error) {
	wg.Add(1)
	defer wg.Done()

	t := time.NewTicker(dur)
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			err := panicWrapper(ctx, fcn)
			if err != nil {
				logger.Errorf(ctx, "error: %v", err)
			}
		}
	}
}

func panicWrapper(ctx context.Context, fcn func(ctx context.Context) error) error {
	defer func() {
		erri := recover()
		if erri != nil {
			logger.Errorf(ctx, "panic: %v", erri)
		}
	}()
	return fcn(ctx)
}
