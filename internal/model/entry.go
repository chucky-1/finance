package model

import "time"

// Entry is one record of expenses or income
type Entry struct {
	Type string    `bson:"type"` // expense or income
	Item string    `bson:"item"` // for example coffee, rent of apartment
	User string    `bson:"user"`
	Date time.Time `bson:"date"`
	Sum  float64   `bson:"sum"`
}
