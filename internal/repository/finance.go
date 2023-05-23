package repository

import (
	"go.mongodb.org/mongo-driver/mongo"
)

type Finance struct {
	cli *mongo.Client
}

func NewFinance(cli *mongo.Client) *Finance {
	return &Finance{
		cli: cli,
	}
}
