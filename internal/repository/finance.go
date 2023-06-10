package repository

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/chucky-1/finance/internal/model"
)

type Recorder interface {
	Add(ctx context.Context, entry *model.Entry, period string) error
}

type Getter interface {
	Get(ctx context.Context, entry *model.Entry, period string) (map[string]float64, error)
	GetByUsers(ctx context.Context, users []string, kind, period string) (map[string]map[string]float64, error)
}

type Cleaner interface {
	DeleteByUsers(ctx context.Context, users []string, kind, period string) error
}

type Mongo struct {
	cli *mongo.Client
}

func NewMongo(cli *mongo.Client) *Mongo {
	return &Mongo{
		cli: cli,
	}
}

func (m *Mongo) Add(ctx context.Context, entry *model.Entry, period string) error {
	_, err := m.cli.Database(entry.Kind).Collection(period).UpdateOne(ctx,
		bson.D{{Key: "user", Value: entry.User}},
		bson.D{{Key: "$inc", Value: bson.D{{Key: entry.Item, Value: entry.Sum}}}}, options.Update().SetUpsert(true))
	if err != nil {
		return fmt.Errorf("mongo couldn't UpdateOne in Add method: %v", err)
	}
	return nil
}

func (m *Mongo) Get(ctx context.Context, entry *model.Entry, period string) (map[string]float64, error) {
	result := m.cli.Database(entry.Kind).Collection(period).FindOne(ctx,
		bson.D{{Key: "user", Value: entry.User}})
	if result.Err() != nil {
		return nil, fmt.Errorf("mongo couldn't FindOne in Get method: %v", result.Err())
	}

	var data bson.D
	err := result.Decode(&data)
	if err != nil {
		return nil, fmt.Errorf("mongo couldn't Decode in Get method: %v", err)
	}
	return unmarshal(&data), nil
}

func (m *Mongo) GetByUsers(ctx context.Context, users []string, kind, period string) (map[string]map[string]float64, error) {
	cursor, err := m.cli.Database(kind).Collection(period).Find(ctx,
		bson.D{{Key: "user", Value: bson.D{{Key: "$in", Value: users}}}})
	if err != nil {
		return nil, fmt.Errorf("mongo couldn't Find in GetByUsers method: %v", err)
	}
	defer func(cursor *mongo.Cursor, ctx context.Context) {
		err = cursor.Close(ctx)
		if err != nil {
			logrus.Errorf("mongo couldn't close cursor in GetByUsers method")
		}
	}(cursor, ctx)

	result := make(map[string]map[string]float64)
	for cursor.Next(ctx) {
		var data bson.D
		if err = cursor.Decode(&data); err != nil {
			return nil, fmt.Errorf("mongo couldn't Decode in GetByUsers method: %v", err)
		}
		user, ok := data.Map()["user"]
		if !ok {
			logrus.Infof("GetByUsers method: user didn't find in map: %v", data.Map())
			continue
		}
		result[user.(string)] = unmarshal(&data)
	}
	if err = cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor err in GetByUsers method: %v", err)
	}
	return result, nil
}

func (m *Mongo) DeleteByUsers(ctx context.Context, users []string, kind, period string) error {
	_, err := m.cli.Database(kind).Collection(period).DeleteMany(ctx,
		bson.D{{Key: "user", Value: bson.D{{Key: "$in", Value: users}}}})
	if err != nil {
		return fmt.Errorf("mongo could't DeleteMany in DeleteByUsers method %v", err)
	}
	return nil
}

func unmarshal(data *bson.D) map[string]float64 {
	m := make(map[string]float64)
	for key, value := range data.Map() {
		if key == "_id" || key == "user" {
			continue
		}
		m[key] = value.(float64)
	}
	return m
}
