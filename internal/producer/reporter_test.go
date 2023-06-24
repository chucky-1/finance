package producer

import (
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func Test_TimerBeforeCreateTicker(t *testing.T) {
	testTable := []struct {
		name    string
		timeUTC time.Time
		result  time.Duration
	}{
		{
			name:    "Before 30 minutes",
			timeUTC: time.Date(2023, 6, 28, 15, 15, 0, 0, time.UTC),
			result:  15 * time.Minute,
		},
		{
			name:    "After 30 minutes",
			timeUTC: time.Date(2023, 6, 28, 15, 40, 0, 0, time.UTC),
			result:  20 * time.Minute,
		},
		{
			name:    "In 30 minutes",
			timeUTC: time.Date(2023, 6, 28, 15, 30, 0, 0, time.UTC),
			result:  30 * time.Minute,
		},
		{
			name:    "In 0 minute",
			timeUTC: time.Date(2023, 6, 28, 15, 0, 0, 0, time.UTC),
			result:  30 * time.Minute,
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			require.Equal(t, testCase.result, durationBeforeCreateTicker(testCase.timeUTC))
		})
	}
}
