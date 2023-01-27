package Model

import (
	_ "fmt"
	//end MongoDB
)

//MongoDB
type CasbinRoleModel struct {
	//primitive
	//ID primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	ID       string `json:"id" bson:"_id,omitempty"`
	RoleName string `json:"v0" bson:"v0,omitempty"`
	Path     string `json:"v1" bson:"v1,omitempty"`
	Method   string `json:"v2" bson:"v2,omitempty"`
}

//end_MongoDB
