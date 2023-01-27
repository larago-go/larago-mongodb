package Model

import (
	_ "fmt"
	//MongoDB
	//"go.mongodb.org/mongo-driver/bson/primitive"
	//end MongoDB
)

//MongoDB
type UserModel struct {
	//primitive
	//ID           primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	ID       string `json:"id" bson:"_id,omitempty"`
	Name     string `json:"name" bson:"name,omitempty"`
	Email    string `json:"email" bson:"email,omitempty"`
	Password string `json:"password" bson:"password,omitempty"`
	//Casbinrole
	Role string `json:"role" bson:"role,omitempty"`
}

//end MongoDB
