package service

import (
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestTimezone_GetUsersWhoseDayChangesInThePositiveTimezone(t *testing.T) {
	tz := timezones{
		timezones: make(map[time.Duration][]string),
	}
	tz.add(3*time.Hour, "Dima")
	tz.add(2*time.Hour, "Pasha")
	tz.add(4*time.Hour, "Luisa")
	tz.add(-3*time.Hour, "Elena")

	users := tz.getUsersWhoseDayChanges(time.Now().UTC().Truncate(24 * time.Hour).Add(21 * time.Hour))
	logrus.Info(users)
	require.Equal(t, 1, len(users))
	require.Equal(t, "Dima", users[0])
}

func TestTimezone_GetUsersWhoseDayChangesInTheNegativeTimezone(t *testing.T) {
	tz := timezones{
		timezones: make(map[time.Duration][]string),
	}
	tz.add(3*time.Hour, "Dima")
	tz.add(2*time.Hour, "Pasha")
	tz.add(4*time.Hour, "Luisa")
	tz.add(-3*time.Hour, "Elena")

	users := tz.getUsersWhoseDayChanges(time.Now().UTC().Truncate(24 * time.Hour).Add(3 * time.Hour))
	logrus.Info(users)
	require.Equal(t, 1, len(users))
	require.Equal(t, "Elena", users[0])
}

func TestTimezone_GetUsersWhoseDayChangesIn12ByUTC(t *testing.T) {
	tz := timezones{
		timezones: make(map[time.Duration][]string),
	}
	tz.add(3*time.Hour, "Dima")
	tz.add(2*time.Hour, "Pasha")
	tz.add(4*time.Hour, "Luiza")
	tz.add(-3*time.Hour, "Elena")
	tz.add(12*time.Hour, "Petrov")
	tz.add(-12*time.Hour, "Julia")

	// The day changes in 12:00 by UTC in timezones +12 and -12
	users := tz.getUsersWhoseDayChanges(time.Now().UTC().Truncate(24 * time.Hour).Add(12 * time.Hour))
	logrus.Info(users)
	require.Equal(t, 2, len(users))
	require.Equal(t, "Petrov", users[0])
	require.Equal(t, "Julia", users[1])
}

func TestTimezone_GetUsersWhoseDayChangesIn12ByUTCUserOnlyInPositiveTimezone(t *testing.T) {
	tz := timezones{
		timezones: make(map[time.Duration][]string),
	}
	tz.add(12*time.Hour, "Julia")

	// The day changes in 12:00 by UTC in timezones +12 and -12
	users := tz.getUsersWhoseDayChanges(time.Now().UTC().Truncate(24 * time.Hour).Add(12 * time.Hour))
	logrus.Info(users)
	require.Equal(t, 1, len(users))
	require.Equal(t, "Julia", users[0])
}

func TestTimezone_GetUsersWhoseDayChangesIn12ByUTCWithoutUsers(t *testing.T) {
	tz := timezones{
		timezones: make(map[time.Duration][]string),
	}
	require.Equal(t, 0, len(tz.getUsersWhoseDayChanges(time.Now().UTC().Truncate(24*time.Hour).Add(12*time.Hour))))
}

func TestTimezone_GetUsersWhoseDayChangesSriLanka(t *testing.T) {
	tz := timezones{
		timezones: make(map[time.Duration][]string),
	}
	tz.add((5*time.Hour)+(30*time.Minute), "Dima")
	users := tz.getUsersWhoseDayChanges(time.Now().UTC().Truncate(24 * time.Hour).Add(18*time.Hour + 30*time.Minute))
	logrus.Info(users)
	require.Equal(t, 1, len(users))
	require.Equal(t, "Dima", users[0])
}

func TestTimezone_GetUsersWhoseMonthChanges(t *testing.T) {
	tz := timezones{
		timezones: make(map[time.Duration][]string),
	}
	tz.add(3*time.Hour, "Dima")

	testTable := []struct {
		name    string
		timeUTC time.Time
		result  []string
	}{
		{
			name:    "Success",
			timeUTC: time.Date(2023, 6, 30, 21, 0, 0, 0, time.UTC),
			result: []string{
				"Dima",
			},
		},
		{
			name:    "Day changes but month not",
			timeUTC: time.Date(2023, 6, 29, 21, 0, 0, 0, time.UTC),
			result:  nil,
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			require.Equal(t, testCase.result, tz.getUsersWhoseMonthChanges(testCase.timeUTC))
		})
	}
}

func TestTimezone_GetUsersWhoseMonthChanges12Timezone(t *testing.T) {
	tz := timezones{
		timezones: make(map[time.Duration][]string),
	}
	tz.add(12*time.Hour, "Dima")
	tz.add(-12*time.Hour, "Liza")

	testTable := []struct {
		name    string
		timeUTC time.Time
		result  []string
	}{
		{
			name:    "Success for positive timezone",
			timeUTC: time.Date(2023, 6, 30, 12, 0, 0, 0, time.UTC),
			result: []string{
				"Dima",
			},
		},
		{
			name:    "Success for negative timezone",
			timeUTC: time.Date(2023, 6, 1, 12, 0, 0, 0, time.UTC),
			result: []string{
				"Liza",
			},
		},
		{
			name:    "Day changes but month not",
			timeUTC: time.Date(2023, 6, 29, 12, 0, 0, 0, time.UTC),
			result:  nil,
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			require.Equal(t, testCase.result, tz.getUsersWhoseMonthChanges(testCase.timeUTC))
		})
	}
}

// TestTimezone_GetTimezoneWhereDayChangesInThePositiveTimezone the day changes in 21:00 by UTC in time zone +3
func TestTimezone_GetTimezoneWhereDayChangesInThePositiveTimezone(t *testing.T) {
	require.Equal(t, getTimezoneWhereDayChanges(21, 0), 3*time.Hour)
}

// TestTimezone_GetTimezoneWhereDayChangesInTheNegativeTimezone the day changes in 3:00 by UTC in time zone -3
func TestTimezone_GetTimezoneWhereDayChangesInTheNegativeTimezone(t *testing.T) {
	require.Equal(t, getTimezoneWhereDayChanges(3, 0), -3*time.Hour)
}

// TestTimezone_GetTimezoneWhereDayChangesSriLanka the day changes in 18:30 by UTC in time zone +5:30
func TestTimezone_GetTimezoneWhereDayChangesSriLanka(t *testing.T) {
	require.Equal(t, getTimezoneWhereDayChanges(18, 30), 5*time.Hour+30*time.Minute)
}

// TestTimezone_GetTimezoneWhereDayChanges12 the day changes in 12:00 by UTC in time zones +12 and -12, but func returns only +12
func TestTimezone_GetTimezoneWhereDayChanges12(t *testing.T) {
	require.Equal(t, getTimezoneWhereDayChanges(12, 0), 12*time.Hour)
}

func TestTimezone_AddGet(t *testing.T) {
	tz := timezones{
		timezones: make(map[time.Duration][]string),
	}
	tz.add(3*time.Hour, "Dima")
	require.Equal(t, tz.timezones[3*time.Hour], tz.get(3*time.Hour))
}

func TestTimezone_GetEmptyResult(t *testing.T) {
	tz := timezones{
		timezones: make(map[time.Duration][]string),
	}
	require.Equal(t, 0, len(tz.get(3*time.Hour)))
}
