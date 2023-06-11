package service

import (
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestTimezone_InThePositiveTimezone(t *testing.T) {
	tz := timezones{
		timezones: make(map[int][]string),
	}
	tz.timezones[3] = append(tz.timezones[3], "Dima")
	tz.timezones[2] = append(tz.timezones[2], "Pasha")
	tz.timezones[4] = append(tz.timezones[4], "Luisa")
	tz.timezones[-3] = append(tz.timezones[-3], "Elena")

	users := tz.getIfTheDayChanges(21)
	logrus.Info(users)
	require.Equal(t, 1, len(users))
	require.Equal(t, "Dima", users[0])
}

func TestTimezone_InTheNegativeTimezone(t *testing.T) {
	tz := timezones{
		timezones: make(map[int][]string),
	}
	tz.timezones[3] = append(tz.timezones[3], "Dima")
	tz.timezones[2] = append(tz.timezones[2], "Pasha")
	tz.timezones[4] = append(tz.timezones[4], "Luisa")
	tz.timezones[-3] = append(tz.timezones[-3], "Elena")

	users := tz.getIfTheDayChanges(3)
	logrus.Info(users)
	require.Equal(t, 1, len(users))
	require.Equal(t, "Elena", users[0])
}

func TestTimezone_In12ByUTC(t *testing.T) {
	tz := timezones{
		timezones: make(map[int][]string),
	}
	tz.timezones[3] = append(tz.timezones[3], "Dima")
	tz.timezones[2] = append(tz.timezones[2], "Pasha")
	tz.timezones[4] = append(tz.timezones[4], "Luisa")
	tz.timezones[-3] = append(tz.timezones[-3], "Elena")
	tz.timezones[12] = append(tz.timezones[12], "Petrov")
	tz.timezones[-12] = append(tz.timezones[-12], "Julia")

	users := tz.getIfTheDayChanges(12)
	logrus.Info(users)
	require.Equal(t, 2, len(users))
	require.Equal(t, "Petrov", users[0])
	require.Equal(t, "Julia", users[1])
}
