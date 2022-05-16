package Model

import (
	_ "fmt"
	//MongoDB
	//"go.mongodb.org/mongo-driver/bson/primitive"
	//end MongoDB
)

//MongoDB
type ResPassUserModel struct {
	//primitive
	//ID           primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	ID       string `json:"id" bson:"_id,omitempty"`
	Email    string `json:"email" bson:"email,omitempty"`
	Url      string `json:"url" bson:"url,omitempty"`
	Url_full string `json:"url_full" bson:"url_full,omitempty"`
}

//end_MongoDB
