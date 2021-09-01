package mongo

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"os"
	"time"
)

var (
	authDatabase string
	database     string
	username     string
	password     string
	host         string
	port         string
)

func InitEnv() {
	authDatabase = os.Getenv("MONGO_AUTH_DB")
	database = os.Getenv("MONGO_DB")
	username = os.Getenv("MONGO_USER")
	password = os.Getenv("MONGO_PASSWD")
	host = os.Getenv("MONGO_HOST")
	port = os.Getenv("MONGO_PORT")
}

func connect() (*mongo.Database, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	o := options.Client().SetAuth(options.Credential{
		Username:   username,
		Password:   password,
		AuthSource: authDatabase,
	}).ApplyURI("mongodb://" + host + ":" + port)
	o.SetMaxPoolSize(10)
	client, err := mongo.Connect(ctx, o)
	if err != nil {
		return nil, err
	}

	return client.Database(database), nil
}
