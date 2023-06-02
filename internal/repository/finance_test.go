package repository

import (
	"context"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/chucky-1/finance/internal/model"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

var (
	financeRepo *Finance
)

func TestFinance_AddGet(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	defer func() {
		_, err := mongoCli.Database("expenses").Collection(time.Now().UTC().Format(layout)).DeleteMany(ctx,
			bson.D{{Key: "user", Value: "Dima"}})
		if err != nil {
			t.Fatal(err)
		}
	}()

	entry := model.Entry{
		Type: "expenses",
		Item: "coffee",
		User: "Dima",
		Date: time.Now().UTC(),
		Sum:  3.6,
	}

	err := financeRepo.Add(ctx, &entry)
	if err != nil {
		t.Fatal(err)
	}

	data, err := financeRepo.Get(ctx, &model.Entry{
		Type: entry.Type,
		User: entry.User,
		Date: entry.Date,
	})
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
		_, err := mongoCli.Database("expenses").Collection(time.Now().UTC().Format(layout)).DeleteMany(ctx,
			bson.D{{Key: "user", Value: "Dima"}})
		if err != nil {
			t.Fatal(err)
		}
	}()

	entry1 := model.Entry{
		Type: "expenses",
		Item: "coffee",
		User: "Dima",
		Date: time.Now().UTC(),
		Sum:  3.5,
	}
	entry2 := model.Entry{
		Type: "expenses",
		Item: "coffee",
		User: "Dima",
		Date: time.Now().UTC(),
		Sum:  2,
	}
	entry3 := model.Entry{
		Type: "expenses",
		Item: "rent",
		User: "Dima",
		Date: time.Now().UTC(),
		Sum:  560,
	}
	for _, entry := range []*model.Entry{
		&entry1, &entry2, &entry3,
	} {
		err := financeRepo.Add(ctx, entry)
		if err != nil {
			t.Fatal(err)
		}
	}

	data, err := financeRepo.Get(ctx, &model.Entry{
		Type: entry1.Type,
		User: entry1.User,
		Date: entry1.Date,
	})
	if err != nil {
		t.Fatal(err)
	}
	logrus.Infof("received entry: %v", data)
	require.Equal(t, 2, len(data))
	require.Equal(t, data["coffee"], entry1.Sum+entry2.Sum)
	require.Equal(t, data["rent"], entry3.Sum)
}

// TestFinance_Replace tests the insertion if the object was not found
func TestFinance_Replace(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	defer func() {
		_, err := mongoCli.Database("expenses").Collection(time.Now().UTC().Format(layout)).DeleteMany(ctx,
			bson.D{{Key: "user", Value: "Dima"}})
		if err != nil {
			t.Fatal(err)
		}
	}()

	entry := model.Entry{
		Type: "expenses",
		Item: "coffee",
		User: "Dima",
		Date: time.Now().UTC(),
		Sum:  3.6,
	}
	err := financeRepo.Replace(ctx, &entry)
	if err != nil {
		t.Fatal(err)
	}

	data, err := financeRepo.Get(ctx, &model.Entry{
		Type: entry.Type,
		User: entry.User,
		Date: entry.Date,
	})
	if err != nil {
		t.Fatal(err)
	}
	logrus.Infof("received entry: %v", data)
	require.Equal(t, 1, len(data))
	require.Equal(t, data["coffee"], 3.6)
}

func TestFinance_Replace2(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	defer func() {
		_, err := mongoCli.Database("expenses").Collection(time.Now().UTC().Format(layout)).DeleteMany(ctx,
			bson.D{{Key: "user", Value: "Dima"}})
		if err != nil {
			t.Fatal(err)
		}
	}()

	entry := model.Entry{
		Type: "expenses",
		Item: "coffee",
		User: "Dima",
		Date: time.Now().UTC(),
		Sum:  3.6,
	}
	err := financeRepo.Replace(ctx, &entry)
	if err != nil {
		t.Fatal(err)
	}

	entry.Item = "rent"
	entry.Sum = 545.65
	err = financeRepo.Replace(ctx, &entry)
	if err != nil {
		t.Fatal(err)
	}

	data, err := financeRepo.Get(ctx, &model.Entry{
		Type: entry.Type,
		User: entry.User,
		Date: entry.Date,
	})
	if err != nil {
		t.Fatal(err)
	}
	logrus.Infof("received entry: %v", data)
	require.Equal(t, 1, len(data))
	require.Equal(t, data["rent"], 545.65)
}

func TestFinance_ClearYesterday(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	defer func() {
		err := mongoCli.Database("today_expenses").Collection(time.Now().UTC().Format(layout)).Drop(ctx)
		if err != nil {
			t.Fatal(err)
		}
	}()

	entry := model.Entry{
		Type: "today_expenses",
		Item: "coffee",
		User: "Dima",
		Date: time.Now().UTC().Add(time.Hour),
		Sum:  3.6,
	}

	err := financeRepo.Add(ctx, &entry)
	if err != nil {
		t.Fatal(err)
	}

	data, err := financeRepo.Get(ctx, &model.Entry{
		Type: entry.Type,
		User: entry.User,
		Date: entry.Date,
	})
	if err != nil {
		t.Fatal(err)
	}
	require.Equal(t, 1, len(data))

	err = financeRepo.ClearYesterday(ctx, &entry)
	if err != nil {
		t.Fatal(err)
	}
	_, err = financeRepo.Get(ctx, &model.Entry{
		Type: entry.Type,
		User: entry.User,
		Date: entry.Date,
	})
	require.Error(t, err)
}
