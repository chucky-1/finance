package service

import (
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestTimezone_InThePositiveTimezone(t *testing.T) {
	tz := timezones{
		timezones: make(map[time.Duration][]string),
	}
	tz.add(3*time.Hour, "Dima")
	tz.add(2*time.Hour, "Pasha")
	tz.add(4*time.Hour, "Luisa")
	tz.add(-3*time.Hour, "Elena")

	// The day changes in 21:00 by UTC in timezone +3
	users := tz.getIfTheDayChanges(21 * time.Hour)
	logrus.Info(users)
	require.Equal(t, 1, len(users))
	require.Equal(t, "Dima", users[0])
}

func TestTimezone_InTheNegativeTimezone(t *testing.T) {
	tz := timezones{
		timezones: make(map[time.Duration][]string),
	}
	tz.add(3*time.Hour, "Dima")
	tz.add(2*time.Hour, "Pasha")
	tz.add(4*time.Hour, "Luisa")
	tz.add(-3*time.Hour, "Elena")

	// The day changes in 3:00 by UTC in timezone -3
	users := tz.getIfTheDayChanges(3 * time.Hour)
	logrus.Info(users)
	require.Equal(t, 1, len(users))
	require.Equal(t, "Elena", users[0])
}

func TestTimezone_In12ByUTC(t *testing.T) {
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
	users := tz.getIfTheDayChanges(12 * time.Hour)
	logrus.Info(users)
	require.Equal(t, 2, len(users))
	require.Equal(t, "Petrov", users[0])
	require.Equal(t, "Julia", users[1])
}

func TestTimezone_SriLanka(t *testing.T) {
	tz := timezones{
		timezones: make(map[time.Duration][]string),
	}
	tz.add((5*time.Hour)+(30*time.Minute), "Dima")
	users := tz.getIfTheDayChanges((18 * time.Hour) + (30 * time.Minute))
	logrus.Info(users)
	require.Equal(t, 1, len(users))
	require.Equal(t, "Dima", users[0])
}
