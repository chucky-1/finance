package repository

import (
	"context"
	"fmt"
	"github.com/chucky-1/finance/internal/model"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"strings"
)

const (
	amount = "Amount"
)

type Recorder interface {
	Add(ctx context.Context, entry *model.Entry, period string) error
}

type Getter interface {
	Get(ctx context.Context, entry *model.Entry, period string) (map[string]float64, error)
	GetByUsernames(ctx context.Context, usernames []string, kind, period string) (map[string]map[string]float64, error)
}

type Cleaner interface {
	DeleteByUsernames(ctx context.Context, users []string, kind, period string) error
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
		bson.D{{Key: "$inc", Value: bson.D{{Key: fmt.Sprintf("%s.%s", entry.Category.Name, amount), Value: entry.Category.Amount}}}},
		options.Update().SetUpsert(true))
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

	categories, err := unmarshal(make(map[string]float64), &data)
	if err != nil {
		return nil, err
	}

	categories = addZeroValuesInEmptyCategory(categories)

	return categories, nil
}

func (m *Mongo) GetByUsernames(ctx context.Context, usernames []string, kind, period string) (map[string]map[string]float64, error) {
	cursor, err := m.cli.Database(kind).Collection(period).Find(ctx,
		bson.D{{Key: "user", Value: bson.D{{Key: "$in", Value: usernames}}}})
	if err != nil {
		return nil, fmt.Errorf("mongo couldn't Find in GetByUsernames method: %v", err)
	}
	defer func(cursor *mongo.Cursor, ctx context.Context) {
		err = cursor.Close(ctx)
		if err != nil {
			logrus.Errorf("mongo couldn't close cursor in GetByUsernames method")
		}
	}(cursor, ctx)

	result := make(map[string]map[string]float64)
	for cursor.Next(ctx) {
		var data bson.D
		if err = cursor.Decode(&data); err != nil {
			return nil, fmt.Errorf("mongo couldn't Decode in GetByUsernames method: %v", err)
		}
		user, ok := data.Map()["user"]
		if !ok {
			logrus.Errorf("GetByUsernames method: user didn't find in map: %v", data.Map())
			continue
		}

		categories, err := unmarshal(make(map[string]float64), &data)
		if err != nil {
			return nil, err
		}

		categories = addZeroValuesInEmptyCategory(categories)

		result[user.(string)] = categories
	}
	if err = cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor err in GetByUsernames method: %v", err)
	}
	return result, nil
}

func (m *Mongo) DeleteByUsernames(ctx context.Context, users []string, kind, period string) error {
	_, err := m.cli.Database(kind).Collection(period).DeleteMany(ctx,
		bson.D{{Key: "user", Value: bson.D{{Key: "$in", Value: users}}}})
	if err != nil {
		return fmt.Errorf("mongo could't DeleteMany in DeleteByUsernames method %v", err)
	}
	return nil
}

func unmarshal(categories map[string]float64, data *bson.D) (map[string]float64, error) {
	for key, object := range data.Map() {
		if key == "_id" || key == "user" {
			continue
		}
		switch object.(type) {
		case primitive.D:
			unmarshalSubCategory(categories, key, object.(primitive.D))
		default:
			return nil, fmt.Errorf("couldn't unmarshal: passed in defauld but it's impossible because all values on first levev should be an Object: %s, %v", key, object)
		}
	}
	return categories, nil
}

func unmarshalSubCategory(categories map[string]float64, parent string, data primitive.D) {
	for key, object := range data.Map() {
		switch object.(type) {
		case primitive.D:
			unmarshalSubCategory(categories, fmt.Sprintf("%s.%s", parent, key), object.(primitive.D))
		default:
			switch key {
			case amount:
				categories[fmt.Sprintf("%s.%s", parent, amount)] = object.(float64)
			}
		}
	}
}

func addZeroValuesInEmptyCategory(categories map[string]float64) map[string]float64 {
	// collect all categories into a slice
	allCategories := make([]string, 0)
	for notEmptyCategory := range categories {
		subs := strings.Split(notEmptyCategory, ".")
		if len(subs) == 2 {
			continue
		}
		var category string
		for i, subCategory := range subs {
			if subCategory == amount {
				continue
			}
			if i == 0 {
				category = subCategory
				allCategories = append(allCategories, fmt.Sprintf("%s.%s", category, amount))
			} else {
				category = fmt.Sprintf("%s.%s", category, subCategory)
				allCategories = append(allCategories, fmt.Sprintf("%s.%s", category, amount))
			}
		}
	}

	for _, category := range allCategories {
		_, ok := categories[category]
		if !ok {
			categories[category] = 0
		}
	}
	return categories
}
