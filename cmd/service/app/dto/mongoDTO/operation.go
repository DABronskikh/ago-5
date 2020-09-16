package mongoDTO

import (
	"errors"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	ErrRequiredOperationUserID = errors.New("required parameter is missing: User ID")
)

type OperationsDTO struct {
	ID         primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	UserId     int64              `json:"user_id" bson:"user_id"`
	Operations []OperationDTO     `json:"operations" bson:"operations"`
}

type OperationDTO struct {
	Icon  string `json:"icon"`
	Title string `json:"title"`
	Url   string `json:"url"`
}
