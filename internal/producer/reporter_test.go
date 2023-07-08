package producer

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"strconv"
	"strings"
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

func Test_ConvertToTGReports(t *testing.T) {
	testTable := []struct {
		name       string
		title      string
		categories map[string]float64
	}{
		{
			name:  "Simple",
			title: "8 Июля\n",
			categories: map[string]float64{
				"Food":  25.6,
				"Relax": 30,
				"Rent":  560,
			},
		},
		{
			name:  "With sub categories",
			title: "8 Июля\n",
			categories: map[string]float64{
				"Food.Amount":                    0,
				"Food.InHouse.Amount":            23.9,
				"Food.Outside.Amount":            66,
				"Food.Outside.Restaurant.Amount": 99,
				"Food.Outside.FastFoof.Amount":   12,
				"Relax.Amount":                   120,
				"Alcohol.Amount":                 33,
				"Drink.Amount":                   17,
			},
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			report := convertToTGReport(testCase.title, testCase.categories)
			fmt.Println(report)
			var total float64
			for _, v := range testCase.categories {
				total += v
			}
			sl := strings.Split(report, "Итого - ")
			require.Equal(t, 2, len(sl))
			ttl, err := strconv.ParseFloat(sl[1], 64)
			if err != nil {
				t.Fatal(err)
			}
			require.Equal(t, total, ttl)
		})
	}
}
