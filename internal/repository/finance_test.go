package repository

import (
	"context"
	"testing"
	"time"

	"github.com/chucky-1/finance/internal/model"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

var (
	period      = "today"
	financeRepo *Mongo
)

func TestFinance_AddGet(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	defer func() {
		err := mongoCli.Database("expenses").Collection(period).Drop(ctx)
		if err != nil {
			t.Fatal(err)
		}
	}()

	entry := model.Entry{
		Kind: "expenses",
		Item: "coffee",
		User: "Dima",
		Date: time.Now().UTC(),
		Sum:  3.6,
	}

	err := financeRepo.Add(ctx, &entry, period)
	if err != nil {
		t.Fatal(err)
	}

	data, err := financeRepo.Get(ctx, &model.Entry{
		Kind: entry.Kind,
		User: entry.User,
		Date: entry.Date,
	}, period)
	if err != nil {
		t.Fatal(err)
	}
	logrus.Infof("received entry: %v", data)
	require.Equal(t, 1, len(data))
	require.Equal(t, data["coffee"], 3.6)
}

func TestFinance_AddUpdate(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	defer func() {
		err := mongoCli.Database("expenses").Collection(period).Drop(ctx)
		if err != nil {
			t.Fatal(err)
		}
	}()

	entry1 := model.Entry{
		Kind: "expenses",
		Item: "coffee",
		User: "Dima",
		Date: time.Now().UTC(),
		Sum:  3.5,
	}
	entry2 := model.Entry{
		Kind: "expenses",
		Item: "coffee",
		User: "Dima",
		Date: time.Now().UTC(),
		Sum:  2,
	}
	entry3 := model.Entry{
		Kind: "expenses",
		Item: "rent",
		User: "Dima",
		Date: time.Now().UTC(),
		Sum:  560,
	}
	for _, entry := range []*model.Entry{
		&entry1, &entry2, &entry3,
	} {
		err := financeRepo.Add(ctx, entry, period)
		if err != nil {
			t.Fatal(err)
		}
	}

	data, err := financeRepo.Get(ctx, &model.Entry{
		Kind: entry1.Kind,
		User: entry1.User,
		Date: entry1.Date,
	}, period)
	if err != nil {
		t.Fatal(err)
	}
	logrus.Infof("received entry: %v", data)
	require.Equal(t, 2, len(data))
	require.Equal(t, data["coffee"], entry1.Sum+entry2.Sum)
	require.Equal(t, data["rent"], entry3.Sum)
}

func TestFinance_GetByUsers(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	defer func() {
		err := mongoCli.Database("expenses").Collection(period).Drop(ctx)
		if err != nil {
			t.Fatal(err)
		}
	}()

	beginningOfDay := time.Now().UTC().Truncate(24 * time.Hour)

	entry1 := model.Entry{
		Kind: "expenses",
		Item: "coffee",
		User: "Dima",
		Date: beginningOfDay.Add(-time.Hour),
		Sum:  3.6,
	}
	entry2 := model.Entry{
		Kind: "expenses",
		Item: "Rent",
		User: "Pasha",
		Date: beginningOfDay.Add(10 * -time.Hour),
		Sum:  560,
	}
	entry3 := model.Entry{
		Kind: "expenses",
		Item: "Food",
		User: "Luisa",
		Date: beginningOfDay.Add(5 * -time.Hour),
		Sum:  20.40,
	}
	entry4 := model.Entry{
		Kind: "expenses",
		Item: "Drags",
		User: "Bad user",
		Date: beginningOfDay.Add(7 * -time.Hour),
		Sum:  999.9,
	}
	for _, e := range []*model.Entry{
		&entry1, &entry2, &entry3, &entry4,
	} {
		err := financeRepo.Add(ctx, e, period)
		if err != nil {
			t.Fatal(err)
		}
	}

	users := []string{
		entry1.User, entry2.User, entry3.User,
	}
	entries, err := financeRepo.GetByUsernames(ctx, users, "expenses", period)
	if err != nil {
		t.Fatal(err)
	}
	require.Equal(t, 3, len(entries))
	require.Equal(t, entries[entry1.User][entry1.Item], entry1.Sum)
	require.Equal(t, entries[entry2.User][entry2.Item], entry2.Sum)
	require.Equal(t, entries[entry3.User][entry3.Item], entry3.Sum)
}

func TestMongo_DeleteByUsers(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	defer func() {
		err := mongoCli.Database("expenses").Collection(period).Drop(ctx)
		if err != nil {
			t.Fatal(err)
		}
	}()

	beginningOfDay := time.Now().UTC().Truncate(24 * time.Hour)

	entry1 := model.Entry{
		Kind: "expenses",
		Item: "coffee",
		User: "Dima",
		Date: beginningOfDay.Add(-time.Hour),
		Sum:  3.6,
	}
	entry2 := model.Entry{
		Kind: "expenses",
		Item: "Rent",
		User: "Pasha",
		Date: beginningOfDay.Add(10 * -time.Hour),
		Sum:  560,
	}
	entry3 := model.Entry{
		Kind: "expenses",
		Item: "Food",
		User: "Luisa",
		Date: beginningOfDay.Add(5 * -time.Hour),
		Sum:  20.40,
	}
	entry4 := model.Entry{
		Kind: "expenses",
		Item: "Drags",
		User: "Bad user",
		Date: beginningOfDay.Add(7 * -time.Hour),
		Sum:  999.9,
	}
	for _, e := range []*model.Entry{
		&entry1, &entry2, &entry3, &entry4,
	} {
		err := financeRepo.Add(ctx, e, period)
		if err != nil {
			t.Fatal(err)
		}
	}

	users := []string{
		entry1.User, entry2.User, entry3.User,
	}

	entries, err := financeRepo.GetByUsernames(ctx, users, "expenses", period)
	if err != nil {
		t.Fatal(err)
	}
	require.Equal(t, 3, len(entries))

	err = financeRepo.DeleteByUsernames(ctx, users, "expenses", period)
	if err != nil {
		t.Fatal(err)
	}

	entries, err = financeRepo.GetByUsernames(ctx, users, "expenses", period)
	if err != nil {
		t.Fatal(err)
	}
	require.Equal(t, 0, len(entries))
}
