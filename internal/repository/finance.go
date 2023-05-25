package repository

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/chucky-1/finance/internal/model"
)

const layout = "2006-01"

type Finance struct {
	cli *mongo.Client
}

func NewFinance(cli *mongo.Client) *Finance {
	return &Finance{
		cli: cli,
	}
}

func (f *Finance) Add(ctx context.Context, entry *model.Entry) error {
	_, err := f.cli.Database(entry.Type).Collection(entry.Date.Format(layout)).UpdateOne(ctx,
		bson.D{{Key: "user", Value: entry.User}},
		bson.D{{Key: "$inc", Value: bson.D{{Key: entry.Item, Value: entry.Sum}}}}, options.Update().SetUpsert(true))
	if err != nil {
		return fmt.Errorf("repository.Finance.Add error: %v", err)
	}
	return nil
}

func (f *Finance) Get(ctx context.Context, entry *model.Entry) (map[string]float64, error) {
	result := f.cli.Database(entry.Type).Collection(entry.Date.Format(layout)).FindOne(ctx,
		bson.D{{Key: "user", Value: entry.User}})
	if result.Err() != nil {
		return nil, fmt.Errorf("repository.Finance.Get FindOne error: %v", result.Err())
	}

	var data bson.D
	err := result.Decode(&data)
	if err != nil {
		return nil, fmt.Errorf("repository.Finance.Get Decode error: %v", err)
	}
	return unmarshal(&data)
}

func unmarshal(data *bson.D) (map[string]float64, error) {
	m := make(map[string]float64)
	for key, value := range data.Map() {
		if key == "_id" || key == "user" {
			continue
		}
		m[key] = value.(float64)
	}
	return m, nil
}
