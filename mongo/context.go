package mongo

import (
	"github.com/ioswarm/golik"
	"go.mongodb.org/mongo-driver/mongo"
)

type MongoContext struct {
	CloveRef golik.CloveRef
	Client *mongo.Client
	Settings *Settings
}

