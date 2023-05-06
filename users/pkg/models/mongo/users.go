package mongo

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
)

type UserModel struct {
	C *mongo.Collection
}

func (m *UserModel) All() ([]model.User, error) {

	// Define variables
	ctx := context.TODO()

}
