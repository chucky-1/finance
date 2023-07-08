package model

import "time"

// Entry is one record of expenses or income
type Entry struct {
	Kind     string    `bson:"kind"` // expense or income
	User     string    `bson:"user"`
	Date     time.Time `bson:"date"`
	Category *Category
}

type Category struct {
	Name   string
	Amount float64
}
