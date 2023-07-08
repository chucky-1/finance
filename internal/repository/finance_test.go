package repository

import (
	"context"
	"fmt"
	"github.com/chucky-1/finance/internal/model"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

var (
	period      = "today"
	financeRepo *Mongo
)

func TestMongo_AddGet(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	defer func() {
		err := mongoCli.Database("expenses").Collection(period).Drop(ctx)
		if err != nil {
			t.Fatal(err)
		}
	}()

	kind := "expenses"
	user := "user"
	e := model.Entry{
		Kind: kind,
		User: user,
		Category: &model.Category{
			Name:   "Food",
			Amount: 25.6,
		},
	}
	err := financeRepo.Add(ctx, &e, period)
	if err != nil {
		t.Fatal(err)
	}

	categories, err := financeRepo.Get(ctx, &model.Entry{
		Kind: kind,
		User: user,
	}, period)
	if err != nil {
		t.Fatal(err)
	}
	logrus.Info(categories)
	require.Equal(t, 1, len(categories))
	require.Equal(t, categories["Food.Amount"], 25.6)
}

func TestMongo_AddGetSubCategories(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	defer func() {
		err := mongoCli.Database("expenses").Collection(period).Drop(ctx)
		if err != nil {
			t.Fatal(err)
		}
	}()

	kind := "expenses"
	user := "user"
	e1 := model.Entry{
		Kind: kind,
		User: user,
		Category: &model.Category{
			Name:   "Food",
			Amount: 10,
		},
	}
	e2 := model.Entry{
		Kind: kind,
		User: user,
		Category: &model.Category{
			Name:   "Food.Ih house",
			Amount: 15.6,
		},
	}
	e3 := model.Entry{
		Kind: kind,
		User: user,
		Category: &model.Category{
			Name:   "Food.Outside",
			Amount: 50,
		},
	}
	e4 := model.Entry{
		Kind: kind,
		User: user,
		Category: &model.Category{
			Name:   "Food.Outside.Fast food",
			Amount: 12.2,
		},
	}
	e5 := model.Entry{
		Kind: kind,
		User: user,
		Category: &model.Category{
			Name:   "Food.Outside.Restaurant",
			Amount: 86.6,
		},
	}
	e6 := model.Entry{
		Kind: kind,
		User: user,
		Category: &model.Category{
			Name:   "Relax",
			Amount: 153.70,
		},
	}
	for _, e := range []*model.Entry{
		&e1, &e2, &e3, &e4, &e5, &e6,
	} {
		err := financeRepo.Add(ctx, e, period)
		if err != nil {
			t.Fatal(err)
		}
	}

	categories, err := financeRepo.Get(ctx, &model.Entry{
		Kind: kind,
		User: user,
	}, period)
	if err != nil {
		t.Fatal(err)
	}
	logrus.Info(categories)
	require.Equal(t, 6, len(categories))
	require.Equal(t, categories["Food.Amount"], e1.Category.Amount)
	require.Equal(t, categories["Food.Ih house.Amount"], e2.Category.Amount)
	require.Equal(t, categories["Food.Outside.Amount"], e3.Category.Amount)
	require.Equal(t, categories["Food.Outside.Fast food.Amount"], e4.Category.Amount)
	require.Equal(t, categories["Food.Outside.Restaurant.Amount"], e5.Category.Amount)
	require.Equal(t, categories["Relax.Amount"], e6.Category.Amount)
}

func TestMongo_AddGetViaCategory(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	defer func() {
		err := mongoCli.Database("expenses").Collection(period).Drop(ctx)
		if err != nil {
			t.Fatal(err)
		}
	}()

	kind := "expenses"
	user := "user"
	e := model.Entry{
		Kind: kind,
		User: user,
		Category: &model.Category{
			Name:   "Food.Outside.Fast food",
			Amount: 10,
		},
	}
	err := financeRepo.Add(ctx, &e, period)
	if err != nil {
		t.Fatal(err)
	}

	categories, err := financeRepo.Get(ctx, &model.Entry{
		Kind: kind,
		User: user,
	}, period)
	if err != nil {
		t.Fatal(err)
	}
	require.Equal(t, 3, len(categories))

	require.Equal(t, categories["Food.Outside.Fast food.Amount"], e.Category.Amount)

	sum, ok := categories["Food.Outside.Amount"]
	require.True(t, ok)
	require.Equal(t, float64(0), sum)

	sum, ok = categories["Food.Amount"]
	require.True(t, ok)
	require.Equal(t, float64(0), sum)
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

	e1 := model.Entry{
		Kind: "expenses",
		User: "Dima",
		Date: time.Now().UTC(),
		Category: &model.Category{
			Name:   "coffee",
			Amount: 3.5,
		},
	}
	e2 := model.Entry{
		Kind: "expenses",
		User: "Dima",
		Date: time.Now().UTC(),
		Category: &model.Category{
			Name:   "coffee",
			Amount: 2,
		},
	}
	e3 := model.Entry{
		Kind: "expenses",
		User: "Dima",
		Date: time.Now().UTC(),
		Category: &model.Category{
			Name:   "rent",
			Amount: 560,
		},
	}
	for _, entry := range []*model.Entry{
		&e1, &e2, &e3,
	} {
		err := financeRepo.Add(ctx, entry, period)
		if err != nil {
			t.Fatal(err)
		}
	}

	data, err := financeRepo.Get(ctx, &model.Entry{
		Kind: e1.Kind,
		User: e1.User,
		Date: e1.Date,
	}, period)
	if err != nil {
		t.Fatal(err)
	}
	require.Equal(t, 2, len(data))
	require.Equal(t, data["coffee.Amount"], e1.Category.Amount+e2.Category.Amount)
	require.Equal(t, data["rent.Amount"], e3.Category.Amount)
}

func TestMongo_AddUpdateViaCategory(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	defer func() {
		err := mongoCli.Database("expenses").Collection(period).Drop(ctx)
		if err != nil {
			t.Fatal(err)
		}
	}()

	e1 := model.Entry{
		Kind: "expenses",
		User: "Dima",
		Date: time.Now().UTC(),
		Category: &model.Category{
			Name:   "drink.coffee",
			Amount: 3.5,
		},
	}
	e2 := model.Entry{
		Kind: "expenses",
		User: "Dima",
		Date: time.Now().UTC(),
		Category: &model.Category{
			Name:   "drink.coffee",
			Amount: 4.5,
		},
	}
	for _, entry := range []*model.Entry{
		&e1, &e2,
	} {
		err := financeRepo.Add(ctx, entry, period)
		if err != nil {
			t.Fatal(err)
		}
	}

	data, err := financeRepo.Get(ctx, &model.Entry{
		Kind: e1.Kind,
		User: e1.User,
		Date: e1.Date,
	}, period)
	if err != nil {
		t.Fatal(err)
	}
	require.Equal(t, 2, len(data))
	require.Equal(t, data["drink.coffee.Amount"], e1.Category.Amount+e2.Category.Amount)
	sum, ok := data["drink.Amount"]
	require.True(t, ok)
	require.Equal(t, float64(0), sum)
}

func TestMongo_AddUpdateInDifferentLevel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	defer func() {
		err := mongoCli.Database("expenses").Collection(period).Drop(ctx)
		if err != nil {
			t.Fatal(err)
		}
	}()

	e1 := model.Entry{
		Kind: "expenses",
		User: "Dima",
		Date: time.Now().UTC(),
		Category: &model.Category{
			Name:   "food.drink.coffee",
			Amount: 3.5,
		},
	}
	e2 := model.Entry{
		Kind: "expenses",
		User: "Dima",
		Date: time.Now().UTC(),
		Category: &model.Category{
			Name:   "food.drink",
			Amount: 4.5,
		},
	}
	for _, entry := range []*model.Entry{
		&e1, &e2,
	} {
		err := financeRepo.Add(ctx, entry, period)
		if err != nil {
			t.Fatal(err)
		}
	}

	data, err := financeRepo.Get(ctx, &model.Entry{
		Kind: e1.Kind,
		User: e1.User,
		Date: e1.Date,
	}, period)
	if err != nil {
		t.Fatal(err)
	}
	require.Equal(t, 3, len(data))
	require.Equal(t, data["food.drink.coffee.Amount"], e1.Category.Amount)
	require.Equal(t, data["food.drink.Amount"], e2.Category.Amount)
	sum, ok := data["food.Amount"]
	require.True(t, ok)
	require.Equal(t, float64(0), sum)
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

	e1 := model.Entry{
		Kind: "expenses",
		User: "Dima",
		Date: beginningOfDay.Add(-time.Hour),
		Category: &model.Category{
			Name:   "coffee",
			Amount: 3.6,
		},
	}
	e2 := model.Entry{
		Kind: "expenses",
		User: "Pasha",
		Date: beginningOfDay.Add(10 * -time.Hour),
		Category: &model.Category{
			Name:   "Rent",
			Amount: 560,
		},
	}
	e3 := model.Entry{
		Kind: "expenses",
		User: "Luisa",
		Date: beginningOfDay.Add(5 * -time.Hour),
		Category: &model.Category{
			Name:   "Food",
			Amount: 20.40,
		},
	}
	e4 := model.Entry{
		Kind: "expenses",
		User: "Bad user",
		Date: beginningOfDay.Add(7 * -time.Hour),
		Category: &model.Category{
			Name:   "Drags",
			Amount: 999.9,
		},
	}
	for _, e := range []*model.Entry{
		&e1, &e2, &e3, &e4,
	} {
		err := financeRepo.Add(ctx, e, period)
		if err != nil {
			t.Fatal(err)
		}
	}

	users := []string{
		e1.User, e2.User, e3.User,
	}
	entries, err := financeRepo.GetByUsernames(ctx, users, "expenses", period)
	if err != nil {
		t.Fatal(err)
	}
	require.Equal(t, 3, len(entries))
	require.Equal(t, entries[e1.User][fmt.Sprintf("%s.%s", e1.Category.Name, amount)], e1.Category.Amount)
	require.Equal(t, entries[e2.User][fmt.Sprintf("%s.%s", e2.Category.Name, amount)], e2.Category.Amount)
	require.Equal(t, entries[e3.User][fmt.Sprintf("%s.%s", e3.Category.Name, amount)], e3.Category.Amount)
}

