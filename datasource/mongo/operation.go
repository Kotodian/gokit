package mongo

import "context"

func CreateCollection(ctx context.Context, name string) error {
	return db.CreateCollection(ctx, name)
}

func Insert(ctx context.Context, collection string, data interface{}) error {
	_, err := db.Collection(collection).InsertOne(ctx, data)
	return err
}
