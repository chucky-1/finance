package consumer

import (
	"context"
	"github.com/chucky-1/finance/internal/service"
	"github.com/sirupsen/logrus"
	"time"
)

type Cleaner struct {
	cleaner service.Cleaner
}

func NewCleaner(cleaner service.Cleaner) *Cleaner {
	return &Cleaner{
		cleaner: cleaner,
	}
}

func (c *Cleaner) Consume(ctx context.Context) {
	logrus.Info("cleaner consumer started")

	t, err := c.createTickerFromBeginningOfTheDay(ctx)
	if err != nil {
		logrus.Infof("cleaner consumer stopped: %v", err)
		return
	}

	newCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	err = c.cleaner.ClearYesterday(newCtx)
	if err != nil {
		logrus.Errorf("cleaner consumer couldn't ClearYesterday: %v", err)
		cancel()
	} else {
		cancel()
		logrus.Infof("cleaner consumer successfully ClearYesterday in %v", time.Now().UTC())
	}

	for {
		select {
		case <-ctx.Done():
			logrus.Infof("cleaner consumer stopped: %v", ctx.Err())
			return
		case <-t.C:
			if time.Now().UTC().Hour() != 0 {
				continue
			}
			newCtx, cancel = context.WithTimeout(ctx, 10*time.Second)
			err = c.cleaner.ClearYesterday(newCtx)
			if err != nil {
				logrus.Errorf("cleaner consumer couldn't ClearYesterday: %v", err)
				cancel()
				continue
			}
			cancel()
			logrus.Infof("cleaner consumer successfully ClearYesterday in: %v", time.Now().UTC())
			continue
		}
	}
}

func (c *Cleaner) createTickerFromBeginningOfTheDay(ctx context.Context) (*time.Ticker, error) {
	t := time.NewTicker(time.Second)
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-t.C:
			if time.Now().UTC().Hour() != 0 {
				continue
			}
			return time.NewTicker(time.Hour), nil
		}
	}
}
