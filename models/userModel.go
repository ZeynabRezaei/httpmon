package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

//User is the model that governs all notes objects retrived or inserted into the DB
type User struct {
	ID            primitive.ObjectID `bson:"_id"`
	Name          *string            `json:"name" validate:"required,min=3""`
	Password      *string            `json:"password" validate:"required,min=6""`
	Token         *string            `json:"token"`
	Refresh_token *string            `json:"refresh_token"`
	Created_at    time.Time          `json:"created_at"`
	Updated_at    time.Time          `json:"updated_at"`
	User_id       string             `json:"user_id"`
	Urls          []URL              `json:"urls"`
}