func TestFinance_GetByUsersSubCategories(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	defer func() {
		err := mongoCli.Database("expenses").Collection(period).Drop(ctx)
		if err != nil {
			t.Fatal(err)
		}
	}()

	beginningOfDay := time.Now().UTC().Truncate(24 * time.Hour)

	e1 := model.Entry{
		Kind: "expenses",
		User: "Dima",
		Date: beginningOfDay.Add(-time.Hour),
		Category: &model.Category{
			Name:   "coffee",
			Amount: 3.6,
		},
	}
	e2 := model.Entry{
		Kind: "expenses",
		User: "Dima",
		Date: beginningOfDay.Add(-time.Hour),
		Category: &model.Category{
			Name:   "coffee.cofix",
			Amount: 10,
		},
	}
	e3 := model.Entry{
		Kind: "expenses",
		User: "Pasha",
		Date: beginningOfDay.Add(-10 * time.Hour),
		Category: &model.Category{
			Name:   "Rent",
			Amount: 560,
		},
	}
	e4 := model.Entry{
		Kind: "expenses",
		User: "Pasha",
		Date: beginningOfDay.Add(-10 * time.Hour),
		Category: &model.Category{
			Name:   "Rent.Sub.SubSub",
			Amount: 56,
		},
	}
	e5 := model.Entry{
		Kind: "expenses",
		User: "Luisa",
		Date: beginningOfDay.Add(-5 * time.Hour),
		Category: &model.Category{
			Name:   "Food",
			Amount: 20.40,
		},
	}
	e6 := model.Entry{
		Kind: "expenses",
		User: "Bad user",
		Date: beginningOfDay.Add(-7 * time.Hour),
		Category: &model.Category{
			Name:   "Drags",
			Amount: 999.9,
		},
	}
	for _, e := range []*model.Entry{
		&e1, &e2, &e3, &e4, &e5, &e6,
	} {
		err := financeRepo.Add(ctx, e, period)
		if err != nil {
			t.Fatal(err)
		}
	}

	users, err := financeRepo.GetByUsernames(ctx, []string{
		e1.User, e3.User, e5.User,
	}, "expenses", period)
	if err != nil {
		t.Fatal(err)
	}
	logrus.Info(users)
	require.Equal(t, 3, len(users))
	require.Equal(t, users[e1.User][fmt.Sprintf("%s.%s", e1.Category.Name, amount)], e1.Category.Amount)
	require.Equal(t, users[e2.User][fmt.Sprintf("%s.%s", e2.Category.Name, amount)], e2.Category.Amount)
	require.Equal(t, users[e3.User][fmt.Sprintf("%s.%s", e3.Category.Name, amount)], e3.Category.Amount)
	require.Equal(t, users[e3.User]["Rent.Sub.Amount"], float64(0))
	require.Equal(t, users[e4.User][fmt.Sprintf("%s.%s", e4.Category.Name, amount)], e4.Category.Amount)
	require.Equal(t, users[e5.User][fmt.Sprintf("%s.%s", e5.Category.Name, amount)], e5.Category.Amount)
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

	e1 := model.Entry{
		Kind: "expenses",
		User: "Dima",
		Date: beginningOfDay.Add(-time.Hour),
		Category: &model.Category{
			Name:   "coffee",
			Amount: 3.6,
		},
	}
	e2 := model.Entry{
		Kind: "expenses",
		User: "Pasha",
		Date: beginningOfDay.Add(10 * -time.Hour),
		Category: &model.Category{
			Name:   "Rent",
			Amount: 560,
		},
	}
	e3 := model.Entry{
		Kind: "expenses",
		User: "Luisa",
		Date: beginningOfDay.Add(5 * -time.Hour),
		Category: &model.Category{
			Name:   "Food",
			Amount: 20.40,
		},
	}
	e4 := model.Entry{
		Kind: "expenses",
		User: "Luisa",
		Date: beginningOfDay.Add(7 * -time.Hour),
		Category: &model.Category{
			Name:   "Food.In house",
			Amount: 50,
		},
	}
	e5 := model.Entry{
		Kind: "expenses",
		User: "Bad user",
		Date: beginningOfDay.Add(7 * -time.Hour),
		Category: &model.Category{
			Name:   "Drags",
			Amount: 999.9,
		},
	}

	for _, e := range []*model.Entry{
		&e1, &e2, &e3, &e4, &e5,
	} {
		err := financeRepo.Add(ctx, e, period)
		if err != nil {
			t.Fatal(err)
		}
	}

	users := []string{
		e1.User, e2.User, e3.User, e5.User,
	}

	entries, err := financeRepo.GetByUsernames(ctx, users, "expenses", period)
	if err != nil {
		t.Fatal(err)
	}
	require.Equal(t, 4, len(entries))

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
